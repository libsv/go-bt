package bt

import (
	"errors"

	mapi "github.com/bitcoin-sv/merchantapi-reference/utils"
)

const (
	// FeeTypeStandard is the fee type for standard tx parts
	FeeTypeStandard = "standard"

	// FeeTypeData is the fee type for data tx parts
	FeeTypeData = "data"
)

// DefaultStandardFee returns the default
// standard fees offered by most miners.
func DefaultStandardFee() *mapi.Fee {
	return &mapi.Fee{
		FeeType: FeeTypeStandard,
		MiningFee: mapi.FeeUnit{
			Satoshis: 5,
			Bytes:    10,
		},
		RelayFee: mapi.FeeUnit{
			Satoshis: 5,
			Bytes:    10,
		},
	}
}

// DefaultDataFee returns the default
// data fees offered by most miners.
func DefaultDataFee() *mapi.Fee {
	return &mapi.Fee{
		FeeType: FeeTypeData,
		MiningFee: mapi.FeeUnit{
			Satoshis: 25,
			Bytes:    100,
		},
		RelayFee: mapi.FeeUnit{
			Satoshis: 25,
			Bytes:    100,
		},
	}
}

// DefaultFees returns an array of the default
// standard and data fees offered by most miners.
func DefaultFees() (f []*mapi.Fee) {
	f = append(f, DefaultStandardFee())
	f = append(f, DefaultDataFee())
	return
}

// GetStandardFee returns the standard fee in the fees array supplied.
func GetStandardFee(fees []*mapi.Fee) (*mapi.Fee, error) {
	for _, f := range fees {
		if f.FeeType == FeeTypeStandard {
			return f, nil
		}
	}

	return nil, errors.New("no " + FeeTypeStandard + " fee supplied")
}

// GetDataFee returns the data fee in the fees array supplied.
func GetDataFee(fees []*mapi.Fee) (*mapi.Fee, error) {
	for _, f := range fees {
		if f.FeeType == FeeTypeData {
			return f, nil
		}
	}

	return nil, errors.New("no " + FeeTypeData + " fee supplied")
}
