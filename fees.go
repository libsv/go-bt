package bt

import (
	"errors"
)

const (
	// FeeTypeStandard is the fee type for standard tx parts
	FeeTypeStandard = "standard"

	// FeeTypeData is the fee type for data tx parts
	FeeTypeData = "data"
)

// FeeUnit displays the amount of Satoshis needed
// for a specific amount of Bytes in a transaction
// see https://github.com/bitcoin-sv-specs/brfc-misc/tree/master/feespec
type FeeUnit struct {
	Satoshis int `json:"satoshis"` // Fee in satoshis of the amount of Bytes
	Bytes    int `json:"bytes"`    // Number of bytes that the Fee covers
}

// Fee displays the MiningFee as well as the RelayFee for a specific
// FeeType, for example 'standard' or 'data'
// see https://github.com/bitcoin-sv-specs/brfc-misc/tree/master/feespec
type Fee struct {
	FeeType   string  `json:"feeType"` // standard || data
	MiningFee FeeUnit `json:"miningFee"`
	RelayFee  FeeUnit `json:"relayFee"` // Fee for retaining Tx in secondary mempool
}

// DefaultStandardFee returns the default
// standard fees offered by most miners.
func DefaultStandardFee() *Fee {
	return &Fee{
		FeeType: FeeTypeStandard,
		MiningFee: FeeUnit{
			Satoshis: 5,
			Bytes:    10,
		},
		RelayFee: FeeUnit{
			Satoshis: 5,
			Bytes:    10,
		},
	}
}

// DefaultDataFee returns the default
// data fees offered by most miners.
func DefaultDataFee() *Fee {
	return &Fee{
		FeeType: FeeTypeData,
		MiningFee: FeeUnit{
			Satoshis: 5,
			Bytes:    10,
		},
		RelayFee: FeeUnit{
			Satoshis: 5,
			Bytes:    10,
		},
	}
}

// DefaultFees returns an array of the default
// standard and data fees offered by most miners.
func DefaultFees() (f []*Fee) {
	f = append(f, DefaultStandardFee())
	f = append(f, DefaultDataFee())
	return
}

// GetStandardFee returns the standard fee in the fees array supplied.
func GetStandardFee(fees []*Fee) (*Fee, error) {
	for _, f := range fees {
		if f.FeeType == FeeTypeStandard {
			return f, nil
		}
	}

	return nil, errors.New("no " + FeeTypeStandard + " fee supplied")
}

// GetDataFee returns the data fee in the fees array supplied.
func GetDataFee(fees []*Fee) (*Fee, error) {
	for _, f := range fees {
		if f.FeeType == FeeTypeData {
			return f, nil
		}
	}

	return nil, errors.New("no " + FeeTypeData + " fee supplied")
}
