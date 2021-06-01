package bt_test

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/libsv/go-bk/wif"
	. "github.com/libsv/go-bk/wif"
	"github.com/libsv/go-bt"
	"github.com/libsv/go-bt/bscript"
	"github.com/stretchr/testify/assert"
)

func TestNewTx(t *testing.T) {
	t.Parallel()

	t.Run("new tx, defaults", func(t *testing.T) {
		tx := bt.NewTx()
		assert.NotNil(t, tx)
		assert.IsType(t, &bt.Tx{}, tx)
		assert.Equal(t, uint32(1), tx.Version)
		assert.Equal(t, uint32(0), tx.LockTime)
		assert.Equal(t, 0, tx.InputCount())
		assert.Equal(t, 0, tx.OutputCount())
		assert.Equal(t, uint64(0), tx.TotalOutputSatoshis())
		assert.Equal(t, uint64(0), tx.TotalInputSatoshis())
	})
}

func TestNewTxFromString(t *testing.T) {
	t.Parallel()

	t.Run("valid tx no inputs", func(t *testing.T) {
		tx, err := bt.NewTxFromString("01000000000100000000000000001a006a07707265666978310c6578616d706c65206461746102133700000000")
		assert.NoError(t, err)
		assert.NotNil(t, tx)
	})

	t.Run("invalid tx", func(t *testing.T) {
		tx, err := bt.NewTxFromString("0")
		assert.Error(t, err)
		assert.Nil(t, tx)
	})

	t.Run("invalid tx - too short", func(t *testing.T) {
		tx, err := bt.NewTxFromString("000000")
		assert.Error(t, err)
		assert.Nil(t, tx)
	})

	t.Run("valid tx, 1 input, 1 output", func(t *testing.T) {
		rawTx := "02000000011ccba787d421b98904da3329b2c7336f368b62e89bc896019b5eadaa28145b9c000000004847304402205cc711985ce2a6d61eece4f9b6edd6337bad3b7eca3aa3ce59bc15620d8de2a80220410c92c48a226ba7d5a9a01105524097f673f31320d46c3b61d2378e6f05320041ffffffff01c0aff629010000001976a91418392a59fc1f76ad6a3c7ffcea20cfcb17bda9eb88ac00000000"
		tx, err := bt.NewTxFromString(rawTx)
		assert.NoError(t, err)
		assert.NotNil(t, tx)

		// Check version, locktime, inputs
		assert.Equal(t, uint32(2), tx.Version)
		assert.Equal(t, uint32(0), tx.LockTime)
		assert.Equal(t, 1, len(tx.Inputs))

		// Create a new unlocking script
		ptid, _ := hex.DecodeString("9c5b1428aaad5e9b0196c89be8628b366f33c7b22933da0489b921d487a7cb1c")
		i := bt.Input{
			PreviousTxIDBytes:  ptid,
			PreviousTxOutIndex: 0,
			SequenceNumber:     bt.DefaultSequenceNumber,
		}
		i.UnlockingScript, err = bscript.NewFromHexString("47304402205cc711985ce2a6d61eece4f9b6edd6337bad3b7eca3aa3ce59bc15620d8de2a80220410c92c48a226ba7d5a9a01105524097f673f31320d46c3b61d2378e6f05320041")
		assert.NoError(t, err)
		assert.NotNil(t, i.UnlockingScript)

		// Check input type
		assert.Equal(t, true, reflect.DeepEqual(*tx.Inputs[0], i))

		// Check output
		assert.Equal(t, 1, len(tx.Outputs))

		// New output
		var ls *bscript.Script
		ls, err = bscript.NewFromHexString("76a91418392a59fc1f76ad6a3c7ffcea20cfcb17bda9eb88ac")
		assert.NoError(t, err)
		assert.NotNil(t, ls)

		// Check the type
		o := bt.Output{Satoshis: 4999000000, LockingScript: ls}
		assert.Equal(t, true, reflect.DeepEqual(*tx.Outputs[0], o))
	})
}

