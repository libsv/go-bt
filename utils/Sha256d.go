package utils

import "crypto/sha256"

// Sha256d calculates hash(hash(b)) and returns the resulting bytes.
func Sha256d(b []byte) []byte {
	first := sha256.Sum256(b)
	second := sha256.Sum256(first[:])
	return second[:]
}
