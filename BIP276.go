package cryptolib

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"strconv"
)

const prefixScript = "bitcoin-script"
const prefixTemplate = "bitcoin-template"

const currentVersion = 1

const networkMainnet = 1
const networkTestnet = 2

func sha256d(b []byte) []byte {
	first := sha256.Sum256(b)
	second := sha256.Sum256(first[:])
	return second[:]
}

// EncodeBIP276 comment
func EncodeBIP276(prefix string, network int, version int, data []byte) string {
	if version == 0 || version > 255 || network == 0 || network > 255 {
		return "ERROR"
	}

	p, c := createBIP276(prefix, network, version, data)

	return p + c
}

func createBIP276(prefix string, network int, version int, data []byte) (string, string) {
	payload := fmt.Sprintf("%s:%.2x%.2x%x", prefix, network, version, data)
	checksum := hex.EncodeToString(sha256d([]byte(payload))[:4])

	return payload, checksum
}

// DecodeBIP276 comment
func DecodeBIP276(text string) (prefix string, version int, network int, data []byte, err error) {
	validBIP276 := regexp.MustCompile(`^(.+?):(\d{2})(\d{2})([0-9A-Fa-f]+)([0-9A-Fa-f]{8})$`)

	res := validBIP276.FindStringSubmatch(text)

	prefix = res[1]

	if version, err = strconv.Atoi(res[2]); err != nil {
		return
	}

	if network, err = strconv.Atoi(res[3]); err != nil {
		return
	}

	if data, err = hex.DecodeString(res[4]); err != nil {
		return
	}

	_, checksum := createBIP276(prefix, network, version, data)
	if res[5] != checksum {
		err = errors.New("Invalid checksum")
		return
	}

	return
}
