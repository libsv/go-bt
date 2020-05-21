package script

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
)

// EncodeParts takes a slice of slices and returns a single slice with the appropriate OP_PUSH commands embedded.
func EncodeParts(parts [][]byte) ([]byte, error) {
	b := make([]byte, 0)

	for i, part := range parts {
		l := int64(len(part))

		if l <= 75 {
			b = append(b, byte(len(part)))
			b = append(b, part...)

		} else if l <= 0xFF {
			b = append(b, 0x4c) // OP_PUSHDATA1
			b = append(b, byte(len(part)))
			b = append(b, part...)

		} else if l <= 0xFFFF {
			b = append(b, 0x4d) // OP_PUSHDATA2
			lenBuf := make([]byte, 2)
			binary.LittleEndian.PutUint16(lenBuf, uint16(len(part)))
			b = append(b, lenBuf...)
			b = append(b, part...)

		} else if l <= 0xFFFFFFFF {
			b = append(b, 0x4e) // OP_PUSHDATA4
			lenBuf := make([]byte, 4)
			binary.LittleEndian.PutUint32(lenBuf, uint32(len(part)))
			b = append(b, lenBuf...)
			b = append(b, part...)

		} else {
			return nil, fmt.Errorf("Part %d is too big", i)
		}
	}

	return b, nil
}

// DecodeStringParts calls DecodeParts.
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
		case OP_PUSHDATA1:
			length := b[1]
			part := b[2 : 2+length]
			r = append(r, part)
			b = b[2+length:]

		case OP_PUSHDATA2:
			length := binary.LittleEndian.Uint16(b[1:])
			part := b[3 : 3+length]
			r = append(r, part)
			b = b[3+length:]

		case OP_PUSHDATA4:
			length := binary.LittleEndian.Uint32(b[1:])
			part := b[5 : 5+length]
			r = append(r, part)
			b = b[5+length:]

		default:
			if b[0] >= 0x01 && b[0] <= 0x4e {
				length := b[0]
				part := b[1 : length+1]
				r = append(r, part)
				b = b[1+length:]
			} else {
				r = append(r, []byte{b[0]})
				b = b[1:]
			}
		}
	}

	return r, nil
}
