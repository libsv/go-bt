package cryptolib

import (
	"crypto/sha256"

	"golang.org/x/crypto/ripemd160"
)

// Hash160 comment
func Hash160(data []byte) []byte {
	ripe := ripemd160.New()
	h := sha256.Sum256(data)
	ripe.Write(h[:])
	return ripe.Sum(nil)
}
