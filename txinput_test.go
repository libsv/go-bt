package bt_test

import (
	"context"
	"encoding/hex"
	"errors"
	"testing"

	"github.com/libsv/go-bt/v2"
	"github.com/stretchr/testify/assert"
)

func TestAddInputFromTx(t *testing.T) {
	pubkey1, _ := hex.DecodeString("0280f642908697e8068c2e921bd998d6c2b90553064656f91b9cb9e98f443aac30")
	pubkey2, _ := hex.DecodeString("02434dc3db4281c0895d7a126bb266e7648caca7d0e2e487bc41f954722d4ee397")

	prvTx := bt.NewTx()
	err := prvTx.AddP2PKHOutputFromPubKeyBytes(pubkey1, uint64(100000))
	assert.NoError(t, err)
	err = prvTx.AddP2PKHOutputFromPubKeyBytes(pubkey1, uint64(100000))
	assert.NoError(t, err)
	err = prvTx.AddP2PKHOutputFromPubKeyBytes(pubkey2, uint64(100000))
	assert.NoError(t, err)

	newTx := bt.NewTx()
	err = newTx.AddP2PKHInputsFromTx(prvTx, pubkey1)
	assert.NoError(t, err)
	assert.Equal(t, newTx.InputCount(), 2) // only 2 utxos added
	assert.Equal(t, newTx.TotalInputSatoshis(), uint64(200000))
}

func TestTx_InputCount(t *testing.T) {
	t.Run("get input count", func(t *testing.T) {
		tx := bt.NewTx()
		assert.NotNil(t, tx)
		err := tx.From(
			"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
			0,
			"76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
			4000000)
		assert.NoError(t, err)
		assert.Equal(t, 1, tx.InputCount())
	})
}

func TestTx_From(t *testing.T) {
	t.Run("invalid locking script (hex decode failed)", func(t *testing.T) {
		tx := bt.NewTx()
		assert.NotNil(t, tx)
		err := tx.From(
			"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
			0,
			"0",
			4000000)
		assert.Error(t, err)

		err = tx.From(
			"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
			0,
			"76a914af2590a45ae4016",
			4000000)
		assert.Error(t, err)
	})

	t.Run("valid script and tx", func(t *testing.T) {
		tx := bt.NewTx()
		assert.NotNil(t, tx)
		err := tx.From(
			"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
			0,
			"76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
			4000000)
		assert.NoError(t, err)

		inputs := tx.Inputs
		assert.Equal(t, 1, len(inputs))
		assert.Equal(t, "07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b", hex.EncodeToString(inputs[0].PreviousTxID()))
		assert.Equal(t, uint32(0), inputs[0].PreviousTxOutIndex)
		assert.Equal(t, uint64(4000000), inputs[0].PreviousTxSatoshis)
		assert.Equal(t, bt.DefaultSequenceNumber, inputs[0].SequenceNumber)
		assert.Equal(t, "76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac", inputs[0].PreviousTxScript.String())
	})
}

