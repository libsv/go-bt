package bt_test

import (
	"testing"

	"github.com/libsv/go-bt"
	"github.com/stretchr/testify/assert"
)

func TestExtractDataFee(t *testing.T) {
	t.Run("get valid data fee", func(t *testing.T) {
		fees := bt.NewFeeQuote()
		fee, err := fees.Fee(bt.FeeTypeData)
		assert.NotNil(t, fee)
		assert.NoError(t, err)
		assert.Equal(t, bt.FeeTypeData, fee.FeeType)
		assert.Equal(t, 25, fee.MiningFee.Satoshis)
		assert.Equal(t, 100, fee.MiningFee.Bytes)
		assert.Equal(t, 25, fee.RelayFee.Satoshis)
		assert.Equal(t, 100, fee.RelayFee.Bytes)
	})

	t.Run("no data fee found", func(t *testing.T) {
		fees := bt.NewFeeQuote()
		fees.AddQuote(bt.FeeTypeData, nil)
		fee, err := fees.Fee(bt.FeeTypeData)
		assert.Nil(t, fee)
		assert.Error(t, err)
	})
}

func TestExtractStandardFee(t *testing.T) {
	t.Run("get valid standard fee", func(t *testing.T) {
		fees := bt.NewFeeQuote()
		fee, err := fees.Fee(bt.FeeTypeStandard)
		assert.NoError(t, err)
		assert.NotNil(t, fee)
		assert.Equal(t, bt.FeeTypeStandard, fee.FeeType)
		assert.Equal(t, 5, fee.MiningFee.Satoshis)
		assert.Equal(t, 10, fee.MiningFee.Bytes)
		assert.Equal(t, 5, fee.RelayFee.Satoshis)
		assert.Equal(t, 10, fee.RelayFee.Bytes)
	})

	t.Run("no standard fee found", func(t *testing.T) {
		fees := bt.NewFeeQuote()
		fees.AddQuote(bt.FeeTypeStandard, nil)
		fee, err := fees.Fee(bt.FeeTypeStandard)
		assert.Error(t, err)
		assert.Nil(t, fee)
	})
}

func TestDefaultFees(t *testing.T) {
	fees := bt.NewFeeQuote()

	fee, err := fees.Fee( bt.FeeTypeData)
	assert.NoError(t, err)
	assert.NotNil(t, fee)
	assert.Equal(t, bt.FeeTypeData, fee.FeeType)

	fee, err = fees.Fee( bt.FeeTypeStandard)
	assert.NoError(t, err)
	assert.NotNil(t, fee)
	assert.Equal(t, bt.FeeTypeStandard, fee.FeeType)
}
