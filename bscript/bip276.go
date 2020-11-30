package bscript

import (
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"strconv"

	"github.com/libsv/go-bt/crypto"
)

// PrefixScript is the prefix in the BIP276 standard which
// specifies if it is a script or template.
const PrefixScript = "bitcoin-script"

// PrefixTemplate is the prefix in the BIP276 standard which
// specifies if it is a script or template.
const PrefixTemplate = "bitcoin-template"

// CurrentVersion provides the ability to
// update the structure of the data that
// follows it.
const CurrentVersion = 1

// NetworkMainnet specifies that the data is only
// valid for use on the main network.
const NetworkMainnet = 1

// NetworkTestnet specifies that the data is only
// valid for use on the test network.
const NetworkTestnet = 2

var validBIP276 = regexp.MustCompile(`^(.+?):(\d{2})(\d{2})([0-9A-Fa-f]+)([0-9A-Fa-f]{8})$`)

// EncodeBIP276 is used to encode specific (non-standard) scripts in BIP276 format.
// See https://github.com/moneybutton/bips/blob/master/bip-0276.mediawiki
func EncodeBIP276(prefix string, network, version int, data []byte) string {
	if version == 0 || version > 255 || network == 0 || network > 255 {
		return "ERROR"
	}

	p, c := createBIP276(prefix, network, version, data)

	return p + c
}

func createBIP276(prefix string, network, version int, data []byte) (string, string) {
	payload := fmt.Sprintf("%s:%.2x%.2x%x", prefix, network, version, data)
	return payload, hex.EncodeToString(crypto.Sha256d([]byte(payload))[:4])
}

// DecodeBIP276 is used to decode BIP276 formatted data into specific (non-standard) scripts.
// See https://github.com/moneybutton/bips/blob/master/bip-0276.mediawiki
func DecodeBIP276(text string) (prefix string, version, network int, data []byte, err error) {

	// Determine if regex match
	res := validBIP276.FindStringSubmatch(text)

	// Check if we got a result from the regex match first
	if len(res) == 0 {
		err = fmt.Errorf("text did not match the BIP276 format")
		return
	}

	// Set the prefix
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

	if _, checkSum := createBIP276(prefix, network, version, data); res[5] != checkSum {
		err = errors.New("invalid checksum")
	}

	return
}