func TestTx_AutoFund(t *testing.T) {
	tests := map[string]struct {
		tx             *bt.Tx
		funds          []bt.Fund
		expTotalInputs int
		expErr         error
	}{
		"tx with exact inputs and surplus funds is covered": {
			tx: func() *bt.Tx {
				tx := bt.NewTx()
				assert.NoError(t, tx.AddP2PKHOutputFromAddress("mtestD3vRB7AoYWK2n6kLdZmAMLbLhDsLr", 1500))
				return tx
			}(),
			funds: []bt.Fund{{
				TxID:          "07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
				OutIndex:      0,
				LockingScript: "76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
				Satoshis:      1000,
			}, {
				TxID:          "07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
				OutIndex:      0,
				LockingScript: "76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
				Satoshis:      1000,
			}},
			expTotalInputs: 2,
		},
		"tx with extra inputs and surplus funds is covered with minimum needed inputs": {
			tx: func() *bt.Tx {
				tx := bt.NewTx()
				assert.NoError(t, tx.AddP2PKHOutputFromAddress("mtestD3vRB7AoYWK2n6kLdZmAMLbLhDsLr", 1500))
				return tx
			}(),
			funds: []bt.Fund{{
				TxID:          "07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
				OutIndex:      0,
				LockingScript: "76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
				Satoshis:      1000,
			}, {
				TxID:          "07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
				OutIndex:      0,
				LockingScript: "76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
				Satoshis:      1000,
			}, {
				TxID:          "07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
				OutIndex:      0,
				LockingScript: "76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
				Satoshis:      1000,
			}},
			expTotalInputs: 2,
		},
		"tx with exact input satshis is covered": {
			tx: func() *bt.Tx {
				tx := bt.NewTx()
				assert.NoError(t, tx.AddP2PKHOutputFromAddress("mtestD3vRB7AoYWK2n6kLdZmAMLbLhDsLr", 1500))
				return tx
			}(),
			funds: []bt.Fund{{
				TxID:          "07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
				OutIndex:      0,
				LockingScript: "76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
				Satoshis:      1000,
			}, {
				TxID:          "07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
				OutIndex:      0,
				LockingScript: "76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
				Satoshis:      670,
			}},
			expTotalInputs: 2,
		},
		"tx with large amount of satoshis is covered": {
			tx: func() *bt.Tx {
				tx := bt.NewTx()
				assert.NoError(t, tx.AddP2PKHOutputFromAddress("mtestD3vRB7AoYWK2n6kLdZmAMLbLhDsLr", 5000))
				return tx
			}(),
			funds: []bt.Fund{{
				TxID:          "07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
				OutIndex:      0,
				LockingScript: "76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
				Satoshis:      500,
			}, {
				TxID:          "07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
				OutIndex:      1,
				LockingScript: "76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
				Satoshis:      670,
			}, {
				TxID:          "07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
				OutIndex:      2,
				LockingScript: "76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
				Satoshis:      700,
			}, {
				TxID:          "07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
				OutIndex:      3,
				LockingScript: "76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
				Satoshis:      1000,
			}, {
				TxID:          "07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
				OutIndex:      4,
				LockingScript: "76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
				Satoshis:      1000,
			}, {
				TxID:          "07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
				OutIndex:      5,
				LockingScript: "76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
				Satoshis:      1000,
			}, {
				TxID:          "07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
				OutIndex:      6,
				LockingScript: "76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
				Satoshis:      1000,
			}, {
				TxID:          "07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
				OutIndex:      7,
				LockingScript: "76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
				Satoshis:      650,
			}},
			expTotalInputs: 7,
		},
		"iterator with no funds error": {
			tx: func() *bt.Tx {
				tx := bt.NewTx()
				assert.NoError(t, tx.AddP2PKHOutputFromAddress("mtestD3vRB7AoYWK2n6kLdZmAMLbLhDsLr", 1500))
				return tx
			}(),
			funds:  []bt.Fund{},
			expErr: errors.New("insufficient funds from iterator"),
		},
		"iterator with insufficient funds errors": {
			tx: func() *bt.Tx {
				tx := bt.NewTx()
				assert.NoError(t, tx.AddP2PKHOutputFromAddress("mtestD3vRB7AoYWK2n6kLdZmAMLbLhDsLr", 25400))
				return tx
			}(),
			funds: []bt.Fund{{
				TxID:          "07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
				OutIndex:      0,
				LockingScript: "76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
				Satoshis:      500,
			}, {
				TxID:          "07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
				OutIndex:      1,
				LockingScript: "76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
				Satoshis:      670,
			}, {
				TxID:          "07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
				OutIndex:      2,
				LockingScript: "76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
				Satoshis:      700,
			}, {
				TxID:          "07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
				OutIndex:      3,
				LockingScript: "76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
				Satoshis:      1000,
			}, {
				TxID:          "07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
				OutIndex:      4,
				LockingScript: "76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
				Satoshis:      1000,
			}, {
				TxID:          "07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
				OutIndex:      5,
				LockingScript: "76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
				Satoshis:      1000,
			}, {
				TxID:          "07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
				OutIndex:      6,
				LockingScript: "76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
				Satoshis:      1000,
			}, {
				TxID:          "07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
				OutIndex:      7,
				LockingScript: "76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
				Satoshis:      650,
			}},
			expErr: errors.New("insufficient funds from iterator"),
		},
		"invalid tx script errors": {
			tx: func() *bt.Tx {
				tx := bt.NewTx()
				assert.NoError(t, tx.AddP2PKHOutputFromAddress("mtestD3vRB7AoYWK2n6kLdZmAMLbLhDsLr", 1500))
				return tx
			}(),
			funds: []bt.Fund{{
				TxID:          "07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
				OutIndex:      7,
				LockingScript: "ohhellotherea45ae401651fdbdf59a76ad43d1862534088ac",
				Satoshis:      650,
			}},
			expErr: errors.New("encoding/hex: invalid byte: U+006F 'o'"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			err := test.tx.FromFunds(context.Background(), bt.NewFeeQuote(), func() bt.FundGetterFunc {
				idx := 0
				return func(ctx context.Context) (*bt.Fund, error) {
					if idx == len(test.funds) {
						return nil, bt.ErrNoFund
					}
					defer func() { idx++ }()
					return &test.funds[idx], nil
				}
			}())

			if test.expErr != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, test.expErr.Error())
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, test.expTotalInputs, test.tx.InputCount())
		})
	}
}
