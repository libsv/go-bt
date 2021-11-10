package bt_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/libsv/go-bk/wif"
	"github.com/libsv/go-bt/v2"
	"github.com/libsv/go-bt/v2/bscript"
	"github.com/stretchr/testify/assert"
)

func TestTxJSON_Node_JSON(t *testing.T) {
	tests := map[string]struct {
		tx  *bt.Tx
		err error
	}{
		"node standard tx should marshal and unmarshall correctly": {
			tx: func() *bt.Tx {
				tx := bt.NewTx()
				assert.NoError(t, tx.From(
					"3c8edde27cb9a9132c22038dac4391496be9db16fd21351565cc1006966fdad5",
					0,
					"76a914eb0bd5edba389198e73f8efabddfc61666969ff788ac",
					2000000,
				))
				assert.NoError(t, tx.PayToAddress("n2wmGVP89x3DsLNqk3NvctfQy9m9pvt7mk", 1000))
				var w *wif.WIF
				w, err := wif.DecodeWIF("KznvCNc6Yf4iztSThoMH6oHWzH9EgjfodKxmeuUGPq5DEX5maspS")
				assert.NoError(t, err)
				assert.NotNil(t, w)

				err = tx.UnlockAll(context.Background(), &bt.LocalUnlockerGetter{PrivateKey: w.PrivKey})
				assert.NoError(t, err)
				return tx
			}(),
		}, "node data tx should marshall correctly": {
			tx: func() *bt.Tx {
				tx := bt.NewTx()
				assert.NoError(t, tx.From(
					"3c8edde27cb9a9132c22038dac4391496be9db16fd21351565cc1006966fdad5",
					0,
					"76a914eb0bd5edba389198e73f8efabddfc61666969ff788ac",
					2000000,
				))
				assert.NoError(t, tx.PayToAddress("n2wmGVP89x3DsLNqk3NvctfQy9m9pvt7mk", 1000))
				var w *wif.WIF
				w, err := wif.DecodeWIF("KznvCNc6Yf4iztSThoMH6oHWzH9EgjfodKxmeuUGPq5DEX5maspS")
				assert.NoError(t, err)
				assert.NotNil(t, w)
				s := &bscript.Script{}
				assert.NoError(t, s.AppendPushDataString("test"))
				tx.AddOutput(&bt.Output{
					LockingScript: s,
				})
				err = tx.UnlockAll(context.Background(), &bt.LocalUnlockerGetter{PrivateKey: w.PrivKey})
				assert.NoError(t, err)
				return tx
			}(),
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			bb, err := json.Marshal(test.tx.NodeJSON())
			assert.NoError(t, err)
			if err != nil {
				return
			}
			tx := &bt.Tx{}
			assert.NoError(t, json.Unmarshal(bb, tx.NodeJSON()))
			assert.Equal(t, test.tx.String(), tx.String())
		})
	}
}

