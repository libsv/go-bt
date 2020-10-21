package fees

import (
	mapi "github.com/bitcoin-sv/merchantapi-reference/utils"
)

// DefaultStandard returns the default
// standard fees offered by most miners.
func DefaultStandard() *mapi.Fee {
	return &mapi.Fee{
		FeeType: "standard",
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
		FeeType: "data",
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
