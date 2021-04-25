package main

import (
	"fmt"
	"time"

	"github.com/tyler-smith/go-bip39"

	mine "github.com/libsv/go-bt/bip39"
)

func main() {
	start := time.Now()
	_, _ = mine.Words(128)
	fmt.Println(time.Since(start).Nanoseconds())
	start = time.Now()
	entropy, _ := bip39.NewEntropy(128)
	_, _ = bip39.NewMnemonic(entropy)

	fmt.Println(time.Since(start).Nanoseconds())
}

