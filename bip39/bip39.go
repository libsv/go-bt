// Package bip39 implements the bip39 protocol https://github.com/bitcoin/bips/blob/master/bip-0039.mediawiki
// This protocol relates closely with bip32 (https://github.com/bitcoin/bips/blob/master/bip-0032.mediawiki)
// and allows a memorable word list to be
// generated for a user of a wallet to store. This is a bit more user friendly
// than users dealing with hex strings etc.
//
// Users can supply an additional entropy by supplying an optional passcode which
// is used to generate the seed.
//
// The seed can then be passed to an hd (bip32) private key generation function to produce the
// wallet private master key.
package bip39

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/crypto/pbkdf2"
)

// Sentinel errors raised by the lib.
var(
	ErrInvalidLength = errors.New("invalid number of bits requested, " +
		"should be a multiple of 32 and between 128 and 256 (inclusive)")
	ErrInvalidMnemonic = errors.New("invalid mnemonic; length should be a multiple of 3 between 12 and 24")
	ErrInvalidWordlist = errors.New("mnemonic contains invalid word (word not in approved wordlist)")
)

// Entropy contains valid entropy lengths, these are multiples of 32 and between 128 & 256.
type Entropy uint32

// Allowed mnemonic lengths to supply to the GenerateEntropy function.
const(
	EntWords12 Entropy = 128
	EntWords15 Entropy = 160
	EntWords18 Entropy = 192
	EntWords21 Entropy = 224
	EntWords24 Entropy = 256
)

// GenerateEntropy will generate a bytearray of cryptographically random bytes
// with the length of bytes determined by the length supplied.
//
// If the length is invalid a bip39.ErrInvalidLength will be returned.
func GenerateEntropy(length Entropy) ([]byte, error){
	if length%32 != 0 || length < 128 || length > 256 {
		return nil, ErrInvalidLength
	}
	bb := make([]byte, length/8)
	_, _ = rand.Read(bb)
	return bb, nil
}


// Mnemonic will create a new mnemonic sentence using the supplied
// entropy bytes, if the entropy supplied is not a valid length a bip39.ErrInvalidLength
// error will be returned.
//
// An optional passcode can be supplied which will be used in the returned seed
// to provide an additional security measure.
func Mnemonic(entropy []byte, passcode string) (mnemonic string, seed []byte, e error){
	ent := len(entropy) * 8
	if ent%32 != 0 || ent < 128 || ent > 256 {
		return "", nil, ErrInvalidLength
	}
	cs := ent / 32
	ms := ent + cs
	entropy = append(entropy, sha256.Sum256(entropy)[0])
	sb := strings.Builder{}
	sb.Grow(ms)
	for _, b := range entropy {
		for t := 7; t >= 0; t--{
			if b & (1 << t) != 0{
				sb.WriteString("1")
				continue
			}
			sb.WriteString("0")
		}
	}
	bitString := sb.String()
	words := make([]string, 0, ms/11)
	for i := 11; i <= ms; i += 11 {
		output, err := strconv.ParseInt(bitString[i-11:i], 2, 32)
		if err != nil{
			return "", nil, fmt.Errorf("failed to convert binary to int %w", err)
		}
		words = append(words, English[output])
	}
	m := strings.Join(words, " ")
	return m, pbkdf2.Key([]byte(m), []byte("mnemonic"+passcode),2048, 64, sha512.New), nil
}

// MnemonicToSeed will validate a mnemonic and then generate the seed to be used
// in a BIP32 masterkey generation call.
//
// This can be used if re-generating a wallet from an existing mnemonic.
func MnemonicToSeed(words, passcode string) ([]byte, error){
	wl := strings.Fields(words)
	wlen := len(wl)
	if wlen%3 != 0 || wlen <12 || wlen >24{
		return nil, ErrInvalidMnemonic
	}

	for _, w := range wl{
		idx := sort.SearchStrings(English, w)
		if English[idx] == w{
			continue
		}
		return nil, ErrInvalidWordlist
	}
	return pbkdf2.Key([]byte(words), []byte("mnemonic"+passcode),2048, 64, sha512.New), nil
}
