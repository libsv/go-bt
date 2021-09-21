package bt_test

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
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
		err := tx.From(&bt.UTXO{
			"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
			0,
			"76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
			4000000,
		})
		assert.NoError(t, err)
		assert.Equal(t, 1, tx.InputCount())
	})
}

func TestTx_From(t *testing.T) {
	t.Run("invalid locking script (hex decode failed)", func(t *testing.T) {
		tx := bt.NewTx()
		assert.NotNil(t, tx)
		err := tx.From(&bt.UTXO{
			"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
			0,
			"0",
			4000000,
		})
		assert.Error(t, err)

		err = tx.From(&bt.UTXO{
			"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
			0,
			"76a914af2590a45ae4016",
			4000000,
		})
		assert.Error(t, err)
	})

	t.Run("valid script and tx", func(t *testing.T) {
		tx := bt.NewTx()
		assert.NotNil(t, tx)
		err := tx.From(&bt.UTXO{
			"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
			0,
			"76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
			4000000,
		})
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

func TestTx_FromUTXOs(t *testing.T) {
	tests := map[string]struct {
		tx                      *bt.Tx
		utxos                   []*bt.UTXO
		utxoGetterFuncOverrider func([]*bt.UTXO) bt.UTXOGetterFunc
		expTotalInputs          int
		expErr                  error
	}{
		"tx with exact inputs and surplus inputs is covered": {
			tx: func() *bt.Tx {
				tx := bt.NewTx()
				assert.NoError(t, tx.AddP2PKHOutputFromAddress("mtestD3vRB7AoYWK2n6kLdZmAMLbLhDsLr", 1500))
				return tx
			}(),
			utxos: []*bt.UTXO{{
				"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
				0,
				"76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
				1000,
			}, {
				"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
				0,
				"76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
				1000,
			}},
			expTotalInputs: 2,
		},
		"tx with extra inputs and surplus inputs is covered with all inputs": {
			tx: func() *bt.Tx {
				tx := bt.NewTx()
				assert.NoError(t, tx.AddP2PKHOutputFromAddress("mtestD3vRB7AoYWK2n6kLdZmAMLbLhDsLr", 1500))
				return tx
			}(),
			utxos: []*bt.UTXO{{
				"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
				0,
				"76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
				1000,
			}, {
				"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
				0,
				"76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
				1000,
			}, {
				"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
				0,
				"76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
				1000,
			}},
			expTotalInputs: 3,
		},
		"tx with extra inputs and surplus inputs that returns correct amount is covered with minimum needed inputs": {
			tx: func() *bt.Tx {
				tx := bt.NewTx()
				assert.NoError(t, tx.AddP2PKHOutputFromAddress("mtestD3vRB7AoYWK2n6kLdZmAMLbLhDsLr", 1500))
				return tx
			}(),
			utxoGetterFuncOverrider: func(utxos []*bt.UTXO) bt.UTXOGetterFunc {
				return func(ctx context.Context, satoshis uint64) ([]*bt.UTXO, error) {
					return utxos[:2], nil
				}
			},
			utxos: []*bt.UTXO{{
				"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
				0,
				"76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
				1000,
			}, {
				"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
				0,
				"76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
				1000,
			}, {
				"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
				0,
				"76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
				1000,
			}},
			expTotalInputs: 2,
		},
		"tx with exact input satshis is covered": {
			tx: func() *bt.Tx {
				tx := bt.NewTx()
				assert.NoError(t, tx.AddP2PKHOutputFromAddress("mtestD3vRB7AoYWK2n6kLdZmAMLbLhDsLr", 1500))
				return tx
			}(),
			utxos: []*bt.UTXO{{
				"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
				0,
				"76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
				1000,
			}, {
				"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
				0,
				"76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
				1000,
			}},
			expTotalInputs: 2,
		},
		"tx with large amount of satoshis is covered with all inputs": {
			tx: func() *bt.Tx {
				tx := bt.NewTx()
				assert.NoError(t, tx.AddP2PKHOutputFromAddress("mtestD3vRB7AoYWK2n6kLdZmAMLbLhDsLr", 5000))
				return tx
			}(),
			utxos: []*bt.UTXO{{
				"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
				0,
				"76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
				500,
			}, {
				"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
				0,
				"76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
				670,
			}, {
				"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
				0,
				"76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
				700,
			}, {
				"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
				0,
				"76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
				1000,
			}, {
				"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
				0,
				"76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
				1000,
			}, {
				"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
				0,
				"76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
				1000,
			}, {
				"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
				0,
				"76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
				1000,
			}, {
				"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
				0,
				"76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
				650,
			}},
			expTotalInputs: 8,
		},
		"tx with large amount of satoshis is covered with needed inputs": {
			tx: func() *bt.Tx {
				tx := bt.NewTx()
				assert.NoError(t, tx.AddP2PKHOutputFromAddress("mtestD3vRB7AoYWK2n6kLdZmAMLbLhDsLr", 5000))
				return tx
			}(),
			utxoGetterFuncOverrider: func(utxos []*bt.UTXO) bt.UTXOGetterFunc {
				utxosCopy := make([]*bt.UTXO, len(utxos))
				copy(utxosCopy, utxos)
				return func(ctx context.Context, sat uint64) ([]*bt.UTXO, error) {
					defer func() { utxosCopy = utxosCopy[1:] }()
					return utxosCopy[:1], nil
				}
			},
			utxos: []*bt.UTXO{{
				"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
				0,
				"76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
				500,
			}, {
				"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
				0,
				"76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
				670,
			}, {
				"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
				0,
				"76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
				700,
			}, {
				"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
				0,
				"76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
				1000,
			}, {
				"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
				0,
				"76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
				1000,
			}, {
				"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
				0,
				"76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
				1000,
			}, {
				"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
				0,
				"76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
				1000,
			}, {
				"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
				0,
				"76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
				650,
			}},
			expTotalInputs: 7,
		},
		"getter with no inputs error": {
			tx: func() *bt.Tx {
				tx := bt.NewTx()
				assert.NoError(t, tx.AddP2PKHOutputFromAddress("mtestD3vRB7AoYWK2n6kLdZmAMLbLhDsLr", 1500))
				return tx
			}(),
			utxos:  []*bt.UTXO{},
			expErr: errors.New("insufficient inputs provided"),
		},
		"getter with insufficient inputs errors": {
			tx: func() *bt.Tx {
				tx := bt.NewTx()
				assert.NoError(t, tx.AddP2PKHOutputFromAddress("mtestD3vRB7AoYWK2n6kLdZmAMLbLhDsLr", 25400))
				return tx
			}(),
			utxos: []*bt.UTXO{{
				"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
				0,
				"76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
				500,
			}, {
				"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
				0,
				"76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
				670,
			}, {
				"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
				0,
				"76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
				700,
			}, {
				"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
				0,
				"76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
				1000,
			}, {
				"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
				0,
				"76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
				1000,
			}, {
				"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
				0,
				"76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
				1000,
			}, {
				"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
				0,
				"76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
				1000,
			}, {
				"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
				0,
				"76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
				650,
			}},
			expErr: errors.New("insufficient inputs provided"),
		},
		"error is returned to the user": {
			tx: func() *bt.Tx {
				tx := bt.NewTx()
				assert.NoError(t, tx.AddP2PKHOutputFromAddress("mtestD3vRB7AoYWK2n6kLdZmAMLbLhDsLr", 100))
				return tx
			}(),
			utxoGetterFuncOverrider: func([]*bt.UTXO) bt.UTXOGetterFunc {
				return func(context.Context, uint64) ([]*bt.UTXO, error) {
					return nil, errors.New("custom error")
				}
			},
			expErr: errors.New("custom error"),
		},
		"tx with large amount of satoshis is covered, with multiple iterations": {
			tx: func() *bt.Tx {
				tx := bt.NewTx()
				assert.NoError(t, tx.AddP2PKHOutputFromAddress("mtestD3vRB7AoYWK2n6kLdZmAMLbLhDsLr", 5000))
				return tx
			}(),
			utxos: []*bt.UTXO{{
				"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
				0,
				"76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
				500,
			}, {
				"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
				0,
				"76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
				670,
			}, {
				"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
				0,
				"76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
				700,
			}, {
				"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
				0,
				"76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
				1000,
			}, {
				"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
				0,
				"76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
				1000,
			}, {
				"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
				0,
				"76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
				1000,
			}, {
				"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
				0,
				"76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
				1000,
			}, {
				"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
				0,
				"76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
				650,
			}},
			utxoGetterFuncOverrider: func(utxos []*bt.UTXO) bt.UTXOGetterFunc {
				idx := 0
				return func(context.Context, uint64) ([]*bt.UTXO, error) {
					defer func() { idx++ }()
					return utxos[idx : idx+1], nil
				}
			},
			expTotalInputs: 7,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			iptFn := func() bt.UTXOGetterFunc {
				idx := 0
				return func(ctx context.Context, deficit uint64) ([]*bt.UTXO, error) {
					if idx == len(test.utxos) {
						return nil, bt.ErrNoUTXO
					}
					defer func() { idx += len(test.utxos) }()
					return test.utxos, nil
				}
			}()
			if test.utxoGetterFuncOverrider != nil {
				iptFn = test.utxoGetterFuncOverrider(test.utxos)
			}

			err := test.tx.FromUTXOs(context.Background(), bt.NewFeeQuote(), iptFn)
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

func TestTx_FromUTXOs_Deficit(t *testing.T) {
	tests := map[string]struct {
		utxos       []*bt.UTXO
		expDeficits []uint64
		iteration   int
		tx          *bt.Tx
	}{
		"1 output worth 5000, 3 utxos worth 6000": {
			tx: func() *bt.Tx {
				tx := bt.NewTx()
				assert.NoError(t, tx.AddP2PKHOutputFromAddress("mtestD3vRB7AoYWK2n6kLdZmAMLbLhDsLr", 5000))

				return tx
			}(),
			utxos: []*bt.UTXO{{
				"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
				0,
				"76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
				2000,
			}, {
				"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
				0,
				"76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
				2000,
			}, {
				"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
				0,
				"76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
				2000,
			}},
			iteration:   1,
			expDeficits: []uint64{5022, 3096, 1170},
		},
		"1 output worth 5000, 3 utxos worth 6000, iterations of 2": {
			tx: func() *bt.Tx {
				tx := bt.NewTx()
				assert.NoError(t, tx.AddP2PKHOutputFromAddress("mtestD3vRB7AoYWK2n6kLdZmAMLbLhDsLr", 5000))

				return tx
			}(),
			utxos: []*bt.UTXO{{
				"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
				0,
				"76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
				2000,
			}, {
				"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
				0,
				"76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
				2000,
			}, {
				"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
				0,
				"76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
				2000,
			}},
			iteration:   2,
			expDeficits: []uint64{5022, 1170},
		},
		"5 outputs worth 35000, 12 utxos worth 37000": {
			tx: func() *bt.Tx {
				tx := bt.NewTx()
				assert.NoError(t, tx.AddP2PKHOutputFromAddress("mtestD3vRB7AoYWK2n6kLdZmAMLbLhDsLr", 5000))
				assert.NoError(t, tx.AddP2PKHOutputFromAddress("mtestD3vRB7AoYWK2n6kLdZmAMLbLhDsLr", 10000))
				assert.NoError(t, tx.AddP2PKHOutputFromAddress("mtestD3vRB7AoYWK2n6kLdZmAMLbLhDsLr", 7000))
				assert.NoError(t, tx.AddP2PKHOutputFromAddress("mtestD3vRB7AoYWK2n6kLdZmAMLbLhDsLr", 3000))
				assert.NoError(t, tx.AddP2PKHOutputFromAddress("mtestD3vRB7AoYWK2n6kLdZmAMLbLhDsLr", 10000))

				return tx
			}(),
			utxos: []*bt.UTXO{{
				"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
				0,
				"76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
				2000,
			}, {
				"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
				0,
				"76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
				2000,
			}, {
				"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
				0,
				"76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
				2000,
			}, {
				"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
				0,
				"76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
				2000,
			}, {
				"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
				0,
				"76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
				4000,
			}, {
				"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
				0,
				"76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
				2000,
			}, {
				"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
				0,
				"76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
				6000,
			}, {
				"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
				0,
				"76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
				4000,
			}, {
				"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
				0,
				"76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
				2000,
			}, {
				"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
				0,
				"76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
				8000,
			}, {
				"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
				0,
				"76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
				3000,
			}},
			iteration:   1,
			expDeficits: []uint64{35090, 33164, 31238, 29312, 27386, 23460, 21534, 15608, 11682, 9756, 1830},
		},
		"5 outputs worth 35000, 12 utxos worth 37000, iteration of 3": {
			tx: func() *bt.Tx {
				tx := bt.NewTx()
				assert.NoError(t, tx.AddP2PKHOutputFromAddress("mtestD3vRB7AoYWK2n6kLdZmAMLbLhDsLr", 5000))
				assert.NoError(t, tx.AddP2PKHOutputFromAddress("mtestD3vRB7AoYWK2n6kLdZmAMLbLhDsLr", 10000))
				assert.NoError(t, tx.AddP2PKHOutputFromAddress("mtestD3vRB7AoYWK2n6kLdZmAMLbLhDsLr", 7000))
				assert.NoError(t, tx.AddP2PKHOutputFromAddress("mtestD3vRB7AoYWK2n6kLdZmAMLbLhDsLr", 3000))
				assert.NoError(t, tx.AddP2PKHOutputFromAddress("mtestD3vRB7AoYWK2n6kLdZmAMLbLhDsLr", 10000))

				return tx
			}(),
			utxos: []*bt.UTXO{{
				"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
				0,
				"76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
				2000,
			}, {
				"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
				0,
				"76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
				2000,
			}, {
				"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
				0,
				"76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
				2000,
			}, {
				"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
				0,
				"76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
				2000,
			}, {
				"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
				0,
				"76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
				4000,
			}, {
				"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
				0,
				"76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
				2000,
			}, {
				"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
				0,
				"76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
				6000,
			}, {
				"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
				0,
				"76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
				4000,
			}, {
				"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
				0,
				"76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
				2000,
			}, {
				"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
				0,
				"76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
				8000,
			}, {
				"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
				0,
				"76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
				3000,
			}},
			iteration:   3,
			expDeficits: []uint64{35090, 29312, 21534, 9756},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			deficits := make([]uint64, 0)
			test.tx.FromUTXOs(context.Background(), bt.NewFeeQuote(), func(ctx context.Context, deficit uint64) ([]*bt.UTXO, error) {
				if len(test.utxos) == 0 {
					return nil, bt.ErrNoUTXO
				}
				step := int(math.Min(float64(test.iteration), float64(len(test.utxos))))
				defer func() {
					test.utxos = test.utxos[step:]
				}()

				deficits = append(deficits, deficit)
				fmt.Println(test.utxos[:step])
				return test.utxos[:step], nil
			})

			assert.Equal(t, test.expDeficits, deficits)
		})
	}
}
