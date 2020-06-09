package crypto

import (
	"bytes"
	"encoding/hex"
	"testing"
)

var hashTests = []struct {
	testName string
	input    string
	expected string
	hashFunc func([]byte) []byte
}{
	{"Test Ripemd160 Empty String", "", "9c1185a5c5e9fc54612808977ee8f548b2258d31", Ripemd160},
	{"Test Ripemd160 String", "I am a test", "09a23f506b4a37cabab8a9e49b541de582fca96b", Ripemd160},

	{"Test Sha256d Empty String", "", "5df6e0e2761359d30a8275058e299fcc0381534545f55cf43e41983f5d4c9456", Sha256d},
	{"Test Sha256 d String", "this is the data I want to hash", "2209ddda5914a3fbad507ff2284c4b6e559c18a669f9fc3ad3b5826a2a999d58", Sha256d},

	{"Test Sha256 Empty String", "", "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", Sha256},
	{"Test Sha256 String", "this is the data I want to hash", "f88eec7ecabf88f9a64c4100cac1e0c0c4581100492137d1b656ea626cad63e3", Sha256},
}

func TestFlagParser(t *testing.T) {

	for _, hashTest := range hashTests {
		t.Run(hashTest.testName, func(t *testing.T) {
			// converts input string to byte
			inputBytes := []byte(hashTest.input)
			// decode input string to byte
			expectedBytes, err := hex.DecodeString(hashTest.expected)
			if err != nil {
				t.Fatalf("Failed to decode string %s", hashTest.input)
			}

			hashResult := hashTest.hashFunc(inputBytes)
			if bytes.Compare(hashResult, expectedBytes) != 0 {
				t.Errorf("got %q, want %q", hashResult, hashTest.expected)
			}
		})
	}
}