func TestTxJSON_Node_MarshallJSON(t *testing.T) {
	tests := map[string]struct {
		tx      *bt.Tx
		expJSON string
	}{
		"transaction with 1 input 1 p2pksh output 1 data output should create valid json": {
			tx: func() *bt.Tx {
				tx, err := bt.NewTxFromString("0100000001abad53d72f342dd3f338e5e3346b492440f8ea821f8b8800e318f461cc5ea5a2010000006a4730440220042edc1302c5463e8397120a56b28ea381c8f7f6d9bdc1fee5ebca00c84a76e2022077069bbdb7ed701c4977b7db0aba80d41d4e693112256660bb5d674599e390cf41210294639d6e4249ea381c2e077e95c78fc97afe47a52eb24e1b1595cd3fdd0afdf8ffffffff02000000000000000008006a0548656c6c6f7f030000000000001976a914b85524abf8202a961b847a3bd0bc89d3d4d41cc588ac00000000")
				assert.NoError(t, err)
				return tx
			}(),
			expJSON: `{
	"version": 1,
	"locktime": 0,
	"txid": "aec245f27b7640c8b1865045107731bfb848115c573f7da38166074b1c9e475d",
	"hash": "aec245f27b7640c8b1865045107731bfb848115c573f7da38166074b1c9e475d",
	"size": 208,
	"hex": "0100000001abad53d72f342dd3f338e5e3346b492440f8ea821f8b8800e318f461cc5ea5a2010000006a4730440220042edc1302c5463e8397120a56b28ea381c8f7f6d9bdc1fee5ebca00c84a76e2022077069bbdb7ed701c4977b7db0aba80d41d4e693112256660bb5d674599e390cf41210294639d6e4249ea381c2e077e95c78fc97afe47a52eb24e1b1595cd3fdd0afdf8ffffffff02000000000000000008006a0548656c6c6f7f030000000000001976a914b85524abf8202a961b847a3bd0bc89d3d4d41cc588ac00000000",
	"vin": [
		{
			"scriptSig": {
				"asm": "30440220042edc1302c5463e8397120a56b28ea381c8f7f6d9bdc1fee5ebca00c84a76e2022077069bbdb7ed701c4977b7db0aba80d41d4e693112256660bb5d674599e390cf41 0294639d6e4249ea381c2e077e95c78fc97afe47a52eb24e1b1595cd3fdd0afdf8",
				"hex": "4730440220042edc1302c5463e8397120a56b28ea381c8f7f6d9bdc1fee5ebca00c84a76e2022077069bbdb7ed701c4977b7db0aba80d41d4e693112256660bb5d674599e390cf41210294639d6e4249ea381c2e077e95c78fc97afe47a52eb24e1b1595cd3fdd0afdf8"
			},
			"txid": "a2a55ecc61f418e300888b1f82eaf84024496b34e3e538f3d32d342fd753adab",
			"vout": 1,
			"sequence": 4294967295
		}
	],
	"vout": [
		{
			"value": 0,
			"n": 0,
			"scriptPubKey": {
				"asm": "OP_FALSE OP_RETURN 48656c6c6f",
				"hex": "006a0548656c6c6f",
				"type": "nulldata"
			}
		},
		{
			"value": 0.00000895,
			"n": 1,
			"scriptPubKey": {
				"asm": "OP_DUP OP_HASH160 b85524abf8202a961b847a3bd0bc89d3d4d41cc5 OP_EQUALVERIFY OP_CHECKSIG",
				"hex": "76a914b85524abf8202a961b847a3bd0bc89d3d4d41cc588ac",
				"reqSigs": 1,
				"type": "pubkeyhash"
			}
		}
	]
}`,
		}, "transaction with multiple Inputs": {
			tx: func() *bt.Tx {
				tx := bt.NewTx()
				assert.NoError(t, tx.From(
					"3c8edde27cb9a9132c22038dac4391496be9db16fd21351565cc1006966fdad5",
					0,
					"76a914eb0bd5edba389198e73f8efabddfc61666969ff788ac",
					10000,
				))
				assert.NoError(t, tx.From(
					"3c8edde27cb9a9132c22038dac4391496be9db16fd21351565cc1006966fdad5",
					2,
					"76a914eb0bd5edba389198e73f8efabddfc61666969ff788ac",
					10000,
				))
				assert.NoError(t, tx.From(
					"3c8edde27cb9a9132c22038dac4391496be9db16fd21351565cc1006966fdad5",
					114,
					"76a914eb0bd5edba389198e73f8efabddfc61666969ff788ac",
					10000,
				))
				assert.NoError(t, tx.PayToAddress("n2wmGVP89x3DsLNqk3NvctfQy9m9pvt7mk", 1000))
				var w *wif.WIF
				w, err := wif.DecodeWIF("KznvCNc6Yf4iztSThoMH6oHWzH9EgjfodKxmeuUGPq5DEX5maspS")
				assert.NoError(t, err)
				assert.NotNil(t, w)
				err = tx.UnlockAll(context.Background(), &bt.LocalUnlockerGetter{PrivateKey: w.PrivKey})
				assert.NoError(t, err)
				return tx
			}(),
			expJSON: `{
	"version": 1,
	"locktime": 0,
	"txid": "41741af6fb64839c69f2385987eb3770c55c42eb6f7900fa2af9d667c42ceb20",
	"hash": "41741af6fb64839c69f2385987eb3770c55c42eb6f7900fa2af9d667c42ceb20",
	"size": 486,
	"hex": "0100000003d5da6f960610cc65153521fd16dbe96b499143ac8d03222c13a9b97ce2dd8e3c000000006b48304502210081214df575da1e9378f1d5a29dfd6811e93466a7222fb010b7c50dd2d44d7f2e0220399bb396336d2e294049e7db009926b1b30018ac834ee0cbca20b9d99f488038412102798913bc057b344de675dac34faafe3dc2f312c758cd9068209f810877306d66ffffffffd5da6f960610cc65153521fd16dbe96b499143ac8d03222c13a9b97ce2dd8e3c0200000069463043021f7059426d6aeb7d74275e52819a309b2bf903bd18b2b4d942d0e8e037681df702203f851f8a45aabfefdca5822f457609600f5d12a173adc09c6e7e2d4fdff7620a412102798913bc057b344de675dac34faafe3dc2f312c758cd9068209f810877306d66ffffffffd5da6f960610cc65153521fd16dbe96b499143ac8d03222c13a9b97ce2dd8e3c720000006b483045022100e7b3837f2818fe00a05293e0f90e9005d59b0c5c8890f22bd31c36190a9b55e9022027de4b77b78139ea21b9fd30876a447bbf29662bd19d7914028c607bccd772e4412102798913bc057b344de675dac34faafe3dc2f312c758cd9068209f810877306d66ffffffff01e8030000000000001976a914eb0bd5edba389198e73f8efabddfc61666969ff788ac00000000",
	"vin": [
		{
			"scriptSig": {
				"asm": "304502210081214df575da1e9378f1d5a29dfd6811e93466a7222fb010b7c50dd2d44d7f2e0220399bb396336d2e294049e7db009926b1b30018ac834ee0cbca20b9d99f48803841 02798913bc057b344de675dac34faafe3dc2f312c758cd9068209f810877306d66",
				"hex": "48304502210081214df575da1e9378f1d5a29dfd6811e93466a7222fb010b7c50dd2d44d7f2e0220399bb396336d2e294049e7db009926b1b30018ac834ee0cbca20b9d99f488038412102798913bc057b344de675dac34faafe3dc2f312c758cd9068209f810877306d66"
			},
			"txid": "3c8edde27cb9a9132c22038dac4391496be9db16fd21351565cc1006966fdad5",
			"vout": 0,
			"sequence": 4294967295
		},
		{
			"scriptSig": {
				"asm": "3043021f7059426d6aeb7d74275e52819a309b2bf903bd18b2b4d942d0e8e037681df702203f851f8a45aabfefdca5822f457609600f5d12a173adc09c6e7e2d4fdff7620a41 02798913bc057b344de675dac34faafe3dc2f312c758cd9068209f810877306d66",
				"hex": "463043021f7059426d6aeb7d74275e52819a309b2bf903bd18b2b4d942d0e8e037681df702203f851f8a45aabfefdca5822f457609600f5d12a173adc09c6e7e2d4fdff7620a412102798913bc057b344de675dac34faafe3dc2f312c758cd9068209f810877306d66"
			},
			"txid": "3c8edde27cb9a9132c22038dac4391496be9db16fd21351565cc1006966fdad5",
			"vout": 2,
			"sequence": 4294967295
		},
		{
			"scriptSig": {
				"asm": "3045022100e7b3837f2818fe00a05293e0f90e9005d59b0c5c8890f22bd31c36190a9b55e9022027de4b77b78139ea21b9fd30876a447bbf29662bd19d7914028c607bccd772e441 02798913bc057b344de675dac34faafe3dc2f312c758cd9068209f810877306d66",
				"hex": "483045022100e7b3837f2818fe00a05293e0f90e9005d59b0c5c8890f22bd31c36190a9b55e9022027de4b77b78139ea21b9fd30876a447bbf29662bd19d7914028c607bccd772e4412102798913bc057b344de675dac34faafe3dc2f312c758cd9068209f810877306d66"
			},
			"txid": "3c8edde27cb9a9132c22038dac4391496be9db16fd21351565cc1006966fdad5",
			"vout": 114,
			"sequence": 4294967295
		}
	],
	"vout": [
		{
			"value": 0.00001,
			"n": 0,
			"scriptPubKey": {
				"asm": "OP_DUP OP_HASH160 eb0bd5edba389198e73f8efabddfc61666969ff7 OP_EQUALVERIFY OP_CHECKSIG",
				"hex": "76a914eb0bd5edba389198e73f8efabddfc61666969ff788ac",
				"reqSigs": 1,
				"type": "pubkeyhash"
			}
		}
	]
}`,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			bb, err := json.MarshalIndent(test.tx.NodeJSON(), "", "\t")
			assert.NoError(t, err)
			assert.Equal(t, test.expJSON, string(bb))
		})
	}
}

