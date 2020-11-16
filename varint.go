package bt

import "encoding/binary"

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
func DecodeVarInt(b []byte) (result uint64, size int) {
	switch b[0] {
	case 0xff:
		result = binary.LittleEndian.Uint64(b[1:9])
		size = 9

	case 0xfe:
		result = uint64(binary.LittleEndian.Uint32(b[1:5]))
		size = 5

	case 0xfd:
		result = uint64(binary.LittleEndian.Uint16(b[1:3]))
		size = 3

	default:
		result = uint64(binary.LittleEndian.Uint16([]byte{b[0], 0x00}))
		size = 1
	}

	return
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
