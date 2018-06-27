package cryptolib

import (
	"encoding/hex"
)

func getHashes(txHashes []string) []string {
	var hashes []string

	for _, tx := range txHashes {
		hashes = append(hashes, ReverseHexString(tx))
	}

	return hashes
}

// GetMerkleBranches comment
func GetMerkleBranches(template []string) []string {
	hashes := getHashes(template)

	var branches []string
	var walkBranch func(hashes []string) []string

	walkBranch = func(hashes []string) []string {
		var results []string

		tot := len(hashes)

		if len(hashes) < 2 {
			return make([]string, 0)
		}

		branches = append(branches, hashes[1])

		for i := 0; i < tot; i += 2 {
			var a, _ = hex.DecodeString(hashes[i])
			var b []byte
			if (i + 1) < tot {
				b, _ = hex.DecodeString(hashes[i+1])
			} else {
				b = a
			}

			concat := append(a, b...)
			hash := Sha256d(concat)
			results = append(results, hex.EncodeToString(hash[:]))
		}

		return walkBranch(results)
	}

	walkBranch(hashes)

	return branches
}