func TestTxJSON_Node_UnmarshalJSON(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		json  string
		expTx *bt.Tx
	}{
		"json with hex should map correctly": {
			json: `{
				"version": 1,
				"locktime": 0,
				"txid": "aec245f27b7640c8b1865045107731bfb848115c573f7da38166074b1c9e475d",
				"hash": "aec245f27b7640c8b1865045107731bfb848115c573f7da38166074b1c9e475d",
				"size": 208,
				"hex": "0100000001abad53d72f342dd3f338e5e3346b492440f8ea821f8b8800e318f461cc5ea5a2010000006a4730440220042edc1302c5463e8397120a56b28ea381c8f7f6d9bdc1fee5ebca00c84a76e2022077069bbdb7ed701c4977b7db0aba80d41d4e693112256660bb5d674599e390cf41210294639d6e4249ea381c2e077e95c78fc97afe47a52eb24e1b1595cd3fdd0afdf8ffffffff02000000000000000008006a0548656c6c6f7f030000000000001976a914b85524abf8202a961b847a3bd0bc89d3d4d41cc588ac00000000",
				"vin": [
					{
						"scriptSig": {
							"asm": "30440220042edc1302c5463e8397120a56b28ea381c8f7f6d9bdc1fee5ebca00c84a76e2022077069bbdb7ed701c4977b7db0aba80d41d4e693112256660bb5d674599e390cf41 0294639d6e4249ea381c2e077e95c78fc97afe47a52eb24e1b1595cd3fdd0afdf8",
							"hex": "4730440220042edc1302c5463e8397120a56b28ea381c8f7f6d9bdc1fee5ebca00c84a76e2022077069bbdb7ed701c4977b7db0aba80d41d4e693112256660bb5d674599e390cf41210294639d6e4249ea381c2e077e95c78fc97afe47a52eb24e1b1595cd3fdd0afdf8"
						},
						"txid": "a2a55ecc61f418e300888b1f82eaf84024496b34e3e538f3d32d342fd753adab",
						"vout": 1,
						"sequence": 4294967295
					}
				],
				"vout": [
					{
						"value": 0,
						"n": 0,
						"scriptPubKey": {
							"asm": "OP_FALSE OP_RETURN 48656c6c6f",
							"hex": "006a0548656c6c6f",
							"type": "nulldata"
						}
					},
					{
						"value": 0.00000895,
						"n": 1,
						"scriptPubKey": {
							"asm": "OP_DUP OP_HASH160 b85524abf8202a961b847a3bd0bc89d3d4d41cc5 OP_EQUALVERIFY OP_CHECKSIG",
							"hex": "76a914b85524abf8202a961b847a3bd0bc89d3d4d41cc588ac",
							"reqSigs": 1,
							"type": "pubkeyhash"
						}
					}
				]
			}`,
			expTx: func() *bt.Tx {
				tx, err := bt.NewTxFromString("0100000001abad53d72f342dd3f338e5e3346b492440f8ea821f8b8800e318f461cc5ea5a2010000006a4730440220042edc1302c5463e8397120a56b28ea381c8f7f6d9bdc1fee5ebca00c84a76e2022077069bbdb7ed701c4977b7db0aba80d41d4e693112256660bb5d674599e390cf41210294639d6e4249ea381c2e077e95c78fc97afe47a52eb24e1b1595cd3fdd0afdf8ffffffff02000000000000000008006a0548656c6c6f7f030000000000001976a914b85524abf8202a961b847a3bd0bc89d3d4d41cc588ac00000000")
				assert.NoError(t, err)
				return tx
			}(),
		}, "ONLY hex should map correctly": {
			json: `{
				"hex": "0100000001abad53d72f342dd3f338e5e3346b492440f8ea821f8b8800e318f461cc5ea5a2010000006a4730440220042edc1302c5463e8397120a56b28ea381c8f7f6d9bdc1fee5ebca00c84a76e2022077069bbdb7ed701c4977b7db0aba80d41d4e693112256660bb5d674599e390cf41210294639d6e4249ea381c2e077e95c78fc97afe47a52eb24e1b1595cd3fdd0afdf8ffffffff02000000000000000008006a0548656c6c6f7f030000000000001976a914b85524abf8202a961b847a3bd0bc89d3d4d41cc588ac00000000"
			}`,
			expTx: func() *bt.Tx {
				tx, err := bt.NewTxFromString("0100000001abad53d72f342dd3f338e5e3346b492440f8ea821f8b8800e318f461cc5ea5a2010000006a4730440220042edc1302c5463e8397120a56b28ea381c8f7f6d9bdc1fee5ebca00c84a76e2022077069bbdb7ed701c4977b7db0aba80d41d4e693112256660bb5d674599e390cf41210294639d6e4249ea381c2e077e95c78fc97afe47a52eb24e1b1595cd3fdd0afdf8ffffffff02000000000000000008006a0548656c6c6f7f030000000000001976a914b85524abf8202a961b847a3bd0bc89d3d4d41cc588ac00000000")
				assert.NoError(t, err)
				return tx
			}(),
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			tx := bt.NewTx()
			err := json.Unmarshal([]byte(test.json), tx.NodeJSON())
			assert.NoError(t, err)
			assert.Equal(t, test.expTx, tx)
		})
	}
}

