package bt

import (
	"encoding/json"
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

// FeeQuotes contains a list of miners and the current fees for each miner as well as their expiry.
//
// This can be used when getting fees from multiple miners, and you want to use the cheapest for example.
//
// Usage setup should be calling NewFeeQuotes(minerName).
type FeeQuotes struct {
	mu     sync.RWMutex
	quotes map[string]*FeeQuote
}

// NewFeeQuotes will set up default feeQuotes for the minerName supplied, ie TAAL etc.
func NewFeeQuotes(minerName string) *FeeQuotes {
	return &FeeQuotes{
		mu:     sync.RWMutex{},
		quotes: map[string]*FeeQuote{minerName: NewFeeQuote()},
	}
}

// AddMinerWithDefault will add a new miner to the quotes map with default fees & immediate expiry.
func (f *FeeQuotes) AddMinerWithDefault(minerName string) *FeeQuotes {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.quotes[minerName] = NewFeeQuote()
	return f
}

// AddMiner will add a new miner to the quotes map with the provided feeQuote.
// If you just want to add default fees use the AddMinerWithDefault method.
func (f *FeeQuotes) AddMiner(minerName string, quote *FeeQuote) *FeeQuotes {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.quotes[minerName] = quote
	return f
}

// Quote will return all fees for a miner.
// If no fees are found an ErrMinerNoQuotes error is returned.
func (f *FeeQuotes) Quote(minerName string) (*FeeQuote, error) {
	if f == nil {
		return nil, ErrFeeQuotesNotInit
	}
	f.mu.RLock()
	defer f.mu.RUnlock()
	q, ok := f.quotes[minerName]
	if !ok {
		return nil, ErrMinerNoQuotes
	}
	return q, nil
}

// Fee is a convenience method for quickly getting a fee by type and miner name.
// If the miner has no fees an ErrMinerNoQuotes error will be returned.
// If the feeType cannot be found an ErrFeeTypeNotFound error will be returned.
func (f *FeeQuotes) Fee(minerName string, feeType FeeType) (*Fee, error) {
	if f == nil {
		return nil, ErrFeeQuotesNotInit
	}
	f.mu.RLock()
	defer f.mu.RUnlock()
	m := f.quotes[minerName]
	if m == nil {
		return nil, ErrMinerNoQuotes
	}
	return m.Fee(feeType)
}

// UpdateMinerFees a convenience method to update a fee quote from a FeeQuotes struct directly.
// This will update the miner feeType with the provided fee. Useful after receiving new quotes from mapi.
func (f *FeeQuotes) UpdateMinerFees(minerName string, feeType FeeType, fee *Fee) (*FeeQuote, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if minerName == "" || feeType == "" || fee == nil {
		return nil, ErrEmptyValues
	}
	m := f.quotes[minerName]
	if m == nil {
		return nil, ErrMinerNoQuotes
	}
	return m.AddQuote(feeType, fee), nil
}

// FeeQuote contains a thread safe map of fees for standard and data
// fees as well as an expiry time for a specific miner.
//
// This can be used if you are only dealing with a single miner and know you
// will always be using a single miner.
// FeeQuote will store the fees for a single miner and can be passed to transactions
// to calculate fees when creating change outputs.
//
// If you are dealing with quotes from multiple miners, use the FeeQuotes structure above.
//
// NewFeeQuote() should be called to get a new instance of a FeeQuote.
//
// When expiry expires ie Expired() == true then you should fetch
// new quotes from a MAPI server and call AddQuote with the fee information.
type FeeQuote struct {
	mu         sync.RWMutex
	fees       map[FeeType]*Fee
	expiryTime time.Time
}

// NewFeeQuote will set up and return a new FeeQuotes struct which
// contains default fees when initially setup. You would then pass this
// data structure to a singleton struct via injection for reading.
// If you are only getting quotes from one miner you can use this directly
// instead of using the NewFeeQuotes() method which is for storing multiple miner quotes.
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
func NewFeeQuote() *FeeQuote {
	fq := &FeeQuote{
		fees:       map[FeeType]*Fee{},
		expiryTime: time.Now().UTC(),
		mu:         sync.RWMutex{},
	}
	fq.AddQuote(FeeTypeStandard, defaultStandardFee()).
		AddQuote(FeeTypeData, defaultDataFee())
	return fq
}

// Fee will return a fee by type if found, nil and an error if not.
func (f *FeeQuote) Fee(t FeeType) (*Fee, error) {
	if f == nil {
		return nil, ErrFeeQuoteNotInit
	}
	f.mu.RLock()
	defer f.mu.RUnlock()
	fee, ok := f.fees[t]
	if fee == nil || !ok {
		return nil, ErrFeeTypeNotFound
	}
	return fee, nil
}

// AddQuote will add new set of quotes for a feetype or update an existing
// quote if it already exists.
func (f *FeeQuote) AddQuote(ft FeeType, fee *Fee) *FeeQuote {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.fees[ft] = fee
	return f
}

// Expiry will return the expiry timestamp for the `bt.FeeQuote` in a threadsafe manner.
func (f *FeeQuote) Expiry() time.Time {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.expiryTime
}

// UpdateExpiry will update the expiry time of the quotes, this will be
// used when you fetch a fresh set of quotes from a MAPI server which
// should return an expiration time.
func (f *FeeQuote) UpdateExpiry(exp time.Time) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.expiryTime = exp
}

// Expired will return true if the expiry time is before UTC now, this
// means we need to fetch fresh quotes from a MAPI server.
func (f *FeeQuote) Expired() bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.expiryTime.Before(time.Now().UTC())
}

// MarshalJSON will convert the FeeQuote to a json object
// with the format as shown:
//  {
//	 "data": {
//		 "miningFee": {
//			 "satoshis": 5,
//			 "bytes": 2
//		 },
//		 "relayFee": {
//			 "satoshis": 8,
//			 "bytes": 4
//		 }
//	 },
//	 "standard": {
//		 "miningFee": {
//			 "satoshis": 100,
//			 "bytes": 10
//		 },
//		 "relayFee": {
//			 "satoshis": 10,
//			 "bytes": 5
//		 }
//	 }
//  }
func (f *FeeQuote) MarshalJSON() ([]byte, error) {
	return json.Marshal(f.fees)
}

// UnmarshalJSON will convert a json encoded FeeQuote back into a fee quote type, the expected
// JSON format is shown above in the MarshalJSON function.
// If the fee type supplied is unknown an ErrUnknownFeeType will be returned.
func (f *FeeQuote) UnmarshalJSON(body []byte) error {
	fees := map[FeeType]*Fee{}
	if err := json.Unmarshal(body, &fees); err != nil {
		return err
	}
	for k, v := range fees {
		if k != FeeTypeData && k != FeeTypeStandard {
			return fmt.Errorf("%w '%s'", ErrUnknownFeeType, k)
		}
		v.FeeType = k
	}
	f.fees = fees
	return nil
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
	FeeType   FeeType `json:"-"` // standard || data
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
			Bytes:    100,
		},
		RelayFee: FeeUnit{
			Satoshis: 5,
			Bytes:    100,
		},
	}
}

// defaultDataFee returns the default
// data fees offered by most miners.
func defaultDataFee() *Fee {
	return &Fee{
		FeeType: FeeTypeData,
		MiningFee: FeeUnit{
			Satoshis: 5,
			Bytes:    100,
		},
		RelayFee: FeeUnit{
			Satoshis: 5,
			Bytes:    100,
		},
	}
}
