package cryptolib

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"strings"
)

// https://github.com/cpacia/bchutil/blob/master/cashaddr.go

var (
	prefixes map[string]string
)

func init() {
	prefixes = make(map[string]string)
	prefixes["BCH"] = "bitcoincash"
	prefixes["TCH"] = "bchtest"
	prefixes["RCH"] = "bchreg"
}

// CHARSET is the cashaddr character set for encoding.
const CHARSET string = "qpzry9x8gf2tvdw0s3jn54khce6mua7l"

var charsetRev = [128]int8{
	-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
	-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
	-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, 15, -1, 10, 17, 21, 20, 26, 30, 7,
	5, -1, -1, -1, -1, -1, -1, -1, 29, -1, 24, 13, 25, 9, 8, 23, -1, 18, 22,
	31, 27, 19, -1, 1, 0, 3, 16, 11, 28, 12, 14, 6, 4, 2, -1, -1, -1, -1,
	-1, -1, 29, -1, 24, 13, 25, 9, 8, 23, -1, 18, 22, 31, 27, 19, -1, 1, 0,
	3, 16, 11, 28, 12, 14, 6, 4, 2, -1, -1, -1, -1, -1,
}

type data []byte

// ValidateWalletAddress will check that the address is valid for the specified
// coin.
func ValidateWalletAddress(coin string, address string) (bool, error) {
	if coin == "BSV" || coin == "TSV" || coin == "RSV" {
		return validateBSVWalletAddress(coin, address)
	} else if coin == "BCH" || coin == "TCH" || coin == "RCH" {
		return validateBCHWalletAddress(coin, address)
	} else if coin == "BTC" || coin == "TTC" || coin == "RTC" {
		return validA58(coin, []byte(address))
	} else {
		return false, errors.New("Invalid coin")
	}

}

func validateBCHWalletAddress(coin string, address string) (bool, error) {
	// Check the prefix is valid
	p := prefixes[coin] + ":"

	if !strings.HasPrefix(address, p) {
		return validA58(coin, []byte(address))
	}

	// Check each character is valid
	for _, c := range address[len(p):] {
		if !strings.Contains(CHARSET, string(c)) {
			return false, errors.New("Non valid character")
		}
	}

	// Check the checksum
	lower := false
	upper := false
	prefixSize := 0

	for i := 0; i < len(address); i++ {
		c := byte(address[i])
		if c >= 'a' && c <= 'z' {
			lower = true
			continue
		}

		if c >= 'A' && c <= 'Z' {
			upper = true
			continue
		}

		if c >= '0' && c <= '9' {
			// We cannot have numbers in the prefix.
			if prefixSize == 0 {
				return false, errors.New("Addresses cannot have numbers in the prefix")
			}

			continue
		}

		if c == ':' {
			// The separator must not be the first character, and there must not
			// be 2 separators.
			if i == 0 || prefixSize != 0 {
				return false, errors.New("The separator must not be the first character")
			}

			prefixSize = i
			continue
		}

		// We have an unexpected character.
		return false, errors.New("Unexpected character")
	}

	// We must have a prefix and a data part and we can't have both uppercase
	// and lowercase.
	if prefixSize == 0 {
		return false, errors.New("Address must have a prefix")
	}

	if upper && lower {
		return false, errors.New("Addresses cannot use both upper and lower case characters")
	}

	// Get the prefix.
	var prefix string
	for i := 0; i < prefixSize; i++ {
		prefix += string(lowerCase(address[i]))
	}

	// Decode values.
	valuesSize := len(address) - 1 - prefixSize
	values := make(data, valuesSize)
	for i := 0; i < valuesSize; i++ {
		c := byte(address[i+prefixSize+1])
		// We have an invalid char in there.
		if c > 127 || charsetRev[c] == -1 {
			return false, errors.New("Invalid character")
		}

		values[i] = byte(charsetRev[c])
	}

	// Verify the checksum.
	if !verifyChecksum(prefix, values) {
		return false, errors.New("checksum mismatch")
	}

	return true, nil
}

func validateBSVWalletAddress(coin string, address string) (bool, error) {
	if strings.HasPrefix(address, "bitcoin-script:") {
		_, _, network, _, err := DecodeBIP276(address)

		if err != nil {
			return false, fmt.Errorf("bitcoin-script invalid [%+v]", err)
		}

		if network == 1 && coin[0] != 'B' {
			return false, fmt.Errorf("bitcoin-script is for mainnet but coin is %s", coin)
		} else if network == 2 && coin[0] == 'B' {
			return false, fmt.Errorf("bitcoin-script is not for mainnet but coin is %s", coin)
		}

		return true, nil
	}

	return validBSVA58(coin, []byte(address))
}