func TestTxJSON_Node_ToJson(t *testing.T) {
	tx, _ := bt.NewTxFromString("0100000001abad53d72f342dd3f338e5e3346b492440f8ea821f8b8800e318f461cc5ea5a2010000006a4730440220042edc1302c5463e8397120a56b28ea381c8f7f6d9bdc1fee5ebca00c84a76e2022077069bbdb7ed701c4977b7db0aba80d41d4e693112256660bb5d674599e390cf41210294639d6e4249ea381c2e077e95c78fc97afe47a52eb24e1b1595cd3fdd0afdf8ffffffff02000000000000000008006a0548656c6c6f7f030000000000001976a914b85524abf8202a961b847a3bd0bc89d3d4d41cc588ac00000000")

	_, err := json.MarshalIndent(tx.NodeJSON(), "", "\t")
	assert.NoError(t, err)
}

func TestTxsJSON_Node(t *testing.T) {
	tx, err := bt.NewTxFromString("0100000001abad53d72f342dd3f338e5e3346b492440f8ea821f8b8800e318f461cc5ea5a2010000006a4730440220042edc1302c5463e8397120a56b28ea381c8f7f6d9bdc1fee5ebca00c84a76e2022077069bbdb7ed701c4977b7db0aba80d41d4e693112256660bb5d674599e390cf41210294639d6e4249ea381c2e077e95c78fc97afe47a52eb24e1b1595cd3fdd0afdf8ffffffff02000000000000000008006a0548656c6c6f7f030000000000001976a914b85524abf8202a961b847a3bd0bc89d3d4d41cc588ac00000000")
	assert.NoError(t, err)

	tx2, err := bt.NewTxFromString("020000000117d2011c2a3b8a309d481930bae86e88017b0f55845ada17f96c464684b3af520000000048473044022014a60c3e84cf0160cb7e4ee7d87a3b78c5efb6dd3b66c76970b680affdb95e8f02207f6d9e3268a934e5e278ae513a3bc6dee3bec7bae37204574480305bfb5dea0e41feffffff0240101024010000001976a9149933e4bad50e7dd4b48c1f0be98436ca7d4392a288ac00e1f505000000001976a914abbe187ad301e4326e59587e43d602edd318364e88ac77000000")
	assert.NoError(t, err)

	txs := bt.Txs{tx, tx2}

	bb, err := json.Marshal(txs.NodeJSON())
	assert.NoError(t, err)

	nTxs := make(bt.Txs, 0)
	assert.NoError(t, json.Unmarshal(bb, nTxs.NodeJSON()))

	assert.Equal(t, len(txs), len(nTxs))
	for i, tx := range txs {
		assert.Equal(t, tx.String(), nTxs[i].String())
	}
}

