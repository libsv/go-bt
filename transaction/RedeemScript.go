package transaction

import (
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"github.com/jadwahab/libsv/crypto"
	hash2 "github.com/jadwahab/libsv/crypto/hash"
	"github.com/jadwahab/libsv/utils"
	"log"

	"github.com/btcsuite/btcutil/base58"
)

// RedeemScript contains the metadata used when creating an unlocking script (SigScript) for a multisig output.
type RedeemScript struct {
	SignaturesRequired int
	PublicKeys         [][]byte
	Signatures         [][]byte
}

// NewRedeemScript creates a new RedeemScript with minimum signature threshold needed.
func NewRedeemScript(signaturesRequired int) (*RedeemScript, error) {
	if signaturesRequired < 2 {
		return nil, errors.New("Must have 2 or more required signatures for multisig")
	}

	if signaturesRequired > 16 {
		return nil, errors.New("More than 16 signatures is not supported")
	}

	rs := &RedeemScript{
		SignaturesRequired: signaturesRequired,
	}

	return rs, nil
}

// NewRedeemScriptFromElectrum  TODO:
func NewRedeemScriptFromElectrum(script string) (*RedeemScript, error) {
	parts, err := utils.DecodeStringParts(script)
	if err != nil {
		return nil, err
	}

	if len(parts) == 0 {
		return nil, errors.New("There should be 5 parts in this redeemScript")
	}

	signaturesRequired := int(parts[0][0]) - utils.OpBASE
	if signaturesRequired < 2 {
		return nil, errors.New("Must have 2 or more required signatures for multisig")
	}

	if signaturesRequired > 15 {
		return nil, errors.New("More than 15 signatures is not supported")
	}

	signatureCount := int(parts[len(parts)-2][0]) - utils.OpBASE

	if parts[len(parts)-1][0] != utils.OpCHECKMULTISIG {
		return nil, errors.New("Script must end with OP_CHECKMULTISIG")
	}

	rs := &RedeemScript{
		SignaturesRequired: signaturesRequired,
	}

	for _, pubkey := range parts[1 : len(parts)-2] {
		if pubkey[0] != 0xff {
			return nil, errors.New("All Electrum pubkeys should start with 0xff")
		}

		pubkey = pubkey[1:]
		xpub := Base58Encode(pubkey[0:78])

		derivationPath := pubkey[78:]
		var s []uint16
		for len(derivationPath) > 0 {
			n := binary.LittleEndian.Uint16(derivationPath[0:2])
			s = append(s, n)
			derivationPath = derivationPath[2:]
		}

		if len(s) != 2 {
			return nil, errors.New("Derivation path should have exactly 2 items")
		}

		// rs.AddPublicKey("", s)
		publicKey, err := crypto.NewPublicKey(xpub)
		if err != nil {
			return nil, err
		}

		p, err := publicKey.Child(uint32(s[0]))
		if err != nil {
			return nil, err
		}
		p, err = p.Child(uint32(s[1]))
		if err != nil {
			return nil, err
		}

		rs.PublicKeys = append(rs.PublicKeys, p.PublicKey)
	}

	if len(rs.PublicKeys) != signatureCount {
		return nil, errors.New("Number of public keys must equal the required")
	}

	return rs, nil
}

func Base58Encode(input []byte) string {
	b := make([]byte, 0, len(input)+4)
	b = append(b, input[:]...)
	cksum := checksum(b)
	b = append(b, cksum[:]...)
	return base58.Encode(b)
}

func checksum(input []byte) (cksum [4]byte) {
	h := sha256.Sum256(input)
	h2 := sha256.Sum256(h[:])
	copy(cksum[:], h2[:4])
	return
}

// AddPublicKey appends a public key to the RedeemScript.
func (rs *RedeemScript) AddPublicKey(pkey string, derivationPath []uint32) error {

	if len(derivationPath) != 2 {
		return errors.New("We only support derivation paths with exactly 2 levels")
	}

	pk, err := crypto.NewPublicKey(pkey)
	if err != nil {
		return err
	}

	result, err := pk.Child(derivationPath[0])
	if err != nil {
		return err
	}

	result2, err := result.Child(derivationPath[1])
	if err != nil {
		return err
	}

	log.Println(result2.PublicKeyStr)

	rs.PublicKeys = append(rs.PublicKeys, result2.PublicKey)

	return nil
}

func (rs *RedeemScript) GetAddress() string {
	script := rs.GetRedeemScript()
	hash := hash2.Hash160(script)
	// hash = append([]byte{0x05}, hash...)
	return base58.CheckEncode(hash, 0x05)
}

func (rs *RedeemScript) getPublicKeys() [][]byte {
	return rs.PublicKeys
}

func (rs *RedeemScript) GetRedeemScript() []byte {
	var b []byte

	b = append(b, byte(utils.OpBASE+rs.SignaturesRequired))
	rs.PublicKeys = utils.SortByteArrays(rs.PublicKeys)
	for _, pk := range rs.PublicKeys {
		b = append(b, byte(len(pk)))
		b = append(b, pk...)
	}

	b = append(b, byte(utils.OpBASE+len(rs.PublicKeys)))

	b = append(b, utils.OpCHECKMULTISIG)

	return b
}

func (rs *RedeemScript) GetRedeemScriptHash() []byte {
	return hash2.Hash160(rs.GetRedeemScript())
}

func (rs *RedeemScript) getScriptPubKey() []byte {
	var b []byte

	h := rs.GetRedeemScriptHash()

	b = append(b, utils.OpHASH160)
	b = append(b, byte(len(h)))
	b = append(b, h...)
	b = append(b, utils.OpEQUAL)

	return b
}
