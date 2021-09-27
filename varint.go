package bt

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

// VarInt takes an unsigned integer and  returns a byte array in VarInt format.
// See http://learnmeabitcoin.com/glossary/varint
func VarInt(i uint64) []byte {
	b := make([]byte, 9)
	if i < 0xfd {
		b[0] = byte(i)
		return b[:1]
	}
	if i < 0x10000 {
		b[0] = 0xfd
		binary.LittleEndian.PutUint16(b[1:3], uint16(i))
		return b[:3]
	}
	if i < 0x100000000 {
		b[0] = 0xfe
		binary.LittleEndian.PutUint32(b[1:5], uint32(i))
		return b[:5]
	}
	b[0] = 0xff
	binary.LittleEndian.PutUint64(b[1:9], i)
	return b
}

// DecodeVarInt takes a byte array in VarInt format and returns the
// decoded unsigned integer value of the length and it's size in bytes.
// See http://learnmeabitcoin.com/glossary/varint
func DecodeVarInt(b []byte) (uint64, int) {
	r := bytes.NewReader(b)

	result, size, err := DecodeVarIntFromReader(r)
	if err != nil {
		return 0, size
	}

	return result, size
}

// DecodeVarIntFromReader takes an io.Reader and returns the
// decoded unsigned integer value of the length.
// See http://learnmeabitcoin.com/glossary/varint
func DecodeVarIntFromReader(r io.Reader) (uint64, int, error) {
	b := make([]byte, 1)
	if n, err := io.ReadFull(r, b); n != 1 || err != nil {
		return 0, 0, fmt.Errorf("Could not read varint type, got %d bytes and err: %w", n, err)
	}

	switch b[0] {
	case 0xff:
		bb := make([]byte, 8)
		if n, err := io.ReadFull(r, bb); n != 8 || err != nil {
			return 0, 9, fmt.Errorf("Could not read varint(8), got %d bytes and err: %w", n, err)
		}
		return binary.LittleEndian.Uint64(bb), 9, nil

	case 0xfe:
		bb := make([]byte, 4)
		if n, err := io.ReadFull(r, bb); n != 4 || err != nil {
			return 0, 5, fmt.Errorf("Could not read varint(4), got %d bytes and err: %w", n, err)
		}
		return uint64(binary.LittleEndian.Uint32(bb)), 5, nil

	case 0xfd:
		bb := make([]byte, 2)
		if n, err := io.ReadFull(r, bb); n != 2 || err != nil {
			return 0, 3, fmt.Errorf("Could not read varint(2), got %d bytes and err: %w", n, err)
		}
		return uint64(binary.LittleEndian.Uint16(bb)), 3, nil

	default:
		return uint64(binary.LittleEndian.Uint16([]byte{b[0], 0x00})), 1, nil
	}
}

// VarIntUpperLimitInc returns true if a number is at the
// upper limit of a VarInt and will result in a VarInt
// length change if incremented. The value returned will
// indicate how many bytes will be increase if the length
// in incremented. -1 will be returned when the upper limit
// of VarInt is reached.
func VarIntUpperLimitInc(length uint64) int {
	switch length {
	case 252, 65535:
		return 2
	case 4294967295:
		return 4
	case 18446744073709551615:
		return -1
	}
	return 0
}
