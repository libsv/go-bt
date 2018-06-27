package cryptolib

import "encoding/hex"

// BuildMerkleRootFromCoinbase comment
func BuildMerkleRootFromCoinbase(coinbaseHash []byte, merkleBranches []string) []byte {
	acc := coinbaseHash

	for i := 0; i < len(merkleBranches); i++ {
		branch, _ := hex.DecodeString(merkleBranches[i])
		concat := append(acc, branch...)
		hash := Sha256d(concat)
		acc = hash[:]
	}
	return acc
}
