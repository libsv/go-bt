package bt

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestExtractDataFee(t *testing.T) {
	t.Run("get valid data fee", func(t *testing.T) {
		fees := NewFeeQuote()
		fee, err := fees.Fee(FeeTypeData)
		assert.NotNil(t, fee)
		assert.NoError(t, err)
		assert.Equal(t, FeeTypeData, fee.FeeType)
		assert.Equal(t, 25, fee.MiningFee.Satoshis)
		assert.Equal(t, 100, fee.MiningFee.Bytes)
		assert.Equal(t, 25, fee.RelayFee.Satoshis)
		assert.Equal(t, 100, fee.RelayFee.Bytes)
	})

	t.Run("no data fee found", func(t *testing.T) {
		fees := NewFeeQuote()
		fees.AddQuote(FeeTypeData, nil)
		fee, err := fees.Fee(FeeTypeData)
		assert.Nil(t, fee)
		assert.Error(t, err)
	})
}

func TestExtractStandardFee(t *testing.T) {
	t.Run("get valid standard fee", func(t *testing.T) {
		fees := NewFeeQuote()
		fee, err := fees.Fee(FeeTypeStandard)
		assert.NoError(t, err)
		assert.NotNil(t, fee)
		assert.Equal(t, FeeTypeStandard, fee.FeeType)
		assert.Equal(t, 5, fee.MiningFee.Satoshis)
		assert.Equal(t, 10, fee.MiningFee.Bytes)
		assert.Equal(t, 5, fee.RelayFee.Satoshis)
		assert.Equal(t, 10, fee.RelayFee.Bytes)
	})

	t.Run("no standard fee found", func(t *testing.T) {
		fees := NewFeeQuote()
		fees.AddQuote(FeeTypeStandard, nil)
		fee, err := fees.Fee(FeeTypeStandard)
		assert.Error(t, err)
		assert.Nil(t, fee)
	})
}

func TestDefaultFees(t *testing.T) {
	fees := NewFeeQuote()

	fee, err := fees.Fee(FeeTypeData)
	assert.NoError(t, err)
	assert.NotNil(t, fee)
	assert.Equal(t, FeeTypeData, fee.FeeType)

	fee, err = fees.Fee(FeeTypeStandard)
	assert.NoError(t, err)
	assert.NotNil(t, fee)
	assert.Equal(t, FeeTypeStandard, fee.FeeType)
}

func TestFeeQuotes_New(t *testing.T) {
	fq := NewFeeQuote()
	assert.NotNil(t, fq.fees)
	assert.NotEmpty(t, fq.expiryTime)
}

func TestFeeQuotes_Expired(t *testing.T) {
	// should always be true as setup sets up a time for now.
	fq := NewFeeQuote()
	time.Sleep(1 * time.Millisecond)
	assert.True(t, fq.Expired())
}

func TestFeeQuotes_AddQuote(t *testing.T) {
	std := &Fee{
		FeeType: FeeTypeStandard,
		MiningFee: FeeUnit{
			Satoshis: 1234,
			Bytes:    5,
		},
		RelayFee: FeeUnit{
			Satoshis: 1234,
			Bytes:    2,
		},
	}
	data := &Fee{
		FeeType: FeeTypeData,
		MiningFee: FeeUnit{
			Satoshis: 5678,
			Bytes:    10,
		},
		RelayFee: FeeUnit{
			Satoshis: 5678,
			Bytes:    4,
		},
	}
	// should always be true as setup sets up a time for now.
	fq := NewFeeQuote().
		AddQuote(FeeTypeStandard, std).
		AddQuote(FeeTypeData, data)
	sdFee, _ := fq.Fee(FeeTypeStandard)
	assert.Equal(t, std, sdFee)
	dFee, _ := fq.Fee(FeeTypeData)
	assert.Equal(t, data, dFee)
}

func TestFeeQuotes_Concurrent(t *testing.T) {
	std := &Fee{
		FeeType: FeeTypeStandard,
		MiningFee: FeeUnit{
			Satoshis: 1234,
			Bytes:    5,
		},
		RelayFee: FeeUnit{
			Satoshis: 1234,
			Bytes:    2,
		},
	}
	data := &Fee{
		FeeType: FeeTypeData,
		MiningFee: FeeUnit{
			Satoshis: 5678,
			Bytes:    10,
		},
		RelayFee: FeeUnit{
			Satoshis: 5678,
			Bytes:    4,
		},
	}
	fq := NewFeeQuote()
	wg := sync.WaitGroup{}
	// spin up go routines each reading and writing.
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fq.AddQuote(FeeTypeStandard, std).
				AddQuote(FeeTypeData, data)
			sdFee, _ := fq.Fee(FeeTypeStandard)
			assert.Equal(t, std, sdFee)
			dFee, _ := fq.Fee(FeeTypeData)
			assert.Equal(t, data, dFee)
		}()
	}
	// wait to finish - should not cause race condition
	wg.Wait()
	sdFee, _ := fq.Fee(FeeTypeStandard)
	assert.Equal(t, std, sdFee)
	dFee, _ := fq.Fee(FeeTypeData)
	assert.Equal(t, data, dFee)
}

func TestFeeQuotes_UpdateExpiry(t *testing.T) {
	fq := NewFeeQuote()
	fq.UpdateExpiry(time.Now().Add(1 * time.Minute))
	assert.False(t, fq.Expired())
}
