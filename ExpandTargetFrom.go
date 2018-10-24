package cryptolib

import (
	"encoding/binary"
	"encoding/hex"
	"math/big"
)

// ExpandTargetFrom comment
func ExpandTargetFrom(bits string) (string, error) {
	bn, err := ExpandTargetFromAsInt(bits)
	if err != nil {
		return "", err
	}

	dst := make([]byte, 32)
	b := bn.Bytes()

	copy(dst[32-len(b):], b)
	return hex.EncodeToString(dst), nil
}

// ExpandTargetFromAsInt comment
func ExpandTargetFromAsInt(bits string) (*big.Int, error) {
	binaryBits, err := hex.DecodeString(bits)
	if err != nil {
		return nil, err
	}
	compact := binary.BigEndian.Uint32(binaryBits)

	// Extract the mantissa, sign bit, and exponent.
	mantissa := compact & 0x007fffff
	isNegative := compact&0x00800000 != 0
	exponent := uint(compact >> 24)

	// Since the base for the exponent is 256, the exponent can be treated
	// as the number of bytes to represent the full 256-bit number.  So,
	// treat the exponent as the number of bytes and shift the mantissa
	// right or left accordingly.  This is equivalent to:
	// N = mantissa * 256^(exponent-3)
	var bn *big.Int
	if exponent <= 3 {
		mantissa >>= 8 * (3 - exponent)
		bn = big.NewInt(int64(mantissa))
	} else {
		bn = big.NewInt(int64(mantissa))
		bn.Lsh(bn, 8*(exponent-3))
	}

	// Make it negative if the sign bit is set.
	if isNegative {
		bn = bn.Neg(bn)
	}

	return bn, nil
}
