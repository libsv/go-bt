package bt

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// FeeType is used to specify which
// type of fee is used depending on
// the type of tx data (eg: standard
// bytes or data bytes).
type FeeType string

const (
	// FeeTypeStandard is the fee type for standard tx parts
	FeeTypeStandard FeeType = "standard"

	// FeeTypeData is the fee type for data tx parts
	FeeTypeData FeeType = "data"
)

// FeeQuotes contains a thread safe map of fees for standard and data
// fees as well as an expiry time.
//
// NewFeeQuote() should be called to get a
//
// When expiry expires ie Expired() == true then you should fetch
// new quotes from a MAPI server and call AddQuote with the fee information.
type FeeQuotes struct {
	mu         sync.RWMutex
	fees       map[FeeType]*Fee
	expiryTime time.Time
}

// NewFeeQuote will setup and return a new FeeQuotes struct which
// contains default fees when initially setup. You would then pass this
// data structure to a singleton struct via injection for reading.
//
//  fq := NewFeeQuote()
//
// The fees have an expiry time which, when initially setup, has an
// expiry of now.UTC. This allows you to check for fq.Expired() and if true
// fetch a new set of fees from a MAPI server. This means the first check
// will always fetch the latest fees. If you want to just use default fees
// always, you can ignore the expired method and simply call the fq.Fee() method.
// https://github.com/bitcoin-sv-specs/brfc-merchantapi#payload
//
// A basic example of usage is shown below:
//
//  func Fee(ft bt.FeeType) *bt.Fee{
//     // you would not call this every time - this is just an example
//     // you'd call this at app startup and store it / pass to a struct
//     fq := NewFeeQuote()
//
//     // fq setup with defaultFees
//     if !fq.Expired() {
//        // not expired, just return fee we have cached
//        return fe.Fee(ft)
//     }
//
//     // cache expired, fetch new quotes
//     var stdFee *bt.Fee
//     var dataFee *bt.Fee
//
//     // fetch quotes from MAPI server
//
//     fq.AddQuote(bt.FeeTypeStandard, stdFee)
//     fq.AddQuote(bt.FeeTypeData, dataFee)
//
//     // MAPI returns a quote expiry
//     exp, _ := time.Parse(time.RFC3339, resp.Quote.ExpirationTime)
//     fq.UpdateExpiry(exp)
//     return fe.Fee(ft)
//  }
// It will set the expiry time to now.UTC which when expires
// will indicate that new quotes should be fetched from a MAPI server.
func NewFeeQuote() *FeeQuotes {
	fq := &FeeQuotes{
		fees:       map[FeeType]*Fee{},
		expiryTime: time.Now().UTC(),
		mu: sync.RWMutex{},
	}
	fq.AddQuote(FeeTypeStandard, defaultStandardFee()).
		AddQuote(FeeTypeData, defaultDataFee())
	return fq
}


// Fee will return a fee by type if found, nil and an error if not.
func (f *FeeQuotes) Fee(t FeeType) (*Fee, error) {
	if f == nil{
		return nil, errors.New("feeQuotes have not been initialized, call NewFeeQuote()")
	}
	f.mu.RLock()
	defer f.mu.RUnlock()
	fee, ok := f.fees[t]
	if fee == nil || !ok{
		return nil, fmt.Errorf("feetype %s not found", t)
	}
	return fee, nil
}

// AddQuote will add new set of quotes for a feetype or update an existing
// quote if it already exists.
func (f *FeeQuotes) AddQuote(ft FeeType, fee *Fee) *FeeQuotes {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.fees[ft] = fee
	return f
}

// UpdateExpiry will update the expiry time of the quotes, this will be
// used when you fetch a fresh set of quotes from a MAPI server which
// should return an expiration time.
func (f *FeeQuotes) UpdateExpiry(exp time.Time) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.expiryTime = exp
}

// Expired will return true if the expiry time is before UTC now, this
// means we need to fetch fresh quotes from a MAPI server.
func (f *FeeQuotes) Expired() bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.expiryTime.Before(time.Now().UTC())
}

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
	FeeType   FeeType `json:"feeType"` // standard || data
	MiningFee FeeUnit `json:"miningFee"`
	RelayFee  FeeUnit `json:"relayFee"` // Fee for retaining Tx in secondary mempool
}

// defaultStandardFee returns the default
// standard fees offered by most miners.
func defaultStandardFee() *Fee {
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

// defaultDataFee returns the default
// data fees offered by most miners.
func defaultDataFee() *Fee {
	return &Fee{
		FeeType: FeeTypeData,
		MiningFee: FeeUnit{
			Satoshis: 25,
			Bytes:    100,
		},
		RelayFee: FeeUnit{
			Satoshis: 25,
			Bytes:    100,
		},
	}
}
