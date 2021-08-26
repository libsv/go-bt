package bscript

import (
	"encoding/hex"
	"fmt"

	"github.com/libsv/go-bk/base58"
	"github.com/libsv/go-bk/bec"
	"github.com/libsv/go-bk/crypto"
)

const (
	hashP2PKH        = 0x00
	hashTestNetP2PKH = 0x6f
	hashP2SH         = 0x05 // TODO: remove deprecated p2sh stuff
	hashTestNetP2SH  = 0xc4
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
	return &Address{
		AddressString: addr,
		PublicKeyHash: pkh,
	}, nil
}

func addressToPubKeyHashStr(address string) (string, error) {
	decoded := base58.Decode(address)

	if len(decoded) != 25 {
		return "", fmt.Errorf("invalid address length for '%s'", address)
	}

	switch decoded[0] {
	case hashP2PKH: // Pubkey hash (P2PKH address)
		return hex.EncodeToString(decoded[1 : len(decoded)-4]), nil

	case hashTestNetP2PKH: // Testnet pubkey hash (P2PKH address)
		return hex.EncodeToString(decoded[1 : len(decoded)-4]), nil

	case hashP2SH: // Script hash (P2SH address)
		fallthrough
	case hashTestNetP2SH: // Testnet script hash (P2SH address)
		fallthrough
	default:
		return "", fmt.Errorf("address %s is not supported", address)
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
	if !mainnet {
		bb[0] = 111
	}
	// nolint: makezero // we need to setup the array with 1
	bb = append(bb, hash...)

	return &Address{
		AddressString: Base58EncodeMissingChecksum(bb),
		PublicKeyHash: hex.EncodeToString(hash),
	}, nil
}

// NewAddressFromPublicKey takes a bec public key and returns an Address struct pointer.
// If mainnet parameter is true it will return a mainnet address (starting with a 1).
// Otherwise (mainnet is false) it will return a testnet address (starting with an m or n).
func NewAddressFromPublicKey(pubKey *bec.PublicKey, mainnet bool) (*Address, error) {
	hash := crypto.Hash160(pubKey.SerialiseCompressed())

	// regtest := 111
	// mainnet: 0

	bb := make([]byte, 1)
	if !mainnet {
		bb[0] = 111
	}
	// nolint: makezero // we need to setup the array with 1
	bb = append(bb, hash...)

	return &Address{
		AddressString: Base58EncodeMissingChecksum(bb),
		PublicKeyHash: hex.EncodeToString(hash),
	}, nil
}

// Base58EncodeMissingChecksum appends a checksum to a byte sequence
// then encodes into base58 encoding.
func Base58EncodeMissingChecksum(input []byte) string {
	b := make([]byte, 0, len(input)+4)
	b = append(b, input[:]...)
	ckSum := checksum(b)
	b = append(b, ckSum[:]...)
	return base58.Encode(b)
}

func checksum(input []byte) (ckSum [4]byte) {
	h := crypto.Sha256d(input)
	copy(ckSum[:], h[:4])
	return
}
