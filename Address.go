package cryptolib

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcutil/base58"
)

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

// AddressFromPublicKey takes a btcec public key and returns a P2PKH address string.
// If mainnet parameter is true it will return a mainnet address (starting with a 1).
// Otherwise (mainnet is false) it will return a testnet address (starting with an m or n).
func AddressFromPublicKey(pubKey *btcec.PublicKey, mainnet bool) string {
	hash := Hash160(pubKey.SerializeCompressed())

	// regtest := 111
	// mainnet: 0

	bb := make([]byte, 1)
	if mainnet == false {
		bb[0] = 111
	}

	bb = append(bb, hash...)
	return base58Encode(bb)
}

// AddressFromPublicKeyHash takes a byte array hash of a public key and returns a P2PKH address string.
// If mainnet parameter is true it will return a mainnet address (starting with a 1).
// Otherwise (mainnet is false) it will return a testnet address (starting with an m or n).
func AddressFromPublicKeyHash(pubKeyHash []byte, mainnet bool) string {
	// regtest := 111
	// mainnet: 0

	bb := make([]byte, 1)
	if mainnet == false {
		bb[0] = 111
	}

	bb = append(bb, pubKeyHash...)
	return base58Encode(bb)
}

// PublicKeyHashFromPublicKeyStr hashes a public key string (in compressed format starting with 03 or 02)
// and returns the hash encoded as a string of hex values.
func PublicKeyHashFromPublicKeyStr(pubKeyStr string) (string, error) {
	pubKeyBytes, err := hex.DecodeString(pubKeyStr)
	if err != nil {
		return "", err
	}

	pubKey, err := btcec.ParsePubKey(pubKeyBytes, btcec.S256())
	if err != nil {
		return "", err
	}

	return PublicKeyHashFromPublicKey(pubKey), nil
}

// PublicKeyHashFromPublicKey hashes a btcec public key (in compressed format starting with 03 or 02)
// and returns the hash encoded as a string of hex values.
func PublicKeyHashFromPublicKey(pubKey *btcec.PublicKey) string {
	hash := Hash160(pubKey.SerializeCompressed())

	return hex.EncodeToString(hash)
}

// AddressToPubKeyHash decodes a Bitcoin address (P2PKH) into the hash of the public key
// encoded as a string of hex values.
func AddressToPubKeyHash(address string) (publicKeyHash string, err error) {
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

// An Address contains the address string as well as the public key hash string.
type Address struct {
	AddressString string
	PublicKeyHash string
}

// NewAddressFromString takes a string address (P2PKH) and returns a pointer to an Address
// which contains the address string as well as the public key hash string.
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

// NewAddressFromPublicKey takes a public key string and returns an Address struct pointer.
// If mainnet parameter is true it will return a mainnet address (starting with a 1).
// Otherwise (mainnet is false) it will return a testnet address (starting with an m or n).
func NewAddressFromPublicKey(pubKey string, mainnet bool) (*Address, error) {
	return NewAddressFromPublicKeyHash(Hash160([]byte(pubKey)), mainnet)
}

// NewAddressFromPublicKeyHash takes a public key hash in bytes and returns an Address struct pointer.
// If mainnet parameter is true it will return a mainnet address (starting with a 1).
// Otherwise (mainnet is false) it will return a testnet address (starting with an m or n).
func NewAddressFromPublicKeyHash(hash []byte, mainnet bool) (*Address, error) {
	a := Address{
		AddressString: addressFromPublicKeyHash(hash, mainnet),
		PublicKeyHash: hex.EncodeToString(hash),
	}
	return &a, nil
}

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
