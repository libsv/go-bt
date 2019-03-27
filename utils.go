package cryptolib

import (
	"encoding/binary"
	"encoding/hex"
	"math/big"
	"strconv"
)

// DifficultyFromBits comment
func DifficultyFromBits(bits string) (float64, error) {
	b, _ := hex.DecodeString(bits)
	ib := binary.BigEndian.Uint32(b)
	return targetToDifficulty(toCompactSize(ib))
}

func toCompactSize(bits uint32) *big.Int {
	t := big.NewInt(int64(bits % 0x01000000))
	t.Mul(t, big.NewInt(2).Exp(big.NewInt(2), big.NewInt(8*(int64(bits/0x01000000)-3)), nil))

	return t
}

func targetToDifficulty(target *big.Int) (float64, error) {
	a := float64(0xFFFF0000000000000000000000000000000000000000000000000000) // genesis difficulty
	b, err := strconv.ParseFloat(target.String(), 64)
	if err != nil {
		return 0.0, err
	}
	return a / b, nil
}

// GetLittleEndianBytes comment
func GetLittleEndianBytes(v uint32, l uint32) []byte {
	// TODO: is v hex encoded?
	buf := make([]byte, l)

	binary.LittleEndian.PutUint32(buf, v)

	return buf
}

// VarInt See http://learnmeabitcoin.com/glossary/varint
func VarInt(i int) []byte {
	uvl := uint64(i)
	b := make([]byte, 9)
	if uvl < 0xfd {
		b[0] = byte(uvl)
		return b[:1]
	}
	if uvl < 0x10000 {
		b[0] = 0xfd
		binary.LittleEndian.PutUint16(b[1:3], uint16(uvl))
		return b[:3]
	}
	if uvl < 0x100000000 {
		b[0] = 0xfe
		binary.LittleEndian.PutUint32(b[1:5], uint32(uvl))
		return b[:5]
	}
	b[0] = 0xff
	binary.LittleEndian.PutUint64(b[1:9], uvl)
	return b
}

// DecodeVarInt comment
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

// DecodeStringParts calls DecodeParts
func DecodeStringParts(s string) ([][]byte, error) {
	b, err := hex.DecodeString(s)
	if err != nil {
		return nil, err
	}
	return DecodeParts(b)
}

// DecodeParts returns an array of strings...
func DecodeParts(b []byte) ([][]byte, error) {
	var r [][]byte
	for len(b) > 0 {
		// Handle OP codes
		switch b[0] {
		case opPUSHDATA1:
			len := b[1]
			part := b[2 : 2+len]
			r = append(r, part)
			b = b[2+len:]

		case opPUSHDATA2:
			len := binary.LittleEndian.Uint16(b[1:])
			part := b[3 : 3+len]
			r = append(r, part)
			b = b[3+len:]

		case opPUSHDATA4:
			len := binary.LittleEndian.Uint32(b[1:])
			part := b[5 : 5+len]
			r = append(r, part)
			b = b[5+len:]

		default:
			if b[0] >= 0x01 && b[0] <= 0x4e {
				len := b[0]
				part := b[1 : len+1]
				r = append(r, part)
				b = b[1+len:]
			} else {
				r = append(r, []byte{b[0]})
				b = b[1:]
			}
		}
	}

	return r, nil
}
