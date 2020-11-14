package bt

import (
	"errors"

	mapi "github.com/bitcoin-sv/merchantapi-reference/utils"
)

const feeStandard = "standard"
const feeData = "data"

// DefaultStandard returns the default
// standard fees offered by most miners.
func DefaultStandard() *mapi.Fee {
	return &mapi.Fee{
		FeeType: feeStandard,
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

// DefaultData returns the default
// data fees offered by most miners.
func DefaultData() *mapi.Fee {
	return &mapi.Fee{
		FeeType: feeData,
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

// Default returns an array of the default
// standard and data fees offered by most miners.
func Default() (f []*mapi.Fee) {
	f = append(f, DefaultStandard())
	f = append(f, DefaultData())
	return
}

// GetStandardFee returns the standard fee in the fees array supplied.
func GetStandardFee(fees []*mapi.Fee) (*mapi.Fee, error) {
	for _, f := range fees {
		if f.FeeType == feeStandard {
			return f, nil
		}
	}

	return nil, errors.New("no " + feeStandard + " fee supplied")
}

// GetDataFee returns the data fee in the fees array supplied.
func GetDataFee(fees []*mapi.Fee) (*mapi.Fee, error) {
	for _, f := range fees {
		if f.FeeType == feeData {
			return f, nil
		}
	}

	return nil, errors.New("no " + feeData + " fee supplied")
}
