package main

import (
	"testing"

	"github.com/tyler-smith/go-bip39"
)



func Benchmark_other(b *testing.B){
	for n := 0; n < b.N; n++ {
		entropy, _ := bip39.NewEntropy(128)
		_, _ = bip39.NewMnemonic(entropy)
	}
}