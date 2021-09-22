package bt_test

import (
	"context"
	"errors"
	"testing"

	. "github.com/libsv/go-bk/wif"
	"github.com/libsv/go-bt/v2"
	"github.com/stretchr/testify/assert"
)

func TestTx_ChangeToAddress(t *testing.T) {
	t.Run("missing address and nil fees", func(t *testing.T) {
		tx := bt.NewTx()
		assert.NotNil(t, tx)
		err := tx.From(
			"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
			0,
			"76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
			4000000)
		assert.NoError(t, err)

		err = tx.ChangeToAddress("", nil)
		assert.Error(t, err)
	})

	t.Run("nil fees, valid address", func(t *testing.T) {
		tx := bt.NewTx()
		assert.NotNil(t, tx)
		err := tx.From(
			"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
			0,
			"76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
			4000000)
		assert.NoError(t, err)

		err = tx.ChangeToAddress("1GHMW7ABrFma2NSwiVe9b9bZxkMB7tuPZi", nil)
		assert.Error(t, err)
	})

	t.Run("valid fees, valid address", func(t *testing.T) {
		tx := bt.NewTx()
		assert.NotNil(t, tx)
		err := tx.From(
			"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
			0,
			"76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
			4000000)
		assert.NoError(t, err)

		err = tx.ChangeToAddress("1GHMW7ABrFma2NSwiVe9b9bZxkMB7tuPZi", bt.NewFeeQuote())
		assert.NoError(t, err)

		assert.Equal(t, 1, tx.OutputCount())
		assert.Equal(t, "76a914a7a1a7fd7d279b57b84e596cbbf82608efdb441a88ac", tx.Outputs[0].LockingScript.String())
	})
}

