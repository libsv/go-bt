package cryptolib

import (
	"crypto/sha256"

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

// AddressFromPublicKeyHash comment
func AddressFromPublicKeyHash(pubKey *btcec.PublicKey, mainnet bool) string {
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
