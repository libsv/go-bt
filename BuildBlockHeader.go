package cryptolib

import (
	"encoding/binary"
	"encoding/hex"
)

// BuildBlockHeader builds the block header byte array from the specific fields in the header.
func BuildBlockHeader(version uint32, previousBlockHash string, merkleRoot []byte, time []byte, bits []byte, nonce []byte) []byte {
	v := make([]byte, 4)
	binary.LittleEndian.PutUint32(v, version)
	p, _ := hex.DecodeString(previousBlockHash)

	p = ReverseBytes(p)

	a := []byte{}
	a = append(a, v...)
	a = append(a, p...)
	a = append(a, merkleRoot...)
	a = append(a, time...)
	a = append(a, bits...)
	a = append(a, nonce...)
	return a
}