func TestTxsJSON_Node_MarshallJSON(t *testing.T) {
	tests := map[string]struct {
		tx      bt.Txs
		expJSON string
	}{
		"transaction with 1 input 1 p2pksh output 1 data output should create valid json": {
			tx: func() bt.Txs {
				tx, err := bt.NewTxFromString("0100000001abad53d72f342dd3f338e5e3346b492440f8ea821f8b8800e318f461cc5ea5a2010000006a4730440220042edc1302c5463e8397120a56b28ea381c8f7f6d9bdc1fee5ebca00c84a76e2022077069bbdb7ed701c4977b7db0aba80d41d4e693112256660bb5d674599e390cf41210294639d6e4249ea381c2e077e95c78fc97afe47a52eb24e1b1595cd3fdd0afdf8ffffffff02000000000000000008006a0548656c6c6f7f030000000000001976a914b85524abf8202a961b847a3bd0bc89d3d4d41cc588ac00000000")
				assert.NoError(t, err)

				tx2, err := bt.NewTxFromString("020000000117d2011c2a3b8a309d481930bae86e88017b0f55845ada17f96c464684b3af520000000048473044022014a60c3e84cf0160cb7e4ee7d87a3b78c5efb6dd3b66c76970b680affdb95e8f02207f6d9e3268a934e5e278ae513a3bc6dee3bec7bae37204574480305bfb5dea0e41feffffff0240101024010000001976a9149933e4bad50e7dd4b48c1f0be98436ca7d4392a288ac00e1f505000000001976a914abbe187ad301e4326e59587e43d602edd318364e88ac77000000")
				assert.NoError(t, err)
				return bt.Txs{tx, tx2}
			}(),
			expJSON: `[
	{
		"version": 1,
		"locktime": 0,
		"txid": "aec245f27b7640c8b1865045107731bfb848115c573f7da38166074b1c9e475d",
		"hash": "aec245f27b7640c8b1865045107731bfb848115c573f7da38166074b1c9e475d",
		"size": 208,
		"hex": "0100000001abad53d72f342dd3f338e5e3346b492440f8ea821f8b8800e318f461cc5ea5a2010000006a4730440220042edc1302c5463e8397120a56b28ea381c8f7f6d9bdc1fee5ebca00c84a76e2022077069bbdb7ed701c4977b7db0aba80d41d4e693112256660bb5d674599e390cf41210294639d6e4249ea381c2e077e95c78fc97afe47a52eb24e1b1595cd3fdd0afdf8ffffffff02000000000000000008006a0548656c6c6f7f030000000000001976a914b85524abf8202a961b847a3bd0bc89d3d4d41cc588ac00000000",
		"vin": [
			{
				"scriptSig": {
					"asm": "30440220042edc1302c5463e8397120a56b28ea381c8f7f6d9bdc1fee5ebca00c84a76e2022077069bbdb7ed701c4977b7db0aba80d41d4e693112256660bb5d674599e390cf41 0294639d6e4249ea381c2e077e95c78fc97afe47a52eb24e1b1595cd3fdd0afdf8",
					"hex": "4730440220042edc1302c5463e8397120a56b28ea381c8f7f6d9bdc1fee5ebca00c84a76e2022077069bbdb7ed701c4977b7db0aba80d41d4e693112256660bb5d674599e390cf41210294639d6e4249ea381c2e077e95c78fc97afe47a52eb24e1b1595cd3fdd0afdf8"
				},
				"txid": "a2a55ecc61f418e300888b1f82eaf84024496b34e3e538f3d32d342fd753adab",
				"vout": 1,
				"sequence": 4294967295
			}
		],
		"vout": [
			{
				"value": 0,
				"n": 0,
				"scriptPubKey": {
					"asm": "OP_FALSE OP_RETURN 48656c6c6f",
					"hex": "006a0548656c6c6f",
					"type": "nulldata"
				}
			},
			{
				"value": 0.00000895,
				"n": 1,
				"scriptPubKey": {
					"asm": "OP_DUP OP_HASH160 b85524abf8202a961b847a3bd0bc89d3d4d41cc5 OP_EQUALVERIFY OP_CHECKSIG",
					"hex": "76a914b85524abf8202a961b847a3bd0bc89d3d4d41cc588ac",
					"reqSigs": 1,
					"type": "pubkeyhash"
				}
			}
		]
	},
	{
		"version": 2,
		"locktime": 119,
		"txid": "35d2d1db9bb3d1398faaa5addfc6aeaa6b3f1357d00098660b1554d4466d99b2",
		"hash": "35d2d1db9bb3d1398faaa5addfc6aeaa6b3f1357d00098660b1554d4466d99b2",
		"size": 191,
		"hex": "020000000117d2011c2a3b8a309d481930bae86e88017b0f55845ada17f96c464684b3af520000000048473044022014a60c3e84cf0160cb7e4ee7d87a3b78c5efb6dd3b66c76970b680affdb95e8f02207f6d9e3268a934e5e278ae513a3bc6dee3bec7bae37204574480305bfb5dea0e41feffffff0240101024010000001976a9149933e4bad50e7dd4b48c1f0be98436ca7d4392a288ac00e1f505000000001976a914abbe187ad301e4326e59587e43d602edd318364e88ac77000000",
		"vin": [
			{
				"scriptSig": {
					"asm": "3044022014a60c3e84cf0160cb7e4ee7d87a3b78c5efb6dd3b66c76970b680affdb95e8f02207f6d9e3268a934e5e278ae513a3bc6dee3bec7bae37204574480305bfb5dea0e41",
					"hex": "473044022014a60c3e84cf0160cb7e4ee7d87a3b78c5efb6dd3b66c76970b680affdb95e8f02207f6d9e3268a934e5e278ae513a3bc6dee3bec7bae37204574480305bfb5dea0e41"
				},
				"txid": "52afb38446466cf917da5a84550f7b01886ee8ba3019489d308a3b2a1c01d217",
				"vout": 0,
				"sequence": 4294967294
			}
		],
		"vout": [
			{
				"value": 48.99999808,
				"n": 0,
				"scriptPubKey": {
					"asm": "OP_DUP OP_HASH160 9933e4bad50e7dd4b48c1f0be98436ca7d4392a2 OP_EQUALVERIFY OP_CHECKSIG",
					"hex": "76a9149933e4bad50e7dd4b48c1f0be98436ca7d4392a288ac",
					"reqSigs": 1,
					"type": "pubkeyhash"
				}
			},
			{
				"value": 1,
				"n": 1,
				"scriptPubKey": {
					"asm": "OP_DUP OP_HASH160 abbe187ad301e4326e59587e43d602edd318364e OP_EQUALVERIFY OP_CHECKSIG",
					"hex": "76a914abbe187ad301e4326e59587e43d602edd318364e88ac",
					"reqSigs": 1,
					"type": "pubkeyhash"
				}
			}
		]
	}
]`,
		}, "transaction with multiple Inputs": {
			tx: func() bt.Txs {
				tx := bt.NewTx()
				assert.NoError(t, tx.From(
					"3c8edde27cb9a9132c22038dac4391496be9db16fd21351565cc1006966fdad5",
					0,
					"76a914eb0bd5edba389198e73f8efabddfc61666969ff788ac",
					10000,
				))
				assert.NoError(t, tx.From(
					"3c8edde27cb9a9132c22038dac4391496be9db16fd21351565cc1006966fdad5",
					2,
					"76a914eb0bd5edba389198e73f8efabddfc61666969ff788ac",
					10000,
				))
				assert.NoError(t, tx.From(
					"3c8edde27cb9a9132c22038dac4391496be9db16fd21351565cc1006966fdad5",
					114,
					"76a914eb0bd5edba389198e73f8efabddfc61666969ff788ac",
					10000,
				))
				assert.NoError(t, tx.PayToAddress("n2wmGVP89x3DsLNqk3NvctfQy9m9pvt7mk", 1000))
				var w *wif.WIF
				w, err := wif.DecodeWIF("KznvCNc6Yf4iztSThoMH6oHWzH9EgjfodKxmeuUGPq5DEX5maspS")
				assert.NoError(t, err)
				assert.NotNil(t, w)
				err = tx.UnlockAll(context.Background(), &bt.LocalUnlockerGetter{PrivateKey: w.PrivKey})
				assert.NoError(t, err)

				tx2, err := bt.NewTxFromString("020000000117d2011c2a3b8a309d481930bae86e88017b0f55845ada17f96c464684b3af520000000048473044022014a60c3e84cf0160cb7e4ee7d87a3b78c5efb6dd3b66c76970b680affdb95e8f02207f6d9e3268a934e5e278ae513a3bc6dee3bec7bae37204574480305bfb5dea0e41feffffff0240101024010000001976a9149933e4bad50e7dd4b48c1f0be98436ca7d4392a288ac00e1f505000000001976a914abbe187ad301e4326e59587e43d602edd318364e88ac77000000")
				assert.NoError(t, err)

				return bt.Txs{tx, tx2}
			}(),
			expJSON: `[
	{
		"version": 1,
		"locktime": 0,
		"txid": "41741af6fb64839c69f2385987eb3770c55c42eb6f7900fa2af9d667c42ceb20",
		"hash": "41741af6fb64839c69f2385987eb3770c55c42eb6f7900fa2af9d667c42ceb20",
		"size": 486,
		"hex": "0100000003d5da6f960610cc65153521fd16dbe96b499143ac8d03222c13a9b97ce2dd8e3c000000006b48304502210081214df575da1e9378f1d5a29dfd6811e93466a7222fb010b7c50dd2d44d7f2e0220399bb396336d2e294049e7db009926b1b30018ac834ee0cbca20b9d99f488038412102798913bc057b344de675dac34faafe3dc2f312c758cd9068209f810877306d66ffffffffd5da6f960610cc65153521fd16dbe96b499143ac8d03222c13a9b97ce2dd8e3c0200000069463043021f7059426d6aeb7d74275e52819a309b2bf903bd18b2b4d942d0e8e037681df702203f851f8a45aabfefdca5822f457609600f5d12a173adc09c6e7e2d4fdff7620a412102798913bc057b344de675dac34faafe3dc2f312c758cd9068209f810877306d66ffffffffd5da6f960610cc65153521fd16dbe96b499143ac8d03222c13a9b97ce2dd8e3c720000006b483045022100e7b3837f2818fe00a05293e0f90e9005d59b0c5c8890f22bd31c36190a9b55e9022027de4b77b78139ea21b9fd30876a447bbf29662bd19d7914028c607bccd772e4412102798913bc057b344de675dac34faafe3dc2f312c758cd9068209f810877306d66ffffffff01e8030000000000001976a914eb0bd5edba389198e73f8efabddfc61666969ff788ac00000000",
		"vin": [
			{
				"scriptSig": {
					"asm": "304502210081214df575da1e9378f1d5a29dfd6811e93466a7222fb010b7c50dd2d44d7f2e0220399bb396336d2e294049e7db009926b1b30018ac834ee0cbca20b9d99f48803841 02798913bc057b344de675dac34faafe3dc2f312c758cd9068209f810877306d66",
					"hex": "48304502210081214df575da1e9378f1d5a29dfd6811e93466a7222fb010b7c50dd2d44d7f2e0220399bb396336d2e294049e7db009926b1b30018ac834ee0cbca20b9d99f488038412102798913bc057b344de675dac34faafe3dc2f312c758cd9068209f810877306d66"
				},
				"txid": "3c8edde27cb9a9132c22038dac4391496be9db16fd21351565cc1006966fdad5",
				"vout": 0,
				"sequence": 4294967295
			},
			{
				"scriptSig": {
					"asm": "3043021f7059426d6aeb7d74275e52819a309b2bf903bd18b2b4d942d0e8e037681df702203f851f8a45aabfefdca5822f457609600f5d12a173adc09c6e7e2d4fdff7620a41 02798913bc057b344de675dac34faafe3dc2f312c758cd9068209f810877306d66",
					"hex": "463043021f7059426d6aeb7d74275e52819a309b2bf903bd18b2b4d942d0e8e037681df702203f851f8a45aabfefdca5822f457609600f5d12a173adc09c6e7e2d4fdff7620a412102798913bc057b344de675dac34faafe3dc2f312c758cd9068209f810877306d66"
				},
				"txid": "3c8edde27cb9a9132c22038dac4391496be9db16fd21351565cc1006966fdad5",
				"vout": 2,
				"sequence": 4294967295
			},
			{
				"scriptSig": {
					"asm": "3045022100e7b3837f2818fe00a05293e0f90e9005d59b0c5c8890f22bd31c36190a9b55e9022027de4b77b78139ea21b9fd30876a447bbf29662bd19d7914028c607bccd772e441 02798913bc057b344de675dac34faafe3dc2f312c758cd9068209f810877306d66",
					"hex": "483045022100e7b3837f2818fe00a05293e0f90e9005d59b0c5c8890f22bd31c36190a9b55e9022027de4b77b78139ea21b9fd30876a447bbf29662bd19d7914028c607bccd772e4412102798913bc057b344de675dac34faafe3dc2f312c758cd9068209f810877306d66"
				},
				"txid": "3c8edde27cb9a9132c22038dac4391496be9db16fd21351565cc1006966fdad5",
				"vout": 114,
				"sequence": 4294967295
			}
		],
		"vout": [
			{
				"value": 0.00001,
				"n": 0,
				"scriptPubKey": {
					"asm": "OP_DUP OP_HASH160 eb0bd5edba389198e73f8efabddfc61666969ff7 OP_EQUALVERIFY OP_CHECKSIG",
					"hex": "76a914eb0bd5edba389198e73f8efabddfc61666969ff788ac",
					"reqSigs": 1,
					"type": "pubkeyhash"
				}
			}
		]
	},
	{
		"version": 2,
		"locktime": 119,
		"txid": "35d2d1db9bb3d1398faaa5addfc6aeaa6b3f1357d00098660b1554d4466d99b2",
		"hash": "35d2d1db9bb3d1398faaa5addfc6aeaa6b3f1357d00098660b1554d4466d99b2",
		"size": 191,
		"hex": "020000000117d2011c2a3b8a309d481930bae86e88017b0f55845ada17f96c464684b3af520000000048473044022014a60c3e84cf0160cb7e4ee7d87a3b78c5efb6dd3b66c76970b680affdb95e8f02207f6d9e3268a934e5e278ae513a3bc6dee3bec7bae37204574480305bfb5dea0e41feffffff0240101024010000001976a9149933e4bad50e7dd4b48c1f0be98436ca7d4392a288ac00e1f505000000001976a914abbe187ad301e4326e59587e43d602edd318364e88ac77000000",
		"vin": [
			{
				"scriptSig": {
					"asm": "3044022014a60c3e84cf0160cb7e4ee7d87a3b78c5efb6dd3b66c76970b680affdb95e8f02207f6d9e3268a934e5e278ae513a3bc6dee3bec7bae37204574480305bfb5dea0e41",
					"hex": "473044022014a60c3e84cf0160cb7e4ee7d87a3b78c5efb6dd3b66c76970b680affdb95e8f02207f6d9e3268a934e5e278ae513a3bc6dee3bec7bae37204574480305bfb5dea0e41"
				},
				"txid": "52afb38446466cf917da5a84550f7b01886ee8ba3019489d308a3b2a1c01d217",
				"vout": 0,
				"sequence": 4294967294
			}
		],
		"vout": [
			{
				"value": 48.99999808,
				"n": 0,
				"scriptPubKey": {
					"asm": "OP_DUP OP_HASH160 9933e4bad50e7dd4b48c1f0be98436ca7d4392a2 OP_EQUALVERIFY OP_CHECKSIG",
					"hex": "76a9149933e4bad50e7dd4b48c1f0be98436ca7d4392a288ac",
					"reqSigs": 1,
					"type": "pubkeyhash"
				}
			},
			{
				"value": 1,
				"n": 1,
				"scriptPubKey": {
					"asm": "OP_DUP OP_HASH160 abbe187ad301e4326e59587e43d602edd318364e OP_EQUALVERIFY OP_CHECKSIG",
					"hex": "76a914abbe187ad301e4326e59587e43d602edd318364e88ac",
					"reqSigs": 1,
					"type": "pubkeyhash"
				}
			}
		]
	}
]`,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			bb, err := json.MarshalIndent(test.tx.NodeJSON(), "", "\t")
			assert.NoError(t, err)
			assert.Equal(t, test.expJSON, string(bb))
		})
	}
}

