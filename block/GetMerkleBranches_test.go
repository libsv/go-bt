package block_test

import (
	"testing"

	"github.com/libsv/libsv/block"
)

func TestMerkleRootFromBranches(t *testing.T) {
	branches := []string{"a99d3ab161f6056edb8fb86191979bc1281476cdc85dfe44b3049dda1afea1d2", "01c81e306c70fb0c44b565a709a33fb9ba175aeec3b666af0b3dc1f100dcb557", "f50cd6a879f9f58d6e87047b4bf0502d0bc072c369fd6ea84516a3fc2256a863", "57c67cbf85be69abe75b999bbb21596b50bf9d489f9d60ee4d6eee1d8207a9d5", "eb9883488e5e59dbce82583f4ee7e3deca61f2d82e5bdef1ff7d877a263a2b2e", "34162fa4f9afcc3312a4d37ab78f8f66b3cb9368a0c14b4ab889eb4de7f7077c"}
	index := 18
	hash := "a2d8d44f302381d90a53078e8d80058e372f6adb59058c53aca0f66636578422"

	root, err := block.MerkleRootFromBranches(hash, index, branches)
	if err != nil {
		t.Error(err)
		return
	}

	expected := "1504316d94e3233e1307253f157a1af5f3e90c2fb9c07049d142ea3494d22194"
	if root != expected {
		t.Errorf("Expected %q, got %q", expected, root)
	}
}
