package bt_test

import (
	"context"
	"testing"

	. "github.com/libsv/go-bk/wif"
	"github.com/libsv/go-bt/v2"
	"github.com/libsv/go-bt/v2/sighash"
	"github.com/stretchr/testify/assert"
)

func TestTx_UnlockInput(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		inputIdx uint32
		shf      sighash.Flag
		unlocker bt.Unlocker
		expHex   string
		expErr   error
	}{
		"standard unlock": {
			inputIdx: 0,
			shf:      sighash.AllForkID,
			unlocker: func() bt.Unlocker {
				var wif *WIF
				wif, err := DecodeWIF("L3MhnEn1pLWcggeYLk9jdkvA2wUK1iWwwrGkBbgQRqv6HPCdRxuw")
				assert.NoError(t, err)

				return &bt.LocalUnlocker{PrivateKey: wif.PrivKey}
			}(),
			expHex: "01000000010b94a1ef0fb352aa2adc54207ce47ba55d5a1c1609afda58fe9520e472299107000000006a473044022049ee0c0f26c00e6a6b3af5990fc8296c66eab3e3e42ab075069b89b1be6fefec02206079e49dd8c9e1117ef06fbe99714d822620b1f0f5d19f32a1128f5d29b7c3c4412102c8803fdd437d902f08e3c2344cb33065c99d7c99982018ff9f7219c3dd352ff0ffffffff01a0083d00000000001976a914af2590a45ae401651fdbdf59a76ad43d1862534088ac00000000",
		},
		"sighash all is used as default": {
			inputIdx: 0,
			unlocker: func() bt.Unlocker {
				var wif *WIF
				wif, err := DecodeWIF("L3MhnEn1pLWcggeYLk9jdkvA2wUK1iWwwrGkBbgQRqv6HPCdRxuw")
				assert.NoError(t, err)

				return &bt.LocalUnlocker{PrivateKey: wif.PrivKey}
			}(),
			expHex: "01000000010b94a1ef0fb352aa2adc54207ce47ba55d5a1c1609afda58fe9520e472299107000000006a473044022049ee0c0f26c00e6a6b3af5990fc8296c66eab3e3e42ab075069b89b1be6fefec02206079e49dd8c9e1117ef06fbe99714d822620b1f0f5d19f32a1128f5d29b7c3c4412102c8803fdd437d902f08e3c2344cb33065c99d7c99982018ff9f7219c3dd352ff0ffffffff01a0083d00000000001976a914af2590a45ae401651fdbdf59a76ad43d1862534088ac00000000",
		},
		"no unlocker errors": {
			inputIdx: 0,
			shf:      sighash.AllForkID,
			expErr:   bt.ErrNoUnlocker,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			tx := bt.NewTx()
			assert.NoError(t, tx.From(
				"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
				0,
				"76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
				4000000,
			))
			assert.NoError(t, tx.ChangeToAddress("mwV3YgnowbJJB3LcyCuqiKpdivvNNFiK7M", bt.NewFeeQuote()))

			err := tx.UnlockInput(context.Background(), bt.UnlockInputParams{
				InputIdx:     test.inputIdx,
				Unlocker:     test.unlocker,
				SigHashFlags: test.shf,
			})
			if test.expErr != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, test.expErr.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expHex, tx.String())
			}
		})
	}
}

func TestTx_UnlockAllInputs(t *testing.T) {
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

		err = tx.ChangeToAddress("mwV3YgnowbJJB3LcyCuqiKpdivvNNFiK7M", bt.NewFeeQuote())
		assert.NoError(t, err)

		var wif *WIF
		wif, err = DecodeWIF("L3MhnEn1pLWcggeYLk9jdkvA2wUK1iWwwrGkBbgQRqv6HPCdRxuw")
		assert.NoError(t, err)
		assert.NotNil(t, wif)

		rawTxBefore := tx.String()

		err = tx.UnlockAllInputs(context.Background(), &bt.LocalUnlockerGetter{PrivateKey: wif.PrivKey})
		assert.NoError(t, err)

		assert.NotEqual(t, rawTxBefore, tx.String())
	})

	t.Run("no input or output", func(t *testing.T) {
		tx := bt.NewTx()
		assert.NotNil(t, tx)

		rawTxBefore := tx.String()

		err := tx.UnlockAllInputs(context.Background(), &bt.LocalUnlockerGetter{PrivateKey: nil})
		assert.NoError(t, err)

		assert.Equal(t, rawTxBefore, tx.String())
	})
}
