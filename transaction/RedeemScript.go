package transaction

import (
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"log"

	"bitbucket.org/simon_ordish/cryptolib"
	"github.com/btcsuite/btcutil/base58"

	"golang.org/x/crypto/ripemd160"
)

const (
	opZERO          = 0x00
	opBASE          = 0x50
	opHASH160       = 0xa9
	opCHECKMULTISIG = 0xae
	opEQUAL         = 0x87
)

// RedeemScript type
type RedeemScript struct {
	SignaturesRequired int
	PublicKeys         [][]byte
	Signatures         [][]byte
}

// NewRedeemScript comment
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

// NewRedeemScriptFromElectrum comment
func NewRedeemScriptFromElectrum(script string) (*RedeemScript, error) {
	parts, err := cryptolib.DecodeStringParts(script)
	if err != nil {
		return nil, err
	}

	if len(parts) == 0 {
		return nil, errors.New("There should be 5 parts in this redeemScript")
	}

	signaturesRequired := int(parts[0][0]) - opBASE
	if signaturesRequired < 2 {
		return nil, errors.New("Must have 2 or more required signatures for multisig")
	}

	if signaturesRequired > 15 {
		return nil, errors.New("More than 15 signatures is not supported")
	}

	signatureCount := int(parts[len(parts)-2][0]) - opBASE

	if parts[len(parts)-1][0] != opCHECKMULTISIG {
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
		xpub := base58Encode(pubkey[0:78])

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
		publicKey, err := cryptolib.NewPublicKey(xpub)
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

func base58Encode(input []byte) string {
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

func hash160(data []byte) []byte {
	sha := sha256.New()
	ripe := ripemd160.New()
	sha.Write(data)
	ripe.Write(sha.Sum(nil))
	return ripe.Sum(nil)
}

// AddPublicKey comment
func (rs *RedeemScript) AddPublicKey(pkey string, derivationPath []uint32) error {

	if len(derivationPath) != 2 {
		return errors.New("We only support derivation paths with exactly 2 levels")
	}

	pk, err := cryptolib.NewPublicKey(pkey)
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

func (rs *RedeemScript) getAddress() string {
	script := rs.getRedeemScript()
	hash := hash160(script)
	// hash = append([]byte{0x05}, hash...)
	return base58.CheckEncode(hash, 0x05)
}

func (rs *RedeemScript) getPublicKeys() [][]byte {
	return rs.PublicKeys
}

func (rs *RedeemScript) getRedeemScript() []byte {
	var b []byte

	b = append(b, byte(opBASE+rs.SignaturesRequired))
	rs.PublicKeys = cryptolib.SortByteArrays(rs.PublicKeys)
	for _, pk := range rs.PublicKeys {
		b = append(b, byte(len(pk)))
		b = append(b, pk...)
	}

	b = append(b, byte(opBASE+len(rs.PublicKeys)))

	b = append(b, opCHECKMULTISIG)

	return b
}

func (rs *RedeemScript) getRedeemScriptHash() []byte {
	return hash160(rs.getRedeemScript())
}

func (rs *RedeemScript) getScriptPubKey() []byte {
	var b []byte

	h := rs.getRedeemScriptHash()

	b = append(b, opHASH160)
	b = append(b, byte(len(h)))
	b = append(b, h...)
	b = append(b, opEQUAL)

	return b
}