func TestTxsJSON_Node_UnmarshalJSON(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		json   string
		expTxs bt.Txs
	}{
		"node json should map correctly": {
			json: `[
	{
		"txid": "aec245f27b7640c8b1865045107731bfb848115c573f7da38166074b1c9e475d",
		"hex": "0100000001abad53d72f342dd3f338e5e3346b492440f8ea821f8b8800e318f461cc5ea5a2010000006a4730440220042edc1302c5463e8397120a56b28ea381c8f7f6d9bdc1fee5ebca00c84a76e2022077069bbdb7ed701c4977b7db0aba80d41d4e693112256660bb5d674599e390cf41210294639d6e4249ea381c2e077e95c78fc97afe47a52eb24e1b1595cd3fdd0afdf8ffffffff02000000000000000008006a0548656c6c6f7f030000000000001976a914b85524abf8202a961b847a3bd0bc89d3d4d41cc588ac00000000",
		"inputs": [
			{
				"unlockingScript": "4730440220042edc1302c5463e8397120a56b28ea381c8f7f6d9bdc1fee5ebca00c84a76e2022077069bbdb7ed701c4977b7db0aba80d41d4e693112256660bb5d674599e390cf41210294639d6e4249ea381c2e077e95c78fc97afe47a52eb24e1b1595cd3fdd0afdf8",
				"txid": "a2a55ecc61f418e300888b1f82eaf84024496b34e3e538f3d32d342fd753adab",
				"vout": 1,
				"sequence": 4294967295
			}
		],
		"outputs": [
			{
				"satoshis": 0,
				"lockingScript": "006a0548656c6c6f"
			},
			{
				"satoshis": 895,
				"lockingScript": "76a914b85524abf8202a961b847a3bd0bc89d3d4d41cc588ac"
			}
		],
		"version": 1,
		"lockTime": 0
	},
	{
		"txid": "35d2d1db9bb3d1398faaa5addfc6aeaa6b3f1357d00098660b1554d4466d99b2",
		"hex": "020000000117d2011c2a3b8a309d481930bae86e88017b0f55845ada17f96c464684b3af520000000048473044022014a60c3e84cf0160cb7e4ee7d87a3b78c5efb6dd3b66c76970b680affdb95e8f02207f6d9e3268a934e5e278ae513a3bc6dee3bec7bae37204574480305bfb5dea0e41feffffff0240101024010000001976a9149933e4bad50e7dd4b48c1f0be98436ca7d4392a288ac00e1f505000000001976a914abbe187ad301e4326e59587e43d602edd318364e88ac77000000",
		"inputs": [
			{
				"unlockingScript": "473044022014a60c3e84cf0160cb7e4ee7d87a3b78c5efb6dd3b66c76970b680affdb95e8f02207f6d9e3268a934e5e278ae513a3bc6dee3bec7bae37204574480305bfb5dea0e41",
				"txid": "52afb38446466cf917da5a84550f7b01886ee8ba3019489d308a3b2a1c01d217",
				"vout": 0,
				"sequence": 4294967294
			}
		],
		"outputs": [
			{
				"satoshis": 4899999808,
				"lockingScript": "76a9149933e4bad50e7dd4b48c1f0be98436ca7d4392a288ac"
			},
			{
				"satoshis": 100000000,
				"lockingScript": "76a914abbe187ad301e4326e59587e43d602edd318364e88ac"
			}
		],
		"version": 2,
		"lockTime": 119
	}
]`,
			expTxs: func() bt.Txs {
				tx, err := bt.NewTxFromString("0100000001abad53d72f342dd3f338e5e3346b492440f8ea821f8b8800e318f461cc5ea5a2010000006a4730440220042edc1302c5463e8397120a56b28ea381c8f7f6d9bdc1fee5ebca00c84a76e2022077069bbdb7ed701c4977b7db0aba80d41d4e693112256660bb5d674599e390cf41210294639d6e4249ea381c2e077e95c78fc97afe47a52eb24e1b1595cd3fdd0afdf8ffffffff02000000000000000008006a0548656c6c6f7f030000000000001976a914b85524abf8202a961b847a3bd0bc89d3d4d41cc588ac00000000")
				assert.NoError(t, err)

				tx2, err := bt.NewTxFromString("020000000117d2011c2a3b8a309d481930bae86e88017b0f55845ada17f96c464684b3af520000000048473044022014a60c3e84cf0160cb7e4ee7d87a3b78c5efb6dd3b66c76970b680affdb95e8f02207f6d9e3268a934e5e278ae513a3bc6dee3bec7bae37204574480305bfb5dea0e41feffffff0240101024010000001976a9149933e4bad50e7dd4b48c1f0be98436ca7d4392a288ac00e1f505000000001976a914abbe187ad301e4326e59587e43d602edd318364e88ac77000000")
				assert.NoError(t, err)

				return bt.Txs{tx, tx2}
			}(),
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			txs := make(bt.Txs, 0)
			err := json.Unmarshal([]byte(test.json), txs.NodeJSON())
			assert.NoError(t, err)
			assert.Equal(t, test.expTxs, txs)
		})
	}
}

