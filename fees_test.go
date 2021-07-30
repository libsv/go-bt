package bt_test

import (
	"testing"

	"github.com/libsv/go-bt/v2"
	"github.com/stretchr/testify/assert"
)

func TestExtractDataFee(t *testing.T) {
	t.Run("get valid data fee", func(t *testing.T) {
		fees := []*bt.Fee{bt.DefaultDataFee()}
		fee, err := bt.ExtractDataFee(fees)
		assert.NoError(t, err)
		assert.NotNil(t, fee)
		assert.Equal(t, bt.FeeTypeData, fee.FeeType)
		assert.Equal(t, 25, fee.MiningFee.Satoshis)
		assert.Equal(t, 100, fee.MiningFee.Bytes)
		assert.Equal(t, 25, fee.RelayFee.Satoshis)
		assert.Equal(t, 100, fee.RelayFee.Bytes)
	})

	t.Run("no data fee found", func(t *testing.T) {
		wrongFee := bt.DefaultDataFee()
		wrongFee.FeeType = "unknown"
		fees := []*bt.Fee{wrongFee}
		fee, err := bt.ExtractDataFee(fees)
		assert.Error(t, err)
		assert.Nil(t, fee)
	})
}

func TestExtractStandardFee(t *testing.T) {
	t.Run("get valid standard fee", func(t *testing.T) {
		fees := []*bt.Fee{bt.DefaultStandardFee()}
		fee, err := bt.ExtractStandardFee(fees)
		assert.NoError(t, err)
		assert.NotNil(t, fee)
		assert.Equal(t, bt.FeeTypeStandard, fee.FeeType)
		assert.Equal(t, 5, fee.MiningFee.Satoshis)
		assert.Equal(t, 10, fee.MiningFee.Bytes)
		assert.Equal(t, 5, fee.RelayFee.Satoshis)
		assert.Equal(t, 10, fee.RelayFee.Bytes)
	})

	t.Run("no standard fee found", func(t *testing.T) {
		wrongFee := bt.DefaultStandardFee()
		wrongFee.FeeType = "unknown"
		fees := []*bt.Fee{wrongFee}
		fee, err := bt.ExtractStandardFee(fees)
		assert.Error(t, err)
		assert.Nil(t, fee)
	})
}

func TestDefaultFees(t *testing.T) {
	fees := bt.DefaultFees()
	assert.Equal(t, 2, len(fees))

	fee, err := bt.ExtractDataFee(fees)
	assert.NoError(t, err)
	assert.NotNil(t, fee)
	assert.Equal(t, bt.FeeTypeData, fee.FeeType)

	fee, err = bt.ExtractStandardFee(fees)
	assert.NoError(t, err)
	assert.NotNil(t, fee)
	assert.Equal(t, bt.FeeTypeStandard, fee.FeeType)
}
