package utils

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/libsv/libsv/crypto"
	"regexp"
	"strconv"
)

const PrefixScript = "bitcoin-script"

//const prefixTemplate = "bitcoin-template"

const CurrentVersion = 1

const NetworkMainnet = 1

//const networkTestnet = 2

// TODO: consider moving to script or utils
// EncodeBIP276 is used to encode specific (non-standard) scripts in BIP276 format.
// See https://github.com/moneybutton/bips/blob/master/bip-0276.mediawiki
func EncodeBIP276(prefix string, network int, version int, data []byte) string {
	if version == 0 || version > 255 || network == 0 || network > 255 {
		return "ERROR"
	}

	p, c := createBIP276(prefix, network, version, data)

	return p + c
}

func createBIP276(prefix string, network int, version int, data []byte) (string, string) {
	payload := fmt.Sprintf("%s:%.2x%.2x%x", prefix, network, version, data)
	checksum := hex.EncodeToString(crypto.Sha256d([]byte(payload))[:4])

	return payload, checksum
}

// DecodeBIP276 is used to decode BIP276 formatted data into specific (non-standard) scripts.
// See https://github.com/moneybutton/bips/blob/master/bip-0276.mediawiki
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
