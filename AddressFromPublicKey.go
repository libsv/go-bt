package cryptolib

import (
	"fmt"

	"github.com/btcsuite/btcd/btcec"
)

// AddressFromPublicKeyHash comment
func AddressFromPublicKeyHash(pubKey *btcec.PublicKey, mainnet bool) string {
	fmt.Println("AddressFromPublicKeyHash() is deprecated - please use NewAddressFromPublicKeyHash() instead")
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