func TestOutput_Node_JSON(t *testing.T) {
	tests := map[string]struct {
		output  *bt.Output
		expJSON string
	}{
		"node json": {
			output: &bt.Output{
				Satoshis: 10000,
				LockingScript: func() *bscript.Script {
					s, err := bscript.NewFromASM("OP_4 OP_2 OP_2 OP_ADD OP_EQUAL")
					assert.NoError(t, err)

					return s
				}(),
			},
			expJSON: `{
	"value": 0.0001,
	"n": 0,
	"scriptPubKey": {
		"asm": "OP_4 OP_2 OP_2 OP_ADD OP_EQUAL",
		"hex": "5452529387",
		"type": "nonstandard"
	}
}`,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			bb, err := json.MarshalIndent(test.output.NodeJSON(), "", "\t")
			assert.NoError(t, err)

			assert.Equal(t, test.expJSON, string(bb))
		})
	}
}

func TestOutput_JSON(t *testing.T) {
	tests := map[string]struct {
		output  *bt.Output
		expJSON string
	}{
		"standard json": {
			output: &bt.Output{
				Satoshis: 10000,
				LockingScript: func() *bscript.Script {
					s, err := bscript.NewFromASM("OP_4 OP_2 OP_2 OP_ADD OP_EQUAL")
					assert.NoError(t, err)

					return s
				}(),
			},
			expJSON: `{
	"satoshis": 10000,
	"lockingScript": "5452529387"
}`,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			bb, err := json.MarshalIndent(test.output, "", "\t")
			assert.NoError(t, err)

			assert.Equal(t, test.expJSON, string(bb))
		})
	}
}

