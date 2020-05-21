package block_test

import (
	"testing"

	"github.com/libsv/libsv/block"
)

func Test2(t *testing.T) {
	// merkleRoot, _ := hex.DecodeString("121ce6bd99e60db0bd00dfbfb84f44e66db57f865959cf3db673654bc921bf13")

	var branches = []string{
		"e967140c9072e7c989a4302af55a8df97abf3ab690609e3c41ca199c9dd25fb8"}

	s, _ := block.MerkleRootFromBranches("b5c72338d19308bae4a978ad4e655b15f1fda722f3db056f78bbcfee014ed93d", 0, branches)

	t.Log(s)

}
