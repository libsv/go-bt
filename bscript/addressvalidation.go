package bscript

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"github.com/libsv/go-bt/crypto"
)

type a25 [25]byte

func (a *a25) embeddedChecksum() (c [4]byte) {
	copy(c[:], a[21:])
	return
}

// computeChecksum returns a four byte checksum computed from the first 21
// bytes of the address.  The embedded checksum is not updated.
func (a *a25) computeChecksum() (c [4]byte) {
	copy(c[:], crypto.Sha256d(a[:21]))
	return
}

// Tmpl and set58 are adapted from the C solution.
// Go has big integers but this technique seems better.
var tmpl = []byte("123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz")

// set58 takes a base58 encoded address and decodes it into the receiver.
// Errors are returned if the argument is not valid base58 or if the decoded
// value does not fit in the 25 byte address.  The address is not otherwise
// checked for validity.
func (a *a25) set58(s []byte) error {
	for _, s1 := range s {
		c := bytes.IndexByte(tmpl, s1)
		if c < 0 {
			return errors.New("bad char")
		}
		for j := 24; j >= 0; j-- {
			c += 58 * int(a[j])
			a[j] = byte(c % 256)
			c /= 256
		}
		if c > 0 {
			return errors.New("too long")
		}
	}
	return nil
}

// ValidateAddress checks if an address string is a valid BitCoin address (ex. P2PKH, BIP276).
// Checks both mainnet and testnet.
func ValidateAddress(address string) (bool, error) {
	if strings.HasPrefix(address, "bitcoin-script:") {
		_, _, _, _, err := DecodeBIP276(address)

		if err != nil {
			return false, fmt.Errorf("bitcoin-script invalid [%w]", err)
		}
		return true, nil
	}

	return validA58([]byte(address))
}

func validA58(a58 []byte) (bool, error) {
	var a a25
	if err := a.set58(a58); err != nil {
		return false, err
	}
	if a[0] != 0 && a[0] != 0x6f {
		return false, errors.New("not version 0 or 6f")
	}

	if a.embeddedChecksum() != a.computeChecksum() {
		return false, errors.New("checksum failed")
	}

	return true, nil
}