func TestNewTxFromBytes(t *testing.T) {
	t.Parallel()

	t.Run("valid tx", func(t *testing.T) {
		rawTx := "02000000011ccba787d421b98904da3329b2c7336f368b62e89bc896019b5eadaa28145b9c0000000049483045022100c4df63202a9aa2bea5c24ebf4418d145e81712072ef744a4b108174f1ef59218022006eb54cf904707b51625f521f8ed2226f7d34b62492ebe4ddcb1c639caf16c3c41ffffffff0140420f00000000001976a91418392a59fc1f76ad6a3c7ffcea20cfcb17bda9eb88ac00000000"
		b, err := hex.DecodeString(rawTx)
		assert.NoError(t, err)

		var tx *bt.Tx
		tx, err = bt.NewTxFromBytes(b)
		assert.NoError(t, err)
		assert.NotNil(t, tx)
	})

	t.Run("invalid tx, too short", func(t *testing.T) {
		rawTx := "000000"
		b, err := hex.DecodeString(rawTx)
		assert.NoError(t, err)

		var tx *bt.Tx
		tx, err = bt.NewTxFromBytes(b)
		assert.Error(t, err)
		assert.Nil(t, tx)
	})
}

func TestTx_TxID(t *testing.T) {
	t.Parallel()

	t.Run("valid tx id", func(t *testing.T) {
		tx, err := bt.NewTxFromString("010000000193a35408b6068499e0d5abd799d3e827d9bfe70c9b75ebe209c91d2507232651000000006b483045022100c1d77036dc6cd1f3fa1214b0688391ab7f7a16cd31ea4e5a1f7a415ef167df820220751aced6d24649fa235132f1e6969e163b9400f80043a72879237dab4a1190ad412103b8b40a84123121d260f5c109bc5a46ec819c2e4002e5ba08638783bfb4e01435ffffffff02404b4c00000000001976a91404ff367be719efa79d76e4416ffb072cd53b208888acde94a905000000001976a91404d03f746652cfcb6cb55119ab473a045137d26588ac00000000")
		assert.NoError(t, err)
		assert.NotNil(t, tx)
		assert.Equal(t, "19dcf16ecc9286c3734fdae3d45d4fc4eb6b25f841131e06460f4939bba0026e", tx.TxID())
	})

	t.Run("new tx, no data, but has default tx id", func(t *testing.T) {
		tx := bt.NewTx()
		assert.NotNil(t, tx)
		assert.Equal(t, "d21633ba23f70118185227be58a63527675641ad37967e2aa461559f577aec43", tx.TxID())
	})
}

func TestVersion(t *testing.T) {
	t.Parallel()

	rawTx := "01000000014c6ec863cf3e0284b407a1a1b8138c76f98280812cb9653231f385a0305fc76f010000006b483045022100f01c1a1679c9437398d691c8497f278fa2d615efc05115688bf2c3335b45c88602201b54437e54fb53bc50545de44ea8c64e9e583952771fcc663c8687dc2638f7854121037e87bbd3b680748a74372640628a8f32d3a841ceeef6f75626ab030c1a04824fffffffff021d784500000000001976a914e9b62e25d4c6f97287dfe62f8063b79a9638c84688ac60d64f00000000001976a914bb4bca2306df66d72c6e44a470873484d8808b8888ac00000000"
	tx, err := bt.NewTxFromString(rawTx)
	assert.NoError(t, err)
	assert.NotNil(t, tx)
	assert.Equal(t, uint32(1), tx.Version)
}

