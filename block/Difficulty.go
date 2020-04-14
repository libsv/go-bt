package block

import (
	"encoding/binary"
	"encoding/hex"
	"log"
	"math"
	"math/big"
	"strconv"
)

var (
	genesisDiff float64
	regDiff     float64
)

func init() {
	bn, err := ExpandTargetFromAsInt("1d00ffff")
	if err != nil {
		log.Printf("Error: %+v", err)
	}

	bf := big.NewFloat(0)
	bf.SetInt(bn)
	genesisDiff, _ = bf.Float64()
	genesisDiff = math.Pow(2, 256) / genesisDiff

	bn, err = ExpandTargetFromAsInt("207fffff")
	if err != nil {
		log.Printf("Error: %+v", err)
	}

	bf.SetInt(bn)
	regDiff, _ = bf.Float64()
	regDiff = math.Pow(2, 256) / regDiff
}

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

// DifficultyToHashrate takes a specific coin ticker, it's difficulty, and target
// and computes the estimated hashrate on that specific coin (or chain).
func DifficultyToHashrate(coin string, diff uint64, targetSeconds float64) float64 {
	genesis := genesisDiff
	if coin[0] == 'R' {
		genesis = regDiff
	}

	return float64(diff) * genesis / targetSeconds
}

// DifficultyFromBits returns the mining difficulty from the nBits field in the block header.
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
