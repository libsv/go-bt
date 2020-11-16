package crypto

import (
	"crypto/sha256"

	"golang.org/x/crypto/ripemd160"
)

// Sha256 calculates hash(b) and returns the resulting bytes.
func Sha256(b []byte) []byte {
	data := sha256.Sum256(b)
	return data[:]
}

// Sha256d calculates hash(hash(b)) and returns the resulting bytes.
func Sha256d(b []byte) []byte {
	first := Sha256(b)
	return Sha256(first[:])
}

// Ripemd160 hashes with RIPEMD160
func Ripemd160(b []byte) []byte {
	ripe := ripemd160.New()
	_, _ = ripe.Write(b[:])
	return ripe.Sum(nil)
}

// Hash160 hashes with SHA256 and then hashes again with RIPEMD160.
func Hash160(b []byte) []byte {
	hash := Sha256(b)
	return Ripemd160(hash[:])
}