func TestTx_IsCoinbase(t *testing.T) {
	t.Parallel()

	t.Run("invalid number of inputs", func(t *testing.T) {
		tx := bt.NewTx()
		assert.NotNil(t, tx)
		assert.Equal(t, false, tx.IsCoinbase())
	})

	t.Run("valid coinbase tx, 1 input", func(t *testing.T) {
		rawTx := "02000000010000000000000000000000000000000000000000000000000000000000000000ffffffff0e5101010a2f4542323030302e302fffffffff0100f2052a01000000232103db233bb9fc387d78b133ec904069d46e95ff17da657671b44afa0bc64e89ac18ac00000000"
		tx, err := bt.NewTxFromString(rawTx)
		assert.NoError(t, err)
		assert.NotNil(t, tx)

		assert.Equal(t, true, tx.IsCoinbase())
		assert.Equal(t, 1, tx.InputCount())
	})

	t.Run("valid coinbase tx", func(t *testing.T) {
		coinbaseTx := "01000000010000000000000000000000000000000000000000000000000000000000000000ffffffff4303bfea07322f53696d6f6e204f726469736820616e642053747561727420467265656d616e206d61646520746869732068617070656e2f9a46434790f7dbdea3430000ffffffff018a08ac4a000000001976a9148bf10d323ac757268eb715e613cb8e8e1d1793aa88ac00000000"
		tx, err := bt.NewTxFromString(coinbaseTx)
		assert.NoError(t, err)
		assert.NotNil(t, tx)
		assert.Equal(t, true, tx.IsCoinbase())
	})

	t.Run("tx is not a coinbase tx", func(t *testing.T) {
		coinbaseTx := "01000000014c6ec863cf3e0284b407a1a1b8138c76f98280812cb9653231f385a0305fc76f010000006b483045022100f01c1a1679c9437398d691c8497f278fa2d615efc05115688bf2c3335b45c88602201b54437e54fb53bc50545de44ea8c64e9e583952771fcc663c8687dc2638f7854121037e87bbd3b680748a74372640628a8f32d3a841ceeef6f75626ab030c1a04824fffffffff021d784500000000001976a914e9b62e25d4c6f97287dfe62f8063b79a9638c84688ac60d64f00000000001976a914bb4bca2306df66d72c6e44a470873484d8808b8888ac00000000"
		tx, err := bt.NewTxFromString(coinbaseTx)
		assert.NoError(t, err)
		assert.NotNil(t, tx)
		assert.Equal(t, false, tx.IsCoinbase())
	})

	t.Run("tx (2) is not a coinbase tx", func(t *testing.T) {
		coinbaseTx := "010000000159ef0cbb7881f2c934d6fb669f68f7c6a9c632f997152f828d1153806b7ac82b010000006b483045022100e775a21994cc6d6d6bf79d295aeea592e7b4cf8d8ecddaf67bb6626d7af82fd302201921a313de67e23a78c81dd5fe9a19322839c0ea1034b9c54e8206dea3aa9e68412103d1c02ee3522ff58df6c6287e67202a797b562fa8b5a9ed86613fe5ee48fb8821ffffffff02000000000000000011006a0e6d657461737472656d652e636f6dc9990200000000001976a914fa1b02ff7e41975d698fec6fb1b2d7e4656f8e7f88ac00000000"
		tx, err := bt.NewTxFromString(coinbaseTx)
		assert.NoError(t, err)
		assert.NotNil(t, tx)
		assert.Equal(t, false, tx.IsCoinbase())
	})
}

func TestTx_CreateTx(t *testing.T) {
	t.Parallel()

	tx := bt.NewTx()
	assert.NotNil(t, tx)

	err := tx.From(
		"3c8edde27cb9a9132c22038dac4391496be9db16fd21351565cc1006966fdad5",
		0,
		"76a914eb0bd5edba389198e73f8efabddfc61666969ff788ac",
		2000000)
	assert.NoError(t, err)

	err = tx.PayTo("n2wmGVP89x3DsLNqk3NvctfQy9m9pvt7mk", 1999942)
	assert.NoError(t, err)

	var wif *WIF
	wif, err = DecodeWIF("KznvCNc6Yf4iztSThoMH6oHWzH9EgjfodKxmeuUGPq5DEX5maspS")
	assert.NoError(t, err)
	assert.NotNil(t, wif)

	_, err = tx.SignAuto(context.Background(), &bt.LocalSigner{PrivateKey: wif.PrivKey})
	assert.NoError(t, err)
}

