package main

import (
	"testing"

	"github.com/tyler-smith/go-bip39"

	mine "github.com/libsv/go-bt/bip39"
)

func Benchmark_perfmine(b *testing.B){
	for n := 0; n < b.N; n++ {
		_, _ = mine.Words(128)
	}
}

func Benchmark_other(b *testing.B){
	for n := 0; n < b.N; n++ {
		entropy, _ := bip39.NewEntropy(128)
		_, _ = bip39.NewMnemonic(entropy)
	}
}