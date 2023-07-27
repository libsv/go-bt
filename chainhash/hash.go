// Copyright (c) 2013-2016 The btcsuite developers
// Copyright (c) 2015 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package chainhash

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

// HashSize of array used to store hashes.  See Hash.
const HashSize = 32

// MaxHashStringSize is the maximum length of a Hash hash string.
const MaxHashStringSize = HashSize * 2

// ErrHashStrSize describes an error that indicates the caller specified a hash
// string that has too many characters.
var ErrHashStrSize = fmt.Errorf("max hash string length is %v bytes", MaxHashStringSize)

// Hash is used in several of the bitcoin messages and common structures.  It
// typically represents the double sha256 of data.
type Hash [HashSize]byte

// String returns the Hash as the hexadecimal string of the byte-reversed
// hash.
func (hash Hash) String() string {
	for i := 0; i < HashSize/2; i++ {
		hash[i], hash[HashSize-1-i] = hash[HashSize-1-i], hash[i]
	}
	return hex.EncodeToString(hash[:])
}

// CloneBytes returns a copy of the bytes which represent the hash as a byte
// slice.
//
// NOTE: It is generally cheaper to just slice the hash directly thereby reusing
// the same bytes rather than calling this method.
func (hash *Hash) CloneBytes() []byte {
	newHash := make([]byte, HashSize)
	copy(newHash, hash[:])

	return newHash
}

// SetBytes sets the bytes which represent the hash.  An error is returned if
// the number of bytes passed in is not HashSize.
func (hash *Hash) SetBytes(newHash []byte) error {
	nhLen := len(newHash)
	if nhLen != HashSize {
		return fmt.Errorf("invalid hash length of %v, want %v", nhLen,
			HashSize)
	}
	copy(hash[:], newHash)

	return nil
}

// IsEqual returns true if target is the same as hash.
func (hash *Hash) IsEqual(target *Hash) bool {
	if hash == nil && target == nil {
		return true
	}
	if hash == nil || target == nil {
		return false
	}
	return *hash == *target
}

func (hash *Hash) MarshalJSON() ([]byte, error) {
	return json.Marshal(hash.String())
}

func (hash *Hash) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	h, err := NewHashFromStr(s)
	if err != nil {
		return err
	}
	*hash = *h
	return nil
}

// NewHash returns a new Hash from a byte slice.  An error is returned if
// the number of bytes passed in is not HashSize.
func NewHash(newHash []byte) (*Hash, error) {
	var sh Hash
	err := sh.SetBytes(newHash)
	if err != nil {
		return nil, err
	}
	return &sh, err
}

// func NewHashNoError(newHash []byte) *Hash {
// 	sh, _ := NewHash(newHash)
// 	return sh
// }

// NewHashFromStr creates a Hash from a hash string.  The string should be
// the hexadecimal string of a byte-reversed hash, but any missing characters
// result in zero padding at the end of the Hash.
func NewHashFromStr(hash string) (*Hash, error) {
	ret := new(Hash)
	err := Decode(ret, hash)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

// func NewHashFromStrNoError(hash string) *Hash {
// 	sh, _ := NewHashFromStr(hash)
// 	return sh
// }

// Decode decodes the byte-reversed hexadecimal string encoding of a Hash to a
// destination.
func Decode(dst *Hash, src string) error {
	// Return error if hash string is too long.
	if len(src) > MaxHashStringSize {
		return ErrHashStrSize
	}

	// Hex decoder expects the hash to be a multiple of two.  When not, pad
	// with a leading zero.
	var srcBytes []byte
	if len(src)%2 == 0 {
		srcBytes = []byte(src)
	} else {
		srcBytes = make([]byte, 1+len(src))
		srcBytes[0] = '0'
		copy(srcBytes[1:], src)
	}

	// Hex decode the source bytes to a temporary destination.
	var reversedHash Hash
	_, err := hex.Decode(reversedHash[HashSize-hex.DecodedLen(len(srcBytes)):], srcBytes)
	if err != nil {
		return err
	}

	// Reverse copy from the temporary hash to destination.  Because the
	// temporary was zeroed, the written result will be correctly padded.
	for i, b := range reversedHash[:HashSize/2] {
		dst[i], dst[HashSize-1-i] = reversedHash[HashSize-1-i], b
	}

	return nil
}
func (hash Hash) Marshal() ([]byte, error) {
	if len(hash) == 0 {
		return nil, nil
	}
	return hash.CloneBytes(), nil
}

func (hash Hash) MarshalTo(data []byte) (n int, err error) {
	if len(hash) == 0 {
		return 0, nil
	}
	copy(data, hash.CloneBytes())
	return 16, nil
}

func (hash *Hash) Unmarshal(data []byte) error {
	if len(data) == 0 {
		return nil
	}
	id := &Hash{}
	copy(id[:], data)
	*hash = *id
	return nil
}

func (hash Hash) Equal(other Hash) bool {
	return bytes.Equal(hash[0:], other[0:])
}

func (hash *Hash) Size() int {
	if hash == nil {
		return 0
	}

	return len(*hash)
}