func TestTx_Change(t *testing.T) {
	t.Parallel()

	t.Run("valid change tx (basic)", func(t *testing.T) {
		expectedTx, err := bt.NewTxFromString("01000000010b94a1ef0fb352aa2adc54207ce47ba55d5a1c1609afda58fe9520e472299107000000006a473044022049ee0c0f26c00e6a6b3af5990fc8296c66eab3e3e42ab075069b89b1be6fefec02206079e49dd8c9e1117ef06fbe99714d822620b1f0f5d19f32a1128f5d29b7c3c4412102c8803fdd437d902f08e3c2344cb33065c99d7c99982018ff9f7219c3dd352ff0ffffffff01a0083d00000000001976a914af2590a45ae401651fdbdf59a76ad43d1862534088ac00000000")
		assert.NoError(t, err)
		assert.NotNil(t, expectedTx)

		tx := bt.NewTx()
		assert.NotNil(t, tx)

		err = tx.From(
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

		err = tx.SignAll(context.Background(), &bt.LocalSignerCreator{PrivateKey: wif.PrivKey})
		assert.NoError(t, err)

		assert.Equal(t, expectedTx.String(), tx.String())
	})

	t.Run("change output is added correctly - fee removed", func(t *testing.T) {

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

		err = tx.SignAll(context.Background(), &bt.LocalSignerCreator{PrivateKey: wif.PrivKey})
		assert.NoError(t, err)

		// Correct fee for the tx
		assert.Equal(t, uint64(3999904), tx.Outputs[0].Satoshis)

		// Correct script hex string
		assert.Equal(t,
			"76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
			tx.OutputIdx(0).LockingScriptHexString(),
		)
	})

	t.Run("determine fees are correct, correct change given", func(t *testing.T) {

		tx := bt.NewTx()
		assert.NotNil(t, tx)

		// utxo
		err := tx.From(
			"b7b0650a7c3a1bd4716369783876348b59f5404784970192cec1996e86950576",
			0,
			"76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
			1000)
		assert.NoError(t, err)

		// pay to
		err = tx.PayToAddress("1C8bzHM8XFBHZ2ZZVvFy2NSoAZbwCXAicL", 500)
		assert.NoError(t, err)

		// add some op return
		err = tx.AddOpReturnPartsOutput([][]byte{[]byte("hi"), []byte("how"), []byte("are"), []byte("you")})
		assert.NoError(t, err)

		err = tx.ChangeToAddress("1D7gaZJo3vPn2Ks3PH694W9P8UVYLNh2jY", bt.NewFeeQuote())
		assert.NoError(t, err)

		var wif *WIF
		wif, err = DecodeWIF("L3MhnEn1pLWcggeYLk9jdkvA2wUK1iWwwrGkBbgQRqv6HPCdRxuw")
		assert.NoError(t, err)
		assert.NotNil(t, wif)

		err = tx.SignAll(context.Background(), &bt.LocalSignerCreator{PrivateKey: wif.PrivKey})
		assert.NoError(t, err)

		assert.Equal(t,
			"0100000001760595866e99c1ce920197844740f5598b34763878696371d41b3a7c0a65b0b7000000006a47304402206b5b0b6546dbaccab4cd9c5698eeab7883f79ddbd4cbc195d4458b48b7dba6460220297a4c4b145e644d23cebdd7593f407e8da9c5bb3c3219767207121d65658ae3412102c8803fdd437d902f08e3c2344cb33065c99d7c99982018ff9f7219c3dd352ff0ffffffff03f4010000000000001976a9147a1980655efbfec416b2b0c663a7b3ac0b6a25d288ac000000000000000011006a02686903686f770361726503796f7577010000000000001976a91484e50b300b009833b297dc671817c79b5459da1d88ac00000000",
			tx.String(),
		)

		feePaid := tx.TotalInputSatoshis() - tx.TotalOutputSatoshis()
		assert.Equal(t, uint64(125), feePaid)

		txSize := len(tx.Bytes())
		assert.Equal(t, 251, txSize)

		feeRate := float64(feePaid) / float64(txSize)
		// note, due to the integer maths uses, this doesn't equal exactly 0.5, the fee is 125.5
		// however, this is rounded down giving us the below final figure.
		// The node will also perform the same deterministic fee cal to arrive at the above 125 sats.
		assert.Equal(t, 0.49800796812749004, feeRate)
	})

	t.Run("spend entire utxo - basic - change address", func(t *testing.T) {
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

		err = tx.SignAll(context.Background(), &bt.LocalSignerCreator{PrivateKey: wif.PrivKey})
		assert.NoError(t, err)

		assert.Equal(t, "01000000010b94a1ef0fb352aa2adc54207ce47ba55d5a1c1609afda58fe9520e472299107000000006a473044022049ee0c0f26c00e6a6b3af5990fc8296c66eab3e3e42ab075069b89b1be6fefec02206079e49dd8c9e1117ef06fbe99714d822620b1f0f5d19f32a1128f5d29b7c3c4412102c8803fdd437d902f08e3c2344cb33065c99d7c99982018ff9f7219c3dd352ff0ffffffff01a0083d00000000001976a914af2590a45ae401651fdbdf59a76ad43d1862534088ac00000000", tx.String())

		assert.Equal(t, uint64(3999904), tx.Outputs[0].Satoshis)
	})

	t.Run("spend entire utxo - multi payouts - expected fee", func(t *testing.T) {
		tx := bt.NewTx()
		assert.NotNil(t, tx)

		err := tx.From(
			"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
			0,
			"76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
			4000000)
		assert.NoError(t, err)

		err = tx.PayToAddress("mxAoAyZFXX6LZBWhoam3vjm6xt9NxPQ15f", 1000000)
		assert.NoError(t, err)

		err = tx.PayToAddress("mxAoAyZFXX6LZBWhoam3vjm6xt9NxPQ15f", 3000000)
		assert.NoError(t, err)

		err = tx.ChangeToAddress("mwV3YgnowbJJB3LcyCuqiKpdivvNNFiK7M", bt.NewFeeQuote())
		assert.NoError(t, err)

		var wif *WIF
		wif, err = DecodeWIF("L3MhnEn1pLWcggeYLk9jdkvA2wUK1iWwwrGkBbgQRqv6HPCdRxuw")
		assert.NoError(t, err)
		assert.NotNil(t, wif)

		err = tx.SignAll(context.Background(), &bt.LocalSignerCreator{PrivateKey: wif.PrivKey})
		assert.NoError(t, err)

		assert.Equal(t, "01000000010b94a1ef0fb352aa2adc54207ce47ba55d5a1c1609afda58fe9520e472299107000000006a47304402206bbb4b23349bdf86e6fbc9067226e9a7b15c977fa530999b39cd0a6d9c83360d02202dd8ffdc610e58b3fc92b44400d99e38c78866765f31acb40d98007a52e7a826412102c8803fdd437d902f08e3c2344cb33065c99d7c99982018ff9f7219c3dd352ff0ffffffff0240420f00000000001976a914b6aa34534d2b11e66b438c7525f819aee01e397c88acc0c62d00000000001976a914b6aa34534d2b11e66b438c7525f819aee01e397c88ac00000000", tx.String())

		assert.Equal(t, uint64(1000000), tx.Outputs[0].Satoshis)
		assert.Equal(t, uint64(3000000), tx.Outputs[1].Satoshis)
	})

	t.Run("spend entire utxo - multi payouts - incorrect fee", func(t *testing.T) {
		tx := bt.NewTx()
		assert.NotNil(t, tx)

		err := tx.From(
			"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
			0,
			"76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
			4000000)
		assert.NoError(t, err)

		err = tx.PayToAddress("mxAoAyZFXX6LZBWhoam3vjm6xt9NxPQ15f", 999995)
		assert.NoError(t, err)

		err = tx.PayToAddress("mxAoAyZFXX6LZBWhoam3vjm6xt9NxPQ15f", 3000000)
		assert.NoError(t, err)

		err = tx.ChangeToAddress("mwV3YgnowbJJB3LcyCuqiKpdivvNNFiK7M", bt.NewFeeQuote())
		assert.NoError(t, err)

		var wif *WIF
		wif, err = DecodeWIF("L3MhnEn1pLWcggeYLk9jdkvA2wUK1iWwwrGkBbgQRqv6HPCdRxuw")
		assert.NoError(t, err)
		assert.NotNil(t, wif)

		err = tx.SignAll(context.Background(), &bt.LocalSignerCreator{PrivateKey: wif.PrivKey})
		assert.NoError(t, err)

		assert.Equal(t, "01000000010b94a1ef0fb352aa2adc54207ce47ba55d5a1c1609afda58fe9520e472299107000000006b483045022100fd07316603e9abf393e695192e8ce1e7f808d2735cc57039109a2210ad32d9a7022000e301e2a988b23ab3872b041df8b6eb0315238e0918944cbaf8b6abdde75cac412102c8803fdd437d902f08e3c2344cb33065c99d7c99982018ff9f7219c3dd352ff0ffffffff023b420f00000000001976a914b6aa34534d2b11e66b438c7525f819aee01e397c88acc0c62d00000000001976a914b6aa34534d2b11e66b438c7525f819aee01e397c88ac00000000", tx.String())

		// todo: expected the pay-to Inputs to change based on the fee :P

		assert.Equal(t, uint64(999995), tx.Outputs[0].Satoshis)
		assert.Equal(t, uint64(3000000), tx.Outputs[1].Satoshis)
	})

	t.Run("multiple Inputs, spend all", func(t *testing.T) {
		tx := bt.NewTx()
		assert.NotNil(t, tx)

		err := tx.From(
			"9e88ca8eec0845e9e864c024bc5e6711e670932c9c7d929f9fccdb2c440ae28e",
			0,
			"76a9147824dec00be2c45dad83c9b5e9f5d7ef05ba3cf988ac",
			5689)
		assert.NoError(t, err)

		err = tx.From(
			"4e25b077d4cbb955b5a215feb53f963cf04688ff1777b9bea097c7ddbdf7ea42",
			0,
			"76a9147824dec00be2c45dad83c9b5e9f5d7ef05ba3cf988ac",
			5689)
		assert.NoError(t, err)

		err = tx.ChangeToAddress("1BxGFoRPSFgYxoAStEncL6HuELqPkV3JVj", bt.NewFeeQuote())
		assert.NoError(t, err)

		var wif *WIF
		wif, err = DecodeWIF("5JXAjNX7cbiWvmkdnj1EnTKPChauttKAJibXLm8tqWtDhXrRbKz")
		assert.NoError(t, err)
		assert.NotNil(t, wif)

		err = tx.SignAll(context.Background(), &bt.LocalSignerCreator{PrivateKey: wif.PrivKey})
		assert.NoError(t, err)

		assert.Equal(t, "01000000028ee20a442cdbcc9f9f927d9c2c9370e611675ebc24c064e8e94508ec8eca889e000000006b483045022100f88298f5a380244dd5b91f70be99394f8e562d2a61976ca8cf2aaeb381ee6e6a0220069243fc951061b624cf96124263b857a65a53400846080b543e4a8c16e097ce4121034aaeabc056f33fd960d1e43fc8a0672723af02f275e54c31381af66a334634caffffffff42eaf7bdddc797a0beb97717ff8846f03c963fb5fe15a2b555b9cbd477b0254e000000006b483045022100afa7a986e6e0faf725a9779fe8e61fd19b5973544dc7707fd758cdd45912332a0220760fe07fc8610d867be5281f29778e3cd1a18a6eef74470d0f1a4ede95c848924121034aaeabc056f33fd960d1e43fc8a0672723af02f275e54c31381af66a334634caffffffff01c82b0000000000001976a9147824dec00be2c45dad83c9b5e9f5d7ef05ba3cf988ac00000000", tx.String())
	})
}

func TestTx_ChangeToOutput(t *testing.T) {
	tests := map[string]struct {
		tx              *bt.Tx
		index           uint
		fees            *bt.FeeQuote
		expOutputTotal  uint64
		expChangeOutput uint64
		err             error
	}{
		"no change to add should return no change output": {
			tx: func() *bt.Tx {
				tx := bt.NewTx()
				assert.NoError(t, tx.From(
					"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
					0,
					"76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
					1000))
				assert.NoError(t, tx.PayToAddress("mxAoAyZFXX6LZBWhoam3vjm6xt9NxPQ15f", 1000))
				return tx
			}(),
			index:           0,
			fees:            bt.NewFeeQuote(),
			expOutputTotal:  1000,
			expChangeOutput: 1000,
			err:             nil,
		}, "change to add should add change to output": {
			tx: func() *bt.Tx {
				tx := bt.NewTx()
				assert.NoError(t, tx.From(
					"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
					0,
					"76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
					1000))
				assert.NoError(t, tx.PayToAddress("mxAoAyZFXX6LZBWhoam3vjm6xt9NxPQ15f", 500))
				return tx
			}(),
			index:           0,
			fees:            bt.NewFeeQuote(),
			expOutputTotal:  904,
			expChangeOutput: 904,
			err:             nil,
		}, "change to add should add change to specified output": {
			tx: func() *bt.Tx {
				tx := bt.NewTx()
				assert.NoError(t, tx.From(
					"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
					0,
					"76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
					2500))
				assert.NoError(t, tx.PayToAddress("mxAoAyZFXX6LZBWhoam3vjm6xt9NxPQ15f", 500))
				assert.NoError(t, tx.PayToAddress("mxAoAyZFXX6LZBWhoam3vjm6xt9NxPQ15f", 500))
				assert.NoError(t, tx.PayToAddress("mxAoAyZFXX6LZBWhoam3vjm6xt9NxPQ15f", 500))
				assert.NoError(t, tx.PayToAddress("mxAoAyZFXX6LZBWhoam3vjm6xt9NxPQ15f", 500))
				return tx
			}(),
			index:           3,
			fees:            bt.NewFeeQuote(),
			expOutputTotal:  2353,
			expChangeOutput: 853,
			err:             nil,
		}, "index out of range should return error": {
			tx: func() *bt.Tx {
				tx := bt.NewTx()
				assert.NoError(t, tx.From(
					"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
					0,
					"76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
					1000))
				assert.NoError(t, tx.PayToAddress("mxAoAyZFXX6LZBWhoam3vjm6xt9NxPQ15f", 500))
				return tx
			}(),
			index: 1,
			fees:  bt.NewFeeQuote(),
			err:   errors.New("index is greater than number of Inputs in transaction"),
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			err := test.tx.ChangeToExistingOutput(test.index, test.fees)
			if test.err != nil {
				assert.Error(t, err)
				assert.Equal(t, test.err, err)
				return
			}
			assert.Equal(t, test.expOutputTotal, test.tx.TotalOutputSatoshis())
			assert.Equal(t, test.expChangeOutput, test.tx.Outputs[test.index].Satoshis)
		})
	}
}
