package bt

import (
	"encoding/binary"
	"io"

	"github.com/pkg/errors"
)

// VarInt (variable integer) is a field used in transaction data to indicate the number of
// upcoming fields, or the length of an upcoming field.
// See http://learnmeabitcoin.com/glossary/varint
type VarInt uint64

// NewVarIntFromBytes takes a byte array in VarInt format and returns the
// decoded unsigned integer value of the length, and it's size in bytes.
// See http://learnmeabitcoin.com/glossary/varint
func NewVarIntFromBytes(bb []byte) (VarInt, int) {
	switch bb[0] {
	case 0xff:
		return VarInt(binary.LittleEndian.Uint64(bb[1:9])), 9
	case 0xfe:
		return VarInt(binary.LittleEndian.Uint32(bb[1:5])), 5
	case 0xfd:
		return VarInt(binary.LittleEndian.Uint16(bb[1:3])), 3
	default:
		return VarInt(binary.LittleEndian.Uint16([]byte{bb[0], 0x00})), 1
	}
}

// Length return the length of the underlying byte representation of the `bt.VarInt`.
func (v VarInt) Length() int {
	if v < 253 {
		return 1
	}
	if v < 65536 {
		return 3
	}
	if v < 4294967296 {
		return 5
	}
	return 9
}

// Bytes takes the underlying unsigned integer and returns a byte array in VarInt format.
// See http://learnmeabitcoin.com/glossary/varint
func (v VarInt) Bytes() []byte {
	b := make([]byte, 9)
	if v < 0xfd {
		b[0] = byte(v)
		return b[:1]
	}
	if v < 0x10000 {
		b[0] = 0xfd
		binary.LittleEndian.PutUint16(b[1:3], uint16(v))
		return b[:3]
	}
	if v < 0x100000000 {
		b[0] = 0xfe
		binary.LittleEndian.PutUint32(b[1:5], uint32(v))
		return b[:5]
	}
	b[0] = 0xff
	binary.LittleEndian.PutUint64(b[1:9], uint64(v))
	return b
}

// ReadFrom reads the next varint from the io.Reader and assigned it to itself.
func (v *VarInt) ReadFrom(r io.Reader) (int64, error) {
	b := make([]byte, 1)
	if _, err := io.ReadFull(r, b); err != nil {
		return 0, errors.Wrap(err, "could not read varint type")
	}

	switch b[0] {
	case 0xff:
		bb := make([]byte, 8)
		if n, err := io.ReadFull(r, bb); err != nil {
			return 9, errors.Wrapf(err, "varint(8): got %d bytes", n)
		}
		*v = VarInt(binary.LittleEndian.Uint64(bb))
		return 9, nil

	case 0xfe:
		bb := make([]byte, 4)
		if n, err := io.ReadFull(r, bb); err != nil {
			return 5, errors.Wrapf(err, "varint(4): got %d bytes", n)
		}
		*v = VarInt(binary.LittleEndian.Uint32(bb))
		return 5, nil

	case 0xfd:
		bb := make([]byte, 2)
		if n, err := io.ReadFull(r, bb); err != nil {
			return 3, errors.Wrapf(err, "varint(2): got %d bytes", n)
		}
		*v = VarInt(binary.LittleEndian.Uint16(bb))
		return 3, nil

	default:
		*v = VarInt(binary.LittleEndian.Uint16([]byte{b[0], 0x00}))
		return 1, nil
	}
}

// UpperLimitInc returns true if a number is at the
// upper limit of a VarInt and will result in a VarInt
// length change if incremented. The value returned will
// indicate how many bytes will be increase if the length
// in incremented. -1 will be returned when the upper limit
// of VarInt is reached.
func (v VarInt) UpperLimitInc() int {
	switch uint64(v) {
	case 252, 65535:
		return 2
	case 4294967295:
		return 4
	case 18446744073709551615:
		return -1
	}

	return 0
}
