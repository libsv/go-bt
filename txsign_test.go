package bt_test

import (
	"context"
	"testing"

	"github.com/libsv/go-bt"
	"github.com/libsv/go-bt/bsvutil"
	"github.com/stretchr/testify/assert"
)

func TestTx_Sign(t *testing.T) {
	// todo: add tests
}

func TestTx_SignAuto(t *testing.T) {
	t.Parallel()

	t.Run("valid tx (basic)", func(t *testing.T) {
		tx := bt.NewTx()
		assert.NotNil(t, tx)

		err := tx.From(
			"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
			0,
			"76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
			4000000)
		assert.NoError(t, err)

		err = tx.ChangeToAddress("mwV3YgnowbJJB3LcyCuqiKpdivvNNFiK7M", bt.DefaultFees())
		assert.NoError(t, err)

		var wif *bsvutil.WIF
		wif, err = bsvutil.DecodeWIF("L3MhnEn1pLWcggeYLk9jdkvA2wUK1iWwwrGkBbgQRqv6HPCdRxuw")
		assert.NoError(t, err)
		assert.NotNil(t, wif)

		rawTxBefore := tx.ToString()

		_, err = tx.SignAuto(context.Background(), &bt.LocalSigner{PrivateKey: wif.PrivKey})
		assert.NoError(t, err)

		assert.NotEqual(t, rawTxBefore, tx.ToString())
	})

	t.Run("no input or output", func(t *testing.T) {
		tx := bt.NewTx()
		assert.NotNil(t, tx)

		rawTxBefore := tx.ToString()

		_, err := tx.SignAuto(context.Background(),&bt.LocalSigner{PrivateKey: nil})
		assert.NoError(t, err)

		assert.Equal(t, rawTxBefore, tx.ToString())
	})

	t.Run("valid tx (wrong wif)", func(t *testing.T) {
		tx := bt.NewTx()
		assert.NotNil(t, tx)

		err := tx.From(
			"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
			0,
			"76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
			4000000)
		assert.NoError(t, err)

		err = tx.ChangeToAddress("mwV3YgnowbJJB3LcyCuqiKpdivvNNFiK7M", bt.DefaultFees())
		assert.NoError(t, err)

		var wif *bsvutil.WIF
		wif, err = bsvutil.DecodeWIF("5KgHn2qiftW5LQgCYFtkbrLYB1FuvisDtacax8NCvumw3UTKdcP")
		assert.NoError(t, err)
		assert.NotNil(t, wif)

		// No signature, wrong wif
		rawTxBefore := tx.ToString()
		_, err = tx.SignAuto(context.Background(),&bt.LocalSigner{PrivateKey: wif.PrivKey})
		assert.NoError(t, err)
		assert.Equal(t, rawTxBefore, tx.ToString())
	})
}