func TestTx_HasDataOutputs(t *testing.T) {
	t.Parallel()

	t.Run("has data outputs", func(t *testing.T) {
		tx := bt.NewTx()
		assert.NotNil(t, tx)

		err := tx.From(
			"3c8edde27cb9a9132c22038dac4391496be9db16fd21351565cc1006966fdad5",
			0,
			"76a914eb0bd5edba389198e73f8efabddfc61666969ff788ac",
			2000000)
		assert.NoError(t, err)

		err = tx.PayTo("n2wmGVP89x3DsLNqk3NvctfQy9m9pvt7mk", 1999942)
		assert.NoError(t, err)

		// Add op return data
		type OpReturnData [][]byte
		ops := OpReturnData{[]byte("prefix1"), []byte("example data"), []byte{0x13, 0x37}}

		err = tx.AddOpReturnPartsOutput(ops)
		assert.NoError(t, err)

		var wif *WIF
		wif, err = DecodeWIF("KznvCNc6Yf4iztSThoMH6oHWzH9EgjfodKxmeuUGPq5DEX5maspS")
		assert.NoError(t, err)
		assert.NotNil(t, wif)

		_, err = tx.SignAuto(context.Background(), &bt.LocalSigner{PrivateKey: wif.PrivKey})
		assert.NoError(t, err)

		assert.Equal(t, true, tx.HasDataOutputs())
	})

	t.Run("no data outputs", func(t *testing.T) {
		tx := bt.NewTx()
		assert.NotNil(t, tx)

		err := tx.From(
			"3c8edde27cb9a9132c22038dac4391496be9db16fd21351565cc1006966fdad5",
			0,
			"76a914eb0bd5edba389198e73f8efabddfc61666969ff788ac",
			2000000)
		assert.NoError(t, err)

		err = tx.PayTo("n2wmGVP89x3DsLNqk3NvctfQy9m9pvt7mk", 1999942)
		assert.NoError(t, err)

		var wif *WIF
		wif, err = DecodeWIF("KznvCNc6Yf4iztSThoMH6oHWzH9EgjfodKxmeuUGPq5DEX5maspS")
		assert.NoError(t, err)
		assert.NotNil(t, wif)

		_, err = tx.SignAuto(context.Background(), &bt.LocalSigner{PrivateKey: wif.PrivKey})
		assert.NoError(t, err)

		assert.Equal(t, false, tx.HasDataOutputs())
	})
}

func TestTx_ToJson(t *testing.T) {
	tx, _ := bt.NewTxFromString("0100000001abad53d72f342dd3f338e5e3346b492440f8ea821f8b8800e318f461cc5ea5a2010000006a4730440220042edc1302c5463e8397120a56b28ea381c8f7f6d9bdc1fee5ebca00c84a76e2022077069bbdb7ed701c4977b7db0aba80d41d4e693112256660bb5d674599e390cf41210294639d6e4249ea381c2e077e95c78fc97afe47a52eb24e1b1595cd3fdd0afdf8ffffffff02000000000000000008006a0548656c6c6f7f030000000000001976a914b85524abf8202a961b847a3bd0bc89d3d4d41cc588ac00000000")

	bb, err := json.MarshalIndent(tx, "", "\t")
	assert.NoError(t, err)
	fmt.Println(string(bb))
}

func TestTx_JSON(t *testing.T) {
	tests := map[string]struct {
		tx  *bt.Tx
		err error
	}{
		"standard tx should marshal and unmarshall correctly": {
			tx: func() *bt.Tx {
				tx := bt.NewTx()
				assert.NoError(t, tx.From(
					"3c8edde27cb9a9132c22038dac4391496be9db16fd21351565cc1006966fdad5",
					0,
					"76a914eb0bd5edba389198e73f8efabddfc61666969ff788ac",
					2000000))
				assert.NoError(t, tx.PayTo("n2wmGVP89x3DsLNqk3NvctfQy9m9pvt7mk", 1000))
				var wif *WIF
				wif, err := DecodeWIF("KznvCNc6Yf4iztSThoMH6oHWzH9EgjfodKxmeuUGPq5DEX5maspS")
				assert.NoError(t, err)
				assert.NotNil(t, wif)

				_, err = tx.SignAuto(context.Background(), &bt.LocalSigner{PrivateKey: wif.PrivKey})
				assert.NoError(t, err)
				return tx
			}(),
		}, "data tx should marshall correctly": {
			tx: func() *bt.Tx {
				tx := bt.NewTx()
				assert.NoError(t, tx.From(
					"3c8edde27cb9a9132c22038dac4391496be9db16fd21351565cc1006966fdad5",
					0,
					"76a914eb0bd5edba389198e73f8efabddfc61666969ff788ac",
					2000000))
				assert.NoError(t, tx.PayTo("n2wmGVP89x3DsLNqk3NvctfQy9m9pvt7mk", 1000))
				var wif *WIF
				wif, err := DecodeWIF("KznvCNc6Yf4iztSThoMH6oHWzH9EgjfodKxmeuUGPq5DEX5maspS")
				assert.NoError(t, err)
				assert.NotNil(t, wif)
				s := &bscript.Script{}
				assert.NoError(t, s.AppendPushDataString("test"))
				tx.AddOutput(&bt.Output{
					LockingScript: s,
				})
				_, err = tx.SignAuto(context.Background(), &bt.LocalSigner{PrivateKey: wif.PrivKey})
				assert.NoError(t, err)
				return tx
			}(),
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			bb, err := json.Marshal(test.tx)
			assert.NoError(t, err)
			if err != nil {
				return
			}
			var tx *bt.Tx
			assert.NoError(t, json.Unmarshal(bb, &tx))
			assert.Equal(t, test.tx.String(), tx.String())
		})
	}
}

