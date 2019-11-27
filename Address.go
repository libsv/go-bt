package cryptolib

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/btcsuite/btcutil/base58"
)

// Address comment
type Address struct {
	AddressString string
	PublicKeyHash string
}

// NewAddressFromString comment
func NewAddressFromString(addr string) (*Address, error) {
	pkh, err := pubKeyHash(addr)
	if err != nil {
		return nil, err
	}
	a := Address{
		AddressString: addr,
		PublicKeyHash: pkh,
	}
	return &a, nil
}

// NewAddressFromPublicKey comment
func NewAddressFromPublicKey(pubKey string, mainnet bool) (*Address, error) {
	return NewAddressFromPublicKeyHash(Hash160([]byte(pubKey)), mainnet)
}

// NewAddressFromPublicKeyHash comment
func NewAddressFromPublicKeyHash(hash []byte, mainnet bool) (*Address, error) {
	a := Address{
		AddressString: addressFromPublicKeyHash(hash, mainnet),
		PublicKeyHash: hex.EncodeToString(hash),
	}
	return &a, nil
}

// pubKeyHash comment
func pubKeyHash(address string) (publicKeyHash string, err error) {
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

// AddressFromPublicKeyHash comment
func addressFromPublicKeyHash(hash []byte, mainnet bool) string {

	// regtest := 111
	// mainnet: 0

	bb := make([]byte, 1)
	if mainnet == false {
		bb[0] = 111
	}

	bb = append(bb, hash...)
	return base58Encode(bb)
}
