package address

import (
	"encoding/hex"
	"fmt"

	"github.com/bitcoinsv/bsvutil/base58"
	"github.com/libsv/libsv/bsvsuite/bsvec"
	"github.com/libsv/libsv/crypto"
)

// An Address struct contains the address string as well as the hash160 hexstring of the public key.
// The address string will be human readable and specific to the network type, but the public key hash
// is useful because it stays the same regardless of the network type (mainnet, testnet).
type Address struct {
	AddressString string
	PublicKeyHash string
}

// NewFromString takes a string address (P2PKH) and returns a pointer to an Address
// which contains the address string as well as the public key hash string.
func NewFromString(addr string) (*Address, error) {
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

func addressToPubKeyHashStr(address string) (string, error) {
	decoded := base58.Decode(address)

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

// NewFromPublicKeyString takes a public key string and returns an Address struct pointer.
// If mainnet parameter is true it will return a mainnet address (starting with a 1).
// Otherwise (mainnet is false) it will return a testnet address (starting with an m or n).
func NewFromPublicKeyString(pubKey string, mainnet bool) (*Address, error) {
	pubKeyBytes, err := hex.DecodeString(pubKey)
	if err != nil {
		return nil, err
	}
	return NewFromPublicKeyHash(crypto.Hash160(pubKeyBytes), mainnet)
}

// NewFromPublicKeyHash takes a public key hash in bytes and returns an Address struct pointer.
// If mainnet parameter is true it will return a mainnet address (starting with a 1).
// Otherwise (mainnet is false) it will return a testnet address (starting with an m or n).
func NewFromPublicKeyHash(hash []byte, mainnet bool) (*Address, error) {

	// regtest := 111
	// mainnet: 0

	bb := make([]byte, 1)
	if mainnet == false {
		bb[0] = 111
	}

	bb = append(bb, hash...)
	addr := Base58EncodeMissingChecksum(bb)

	a := Address{
		AddressString: addr,
		PublicKeyHash: hex.EncodeToString(hash),
	}
	return &a, nil
}

// NewFromPublicKey takes a bsvec public key and returns an Address struct pointer.
// If mainnet parameter is true it will return a mainnet address (starting with a 1).
// Otherwise (mainnet is false) it will return a testnet address (starting with an m or n).
func NewFromPublicKey(pubKey *bsvec.PublicKey, mainnet bool) (*Address, error) {
	hash := crypto.Hash160(pubKey.SerializeCompressed())

	// regtest := 111
	// mainnet: 0

	bb := make([]byte, 1)
	if mainnet == false {
		bb[0] = 111
	}

	bb = append(bb, hash...)
	addr := Base58EncodeMissingChecksum(bb)

	a := Address{
		AddressString: addr,
		PublicKeyHash: hex.EncodeToString(hash),
	}
	return &a, nil
}

// Base58EncodeMissingChecksum appends a checksum to a byte sequence
// then encodes into base58 encoding.
func Base58EncodeMissingChecksum(input []byte) string {
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
