package cryptolib

import (
	"log"
	"math"
	"math/big"
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

// DifficultyToHashrate takes a specific coin ticker, it's difficulty, and target
// and compuptes the estimated hashrate on that specific coin (or chain).
func DifficultyToHashrate(coin string, diff uint64, targetSeconds float64) float64 {
	genesis := genesisDiff
	if coin[0] == 'R' {
		genesis = regDiff
	}

	return float64(diff) * genesis / targetSeconds
}
