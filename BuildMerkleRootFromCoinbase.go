package cryptolib

import "encoding/hex"

// BuildMerkleRootFromCoinbase builds the merkle root of the block from the coinbase transaction hash (txid)
// and the merkle branches needed to work up the merkle tree and returns the merkle root byte array.
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
