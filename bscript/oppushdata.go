package bscript

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
)

// EncodeParts takes an array of byte slices and returns a single byte
// slice with the appropriate OP_PUSH commands embedded. The output
// can be encoded to a hex string and viewed as a BitCoin script hex
// string.
//
// For example '76a9140d6cf2ef7bc915d109f77357a71b64fc25e2e11488ac' is
// the hex string of a P2PKH output script.
func EncodeParts(parts [][]byte) ([]byte, error) {
	b := make([]byte, 0)

	for i, part := range parts {
		pd, err := GetPushDataPrefix(part)
		if err != nil {
			return nil, fmt.Errorf("part %d is too big", i)
		}

		b = append(b, pd...)
		b = append(b, part...)
	}

	return b, nil
}

// GetPushDataPrefix takes a single byte slice of data and returns its
// OP_PUSHDATA BitCoin encoding prefix based on its length.
//
// For example, the data byte slice '022a8c1a18378885db9054676f17a27f4219045e'
// would be encoded as '14022a8c1a18378885db9054676f17a27f4219045e' in BitCoin.
// The OP_PUSHDATA prefix is '14' since the length of the data is
// 20 bytes (0x14 in decimal is 20).
func GetPushDataPrefix(data []byte) ([]byte, error) {
	b := make([]byte, 0)
	l := int64(len(data))

	if l <= 75 {
		b = append(b, byte(l))

	} else if l <= 0xFF {
		b = append(b, 0x4c) // OP_PUSHDATA1
		b = append(b, byte(len(data)))

	} else if l <= 0xFFFF {
		b = append(b, 0x4d) // OP_PUSHDATA2
		lenBuf := make([]byte, 2)
		binary.LittleEndian.PutUint16(lenBuf, uint16(len(data)))
		b = append(b, lenBuf...)

	} else if l <= 0xFFFFFFFF { // bt.DefaultSequenceNumber
		b = append(b, 0x4e) // OP_PUSHDATA4
		lenBuf := make([]byte, 4)
		binary.LittleEndian.PutUint32(lenBuf, uint32(len(data)))
		b = append(b, lenBuf...)

	} else {
		return nil, fmt.Errorf("data too big")
	}

	return b, nil
}

// DecodeStringParts takes a hex string and decodes the opcodes in it
// returning an array of opcode parts (which could be opcodes or data
// pushed to the stack).
func DecodeStringParts(s string) ([][]byte, error) {
	b, err := hex.DecodeString(s)
	if err != nil {
		return nil, err
	}
	return DecodeParts(b)
}

// DecodeParts takes bytes and decodes the opcodes in it
// returning an array of opcode parts (which could be opcodes or data
// pushed to the stack).
func DecodeParts(b []byte) ([][]byte, error) {
	var r [][]byte
	for len(b) > 0 {
		// Handle OP codes
		switch b[0] {
		case OpPUSHDATA1:
			if len(b) < 2 {
				return r, errors.New("not enough data")
			}
			l := uint64(b[1])
			if len(b) < int(2+l) {
				return r, errors.New("not enough data")
			}
			part := b[2 : 2+l]
			r = append(r, part)
			b = b[2+l:]

		case OpPUSHDATA2:
			if len(b) < 3 {
				return r, errors.New("not enough data")
			}
			l := binary.LittleEndian.Uint16(b[1:])
			if len(b) < int(3+l) {
				return r, errors.New("not enough data")
			}
			part := b[3 : 3+l]
			r = append(r, part)
			b = b[3+l:]

		case OpPUSHDATA4:
			if len(b) < 5 {
				return r, errors.New("Not enough data")
			}
			l := binary.LittleEndian.Uint32(b[1:])
			if len(b) < int(5+l) {
				return r, errors.New("not enough data")
			}
			part := b[5 : 5+l]
			r = append(r, part)
			b = b[5+l:]

		default:
			if b[0] >= 0x01 && b[0] <= 0x4e {
				l := b[0]
				if len(b) < int(1+l) {
					return r, errors.New("not enough data")
				}
				part := b[1 : l+1]
				r = append(r, part)
				b = b[1+l:]
			} else {
				r = append(r, []byte{b[0]})
				b = b[1:]
			}
		}
	}

	return r, nil
}
