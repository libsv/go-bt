package address

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcutil/base58"
	"github.com/libsv/libsv/crypto"
	"github.com/libsv/libsv/utils"
)

// An Address struct contains the address string as well as the hash160 hexstring of the public key.
// The address string will be human readable and specific to the network type, but the public key hash
// is useful because it stays the same regardless of the network type (mainnet, testnet).
type Address struct {
	AddressString string
	PublicKeyHash string
}

// NewAddressFromString takes a string address (P2PKH) and returns a pointer to an Address
// which contains the address string as well as the public key hash string.
func NewAddressFromString(addr string) (*Address, error) {
	pkh, err := addressToPubKeyHashStr(addr)
	if err != nil {
		return nil, err
	}
	a := Address{
		AddressString: addr,
		PublicKeyHash: pkh,
	}
	return &a, nil
}

func addressToPubKeyHashStr(address string) (publicKeyHash string, err error) {
	decoded, err := DecodeString(address)

	if err != nil {
		return "", err
	}

	if len(decoded) != 25 {
		return "", fmt.Errorf("invalid address length for '%s'", address)
	}

	switch decoded[0] {
	case 0x00: // Pubkey hash (P2PKH address)
		return hex.EncodeToString(decoded[1 : len(decoded)-4]), nil

	case 0x6f: // Testnet pubkey hash (P2PKH address)
		return hex.EncodeToString(decoded[1 : len(decoded)-4]), nil

	case 0x05: // Script hash (P2SH address)
		fallthrough
	case 0xc4: // Testnet script hash (P2SH address)
		fallthrough

	default:
		return "", fmt.Errorf("Address %s is not supported", address)
	}
}

// NewAddressFromPublicKeyString takes a public key string and returns an Address struct pointer.
// If mainnet parameter is true it will return a mainnet address (starting with a 1).
// Otherwise (mainnet is false) it will return a testnet address (starting with an m or n).
func NewAddressFromPublicKeyString(pubKey string, mainnet bool) (*Address, error) {
	pubKeyBytes, err := hex.DecodeString(pubKey)
	if err != nil {
		return nil, err
	}
	return NewAddressFromPublicKeyHash(crypto.Hash160(pubKeyBytes), mainnet)
}

// NewAddressFromPublicKeyHash takes a public key hash in bytes and returns an Address struct pointer.
// If mainnet parameter is true it will return a mainnet address (starting with a 1).
// Otherwise (mainnet is false) it will return a testnet address (starting with an m or n).
func NewAddressFromPublicKeyHash(hash []byte, mainnet bool) (*Address, error) {

	// regtest := 111
	// mainnet: 0

	bb := make([]byte, 1)
	if mainnet == false {
		bb[0] = 111
	}

	bb = append(bb, hash...)
	addr := base58Encode(bb)

	a := Address{
		AddressString: addr,
		PublicKeyHash: hex.EncodeToString(hash),
	}
	return &a, nil
}

// NewAddressFromPublicKey takes a btcec public key and returns an Address struct pointer.
// If mainnet parameter is true it will return a mainnet address (starting with a 1).
// Otherwise (mainnet is false) it will return a testnet address (starting with an m or n).
func NewAddressFromPublicKey(pubKey *btcec.PublicKey, mainnet bool) (*Address, error) {
	hash := crypto.Hash160(pubKey.SerializeCompressed())

	// regtest := 111
	// mainnet: 0

	bb := make([]byte, 1)
	if mainnet == false {
		bb[0] = 111
	}

	bb = append(bb, hash...)
	addr := base58Encode(bb)

	a := Address{
		AddressString: addr,
		PublicKeyHash: hex.EncodeToString(hash),
	}
	return &a, nil
}

func base58Encode(input []byte) string {
	b := make([]byte, 0, len(input)+4)
	b = append(b, input[:]...)
	cksum := checksum(b)
	b = append(b, cksum[:]...)
	return base58.Encode(b)
}

func checksum(input []byte) (cksum [4]byte) {
	h := crypto.Sha256d(input)
	copy(cksum[:], h[:4])
	return
}

// ValidateAddress checks if an address string is a valid Bitcoin address (ex. P2PKH, BIP276).
// Checks both mainnet and testnet.
func ValidateAddress(address string) (bool, error) {
	if strings.HasPrefix(address, "bitcoin-script:") {
		_, _, _, _, err := utils.DecodeBIP276(address)

		if err != nil {
			return false, fmt.Errorf("bitcoin-script invalid [%+v]", err)
		}
		return true, nil
	}

	return validA58([]byte(address))
}

func validA58(a58 []byte) (bool, error) {
	var a a25
	if err := a.set58(a58); err != nil {
		return false, err
	}
	if a[0] != 0 && a[0] != 0x6f {
		return false, errors.New("not version 0 or 6f")
	}

	checksumOK := a.embeddedChecksum() == a.computeChecksum()

	if !checksumOK {
		return false, errors.New("checksum failed")
	}

	return true, nil
}