func TestOutput_Node_UnmarshalJSON(t *testing.T) {
	tests := map[string]struct {
		json      string
		expOutput *bt.Output
	}{
		"node json": {
			json: `{
	"value": 0.0001,
	"n": 0,
	"scriptPubKey": {
		"asm": "OP_4 OP_2 OP_2 OP_ADD OP_EQUAL",
		"hex": "5452529387",
		"type": "nonstandard"
	}
}`,
			expOutput: &bt.Output{
				Satoshis: 10000,
				LockingScript: func() *bscript.Script {
					s, err := bscript.NewFromASM("OP_4 OP_2 OP_2 OP_ADD OP_EQUAL")
					assert.NoError(t, err)

					return s
				}(),
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			output := &bt.Output{}
			assert.NoError(t, json.Unmarshal([]byte(test.json), output.NodeJSON()))

			assert.Equal(t, *test.expOutput, *output)
		})
	}
}

func TestOutput_UnmarshalJSON(t *testing.T) {
	tests := map[string]struct {
		json      string
		expOutput *bt.Output
	}{
		"node json": {
			json: `{
	"satoshis": 10000,
	"lockingScript": "5452529387"
}`,
			expOutput: &bt.Output{
				Satoshis: 10000,
				LockingScript: func() *bscript.Script {
					s, err := bscript.NewFromASM("OP_4 OP_2 OP_2 OP_ADD OP_EQUAL")
					assert.NoError(t, err)

					return s
				}(),
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			output := &bt.Output{}
			assert.NoError(t, json.Unmarshal([]byte(test.json), output))

			assert.Equal(t, *test.expOutput, *output)
		})
	}
}