func TestTx_MarshallJSON(t *testing.T) {
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
			"unlockingScript": {
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
			"satoshis": 0,
			"n": 0,
			"lockingScript": {
				"asm": "OP_FALSE OP_RETURN 48656c6c6f",
				"hex": "006a0548656c6c6f",
				"type": "nulldata"
			}
		},
		{
			"value": 0.00000895,
			"satoshis": 895,
			"n": 1,
			"lockingScript": {
				"asm": "OP_DUP OP_HASH160 b85524abf8202a961b847a3bd0bc89d3d4d41cc5 OP_EQUALVERIFY OP_CHECKSIG",
				"hex": "76a914b85524abf8202a961b847a3bd0bc89d3d4d41cc588ac",
				"reqSigs": 1,
				"type": "pubkeyhash"
			}
		}
	]
}`,
		}, "transaction with multiple inputs": {
			tx: func() *bt.Tx {
				tx := bt.NewTx()
				assert.NoError(t, tx.From(
					"3c8edde27cb9a9132c22038dac4391496be9db16fd21351565cc1006966fdad5",
					0,
					"76a914eb0bd5edba389198e73f8efabddfc61666969ff788ac",
					10000))
				assert.NoError(t, tx.From(
					"3c8edde27cb9a9132c22038dac4391496be9db16fd21351565cc1006966fdad5",
					2,
					"76a914eb0bd5edba389198e73f8efabddfc61666969ff788ac",
					10000))
				assert.NoError(t, tx.From(
					"3c8edde27cb9a9132c22038dac4391496be9db16fd21351565cc1006966fdad5",
					114,
					"76a914eb0bd5edba389198e73f8efabddfc61666969ff788ac",
					10000))
				assert.NoError(t, tx.PayTo("n2wmGVP89x3DsLNqk3NvctfQy9m9pvt7mk", 1000))
				var w *wif.WIF
				w, err := wif.DecodeWIF("KznvCNc6Yf4iztSThoMH6oHWzH9EgjfodKxmeuUGPq5DEX5maspS")
				assert.NoError(t, err)
				assert.NotNil(t, w)
				_, err = tx.SignAuto(context.Background(), &bt.LocalSigner{PrivateKey: w.PrivKey})
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
			"unlockingScript": {
				"asm": "304502210081214df575da1e9378f1d5a29dfd6811e93466a7222fb010b7c50dd2d44d7f2e0220399bb396336d2e294049e7db009926b1b30018ac834ee0cbca20b9d99f48803841 02798913bc057b344de675dac34faafe3dc2f312c758cd9068209f810877306d66",
				"hex": "48304502210081214df575da1e9378f1d5a29dfd6811e93466a7222fb010b7c50dd2d44d7f2e0220399bb396336d2e294049e7db009926b1b30018ac834ee0cbca20b9d99f488038412102798913bc057b344de675dac34faafe3dc2f312c758cd9068209f810877306d66"
			},
			"txid": "3c8edde27cb9a9132c22038dac4391496be9db16fd21351565cc1006966fdad5",
			"vout": 0,
			"sequence": 4294967295
		},
		{
			"unlockingScript": {
				"asm": "3043021f7059426d6aeb7d74275e52819a309b2bf903bd18b2b4d942d0e8e037681df702203f851f8a45aabfefdca5822f457609600f5d12a173adc09c6e7e2d4fdff7620a41 02798913bc057b344de675dac34faafe3dc2f312c758cd9068209f810877306d66",
				"hex": "463043021f7059426d6aeb7d74275e52819a309b2bf903bd18b2b4d942d0e8e037681df702203f851f8a45aabfefdca5822f457609600f5d12a173adc09c6e7e2d4fdff7620a412102798913bc057b344de675dac34faafe3dc2f312c758cd9068209f810877306d66"
			},
			"txid": "3c8edde27cb9a9132c22038dac4391496be9db16fd21351565cc1006966fdad5",
			"vout": 2,
			"sequence": 4294967295
		},
		{
			"unlockingScript": {
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
			"satoshis": 1000,
			"n": 0,
			"lockingScript": {
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
			bb, err := json.MarshalIndent(test.tx, "", "\t")
			assert.NoError(t, err)
			assert.Equal(t, test.expJSON, string(bb))
		})
	}
}

func TestTx_UnmarshalJSON(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		json  string
		expTX *bt.Tx
	}{
		"our json with hex should map correctly": {
			json: `{
				"version": 1,
				"locktime": 0,
				"txid": "aec245f27b7640c8b1865045107731bfb848115c573f7da38166074b1c9e475d",
				"hash": "aec245f27b7640c8b1865045107731bfb848115c573f7da38166074b1c9e475d",
				"size": 208,
				"hex": "0100000001abad53d72f342dd3f338e5e3346b492440f8ea821f8b8800e318f461cc5ea5a2010000006a4730440220042edc1302c5463e8397120a56b28ea381c8f7f6d9bdc1fee5ebca00c84a76e2022077069bbdb7ed701c4977b7db0aba80d41d4e693112256660bb5d674599e390cf41210294639d6e4249ea381c2e077e95c78fc97afe47a52eb24e1b1595cd3fdd0afdf8ffffffff02000000000000000008006a0548656c6c6f7f030000000000001976a914b85524abf8202a961b847a3bd0bc89d3d4d41cc588ac00000000",
				"vin": [
			{
				"unlockingScript": {
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
				"satoshis": 0,
				"n": 0,
				"lockingScript": {
				"asm": "OP_FALSE OP_RETURN 48656c6c6f",
				"hex": "006a0548656c6c6f",
				"type": "nulldata"
			}
			},
			{
				"value": 0.00000895,
				"satoshis": 895,
				"n": 1,
				"lockingScript": {
				"asm": "OP_DUP OP_HASH160 b85524abf8202a961b847a3bd0bc89d3d4d41cc5 OP_EQUALVERIFY OP_CHECKSIG",
				"hex": "76a914b85524abf8202a961b847a3bd0bc89d3d4d41cc588ac",
				"reqSigs": 1,
				"type": "pubkeyhash"
			}
			}
			]
			}`,
			expTX: func() *bt.Tx {
				tx, err := bt.NewTxFromString("0100000001abad53d72f342dd3f338e5e3346b492440f8ea821f8b8800e318f461cc5ea5a2010000006a4730440220042edc1302c5463e8397120a56b28ea381c8f7f6d9bdc1fee5ebca00c84a76e2022077069bbdb7ed701c4977b7db0aba80d41d4e693112256660bb5d674599e390cf41210294639d6e4249ea381c2e077e95c78fc97afe47a52eb24e1b1595cd3fdd0afdf8ffffffff02000000000000000008006a0548656c6c6f7f030000000000001976a914b85524abf8202a961b847a3bd0bc89d3d4d41cc588ac00000000")
				assert.NoError(t, err)
				return tx
			}(),
		}, "ONLY hex should map correctly": {
			json: `{
				"hex": "0100000001abad53d72f342dd3f338e5e3346b492440f8ea821f8b8800e318f461cc5ea5a2010000006a4730440220042edc1302c5463e8397120a56b28ea381c8f7f6d9bdc1fee5ebca00c84a76e2022077069bbdb7ed701c4977b7db0aba80d41d4e693112256660bb5d674599e390cf41210294639d6e4249ea381c2e077e95c78fc97afe47a52eb24e1b1595cd3fdd0afdf8ffffffff02000000000000000008006a0548656c6c6f7f030000000000001976a914b85524abf8202a961b847a3bd0bc89d3d4d41cc588ac00000000"
			}`,
			expTX: func() *bt.Tx {
				tx, err := bt.NewTxFromString("0100000001abad53d72f342dd3f338e5e3346b492440f8ea821f8b8800e318f461cc5ea5a2010000006a4730440220042edc1302c5463e8397120a56b28ea381c8f7f6d9bdc1fee5ebca00c84a76e2022077069bbdb7ed701c4977b7db0aba80d41d4e693112256660bb5d674599e390cf41210294639d6e4249ea381c2e077e95c78fc97afe47a52eb24e1b1595cd3fdd0afdf8ffffffff02000000000000000008006a0548656c6c6f7f030000000000001976a914b85524abf8202a961b847a3bd0bc89d3d4d41cc588ac00000000")
				assert.NoError(t, err)
				return tx
			}(),
		}, "Node json with hex should map correctly": {
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
			expTX: func() *bt.Tx {
				tx, err := bt.NewTxFromString("0100000001abad53d72f342dd3f338e5e3346b492440f8ea821f8b8800e318f461cc5ea5a2010000006a4730440220042edc1302c5463e8397120a56b28ea381c8f7f6d9bdc1fee5ebca00c84a76e2022077069bbdb7ed701c4977b7db0aba80d41d4e693112256660bb5d674599e390cf41210294639d6e4249ea381c2e077e95c78fc97afe47a52eb24e1b1595cd3fdd0afdf8ffffffff02000000000000000008006a0548656c6c6f7f030000000000001976a914b85524abf8202a961b847a3bd0bc89d3d4d41cc588ac00000000")
				assert.NoError(t, err)
				return tx
			}(),
		}, "Node json without hex should map correctly": {
			json: `{
	"version": 1,
	"locktime": 0,
	"txid": "aec245f27b7640c8b1865045107731bfb848115c573f7da38166074b1c9e475d",
	"hash": "aec245f27b7640c8b1865045107731bfb848115c573f7da38166074b1c9e475d",
	"size": 208,
	"vin": [{
		"scriptSig": {
			"asm": "30440220042edc1302c5463e8397120a56b28ea381c8f7f6d9bdc1fee5ebca00c84a76e2022077069bbdb7ed701c4977b7db0aba80d41d4e693112256660bb5d674599e390cf41 0294639d6e4249ea381c2e077e95c78fc97afe47a52eb24e1b1595cd3fdd0afdf8",
			"hex": "4730440220042edc1302c5463e8397120a56b28ea381c8f7f6d9bdc1fee5ebca00c84a76e2022077069bbdb7ed701c4977b7db0aba80d41d4e693112256660bb5d674599e390cf41210294639d6e4249ea381c2e077e95c78fc97afe47a52eb24e1b1595cd3fdd0afdf8"
		},
		"txid": "a2a55ecc61f418e300888b1f82eaf84024496b34e3e538f3d32d342fd753adab",
		"vout": 1,
		"sequence": 4294967295
	}],
	"vout": [{
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
			expTX: func() *bt.Tx {
				tx, err := bt.NewTxFromString("0100000001abad53d72f342dd3f338e5e3346b492440f8ea821f8b8800e318f461cc5ea5a2010000006a4730440220042edc1302c5463e8397120a56b28ea381c8f7f6d9bdc1fee5ebca00c84a76e2022077069bbdb7ed701c4977b7db0aba80d41d4e693112256660bb5d674599e390cf41210294639d6e4249ea381c2e077e95c78fc97afe47a52eb24e1b1595cd3fdd0afdf8ffffffff02000000000000000008006a0548656c6c6f7f030000000000001976a914b85524abf8202a961b847a3bd0bc89d3d4d41cc588ac00000000")
				assert.NoError(t, err)
				return tx
			}(),
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			var tx *bt.Tx
			err := json.Unmarshal([]byte(test.json), &tx)
			assert.NoError(t, err)
			assert.Equal(t, test.expTX, tx)
		})
	}
}
