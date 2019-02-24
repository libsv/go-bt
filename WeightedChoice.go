package cryptolib

import (
	"crypto/rand"
	"errors"
	"math/big"
)

//Choice is used by WeightedChoice
type Choice struct {
	Weight int
	Item   interface{}
}

// WeightedChoice used weighted random selection to return one of the supplied
// choices.  Weights of 0 are never selected.  All other weight values are
// relative.  E.g. if you have two choices both weighted 3, they will be
// returned equally often; and each will be returned 3 times as often as a
// choice weighted 1.
func WeightedChoice(choices []Choice) (Choice, error) {
	// Based on this algorithm:
	//     http://eli.thegreenplace.net/2010/01/22/weighted-random-generation-in-python/
	var ret Choice

	sum := 0
	for _, c := range choices {
		sum += c.Weight
	}

	r, err := intRange(0, sum)
	if err != nil {
		return ret, err
	}

	for _, c := range choices {
		r -= c.Weight
		if r < 0 {
			return c, nil
		}
	}

	err = errors.New("Internal error - code should not reach this point")
	return ret, err
}

// intRange returns a random integer in the range from min to max.
func intRange(min, max int) (int, error) {
	var result int
	switch {
	case min > max:
		// Fail with error
		return result, errors.New("Min cannot be greater than max")
	case max == min:
		result = max
	case max > min:
		maxRand := max - min
		b, err := rand.Int(rand.Reader, big.NewInt(int64(maxRand)))
		if err != nil {
			return result, err
		}
		result = min + int(b.Int64())
	}
	return result, nil
}