func polyMod(v []byte) uint64 {
	c := uint64(1)
	for _, d := range v {
		c0 := byte(c >> 35)

		c = ((c & 0x07ffffffff) << 5) ^ uint64(d)

		if c0&0x01 > 0 {
			c ^= 0x98f2bc8e61
		}

		if c0&0x02 > 0 {
			c ^= 0x79b76d99e2
		}

		if c0&0x04 > 0 {
			c ^= 0xf33e5fb3c4
		}

		if c0&0x08 > 0 {
			c ^= 0xae2eabe2a8
		}

		if c0&0x10 > 0 {
			c ^= 0x1e4f43e470
		}
	}

	return c ^ 1
}

func lowerCase(c byte) byte {
	// ASCII black magic.
	return c | 0x20
}

/**
 * Expand the address prefix for the checksum computation.
 */
func expandPrefix(prefix string) data {
	ret := make(data, len(prefix)+1)
	for i := 0; i < len(prefix); i++ {
		ret[i] = byte(prefix[i]) & 0x1f
	}

	ret[len(prefix)] = 0
	return ret
}

func verifyChecksum(prefix string, payload data) bool {
	return polyMod(cat(expandPrefix(prefix), payload)) == 0
}

func cat(x, y data) data {
	return append(x, y...)
}

type a25 [25]byte

func (a *a25) embeddedChecksum() (c [4]byte) {
	copy(c[:], a[21:])
	return
}

// DoubleSHA256 computes a double sha256 hash of the first 21 bytes of the
// address.  This is the one function shared with the other bitcoin RC task.
// Returned is the full 32 byte sha256 hash.  (The bitcoin checksum will be
// the first four bytes of the slice.)
func (a *a25) doubleSHa256() []byte {
	h := sha256.New()
	h.Write(a[:21])
	d := h.Sum([]byte{})
	h = sha256.New()
	h.Write(d)
	return h.Sum(d[:0])
}

// computeChecksum returns a four byte checksum computed from the first 21
// bytes of the address.  The embedded checksum is not updated.
func (a *a25) computeChecksum() (c [4]byte) {
	copy(c[:], a.doubleSHa256())
	return
} /* {{header|Go}} */

// Tmpl and set58 are adapted from the C solution.
// Go has big integers but this techinique seems better.
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

// ValidA58 validates a base58 encoded bitcoin address.  An address is valid
// if it can be decoded into a 25 byte address, the version number is 0,
// and the checksum validates.  Return value ok will be true for valid
// addresses.  If ok is false, the address is invalid and the error value
// may indicate why.
func validA58(coin string, a58 []byte) (bool, error) {
	var a a25
	if err := a.set58(a58); err != nil {
		return false, err
	}
	if a[0] != 0 && a[0] != 5 && a[0] != 0x6f && a[0] != 0xc4 {
		return false, errors.New("not version 0 or 5, 6f or c4")
	}

	checksumOK := a.embeddedChecksum() == a.computeChecksum()

	if !checksumOK {
		return false, errors.New("checksum failed")
	}

	if (a[0] == 0 || a[0] == 5) && coin[0] != 'B' {
		return false, fmt.Errorf("address is for mainnet but coin is %s", coin)
	}

	if (a[0] == 0x6f || a[0] == 0xc4) && coin[0] == 'B' {
		return false, fmt.Errorf("address is not for mainnet but coin is %s", coin)
	}

	return true, nil
}

func validBSVA58(coin string, a58 []byte) (bool, error) {
	var a a25
	if err := a.set58(a58); err != nil {
		return false, err
	}
	if a[0] != 0 && a[0] != 0x6f {
		return false, errors.New("not version 0 or 6f")
	}

	checksumOK := a.embeddedChecksum() == a.computeChecksum()

	if !checksumOK {
		return false, errors.New("checksum failed")
	}

	if a[0] == 0 && coin[0] != 'B' {
		return false, fmt.Errorf("address is for mainnet but coin is %s", coin)
	}

	if a[0] == 0x6f && coin[0] == 'B' {
		return false, fmt.Errorf("address is not for mainnet but coin is %s", coin)
	}

	return true, nil
}
