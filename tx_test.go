package bt_test

import (
	"encoding/hex"
	"reflect"
	"testing"

	"github.com/bitcoinsv/bsvutil"
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
		assert.Equal(t, uint64(0), tx.GetTotalOutputSatoshis())
		assert.Equal(t, uint64(0), tx.GetTotalInputSatoshis())
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
			PreviousTxID:       "9c5b1428aaad5e9b0196c89be8628b366f33c7b22933da0489b921d487a7cb1c",
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

func TestAddInputFromTx(t *testing.T) {
	pubkey1 := []byte{1, 2, 3} // utxo test owner
	pubkey2 := []byte{1, 2, 4}

	output1, err1 := bt.NewP2PKHOutputFromPubKeyBytes(pubkey1, uint64(100000))
	assert.NoError(t, err1)
	output2, err2 := bt.NewP2PKHOutputFromPubKeyBytes(pubkey1, uint64(100000))
	assert.NoError(t, err2)
	output3, err3 := bt.NewP2PKHOutputFromPubKeyBytes(pubkey2, uint64(5000000))
	assert.NoError(t, err3)

	prvTx := bt.NewTx()
	prvTx.AddOutput(output1)
	prvTx.AddOutput(output2)
	prvTx.AddOutput(output3)
	newTx := bt.NewTx()
	err := newTx.AddInputFromTx(prvTx, pubkey1)
	assert.NoError(t, err)
	assert.Equal(t, newTx.InputCount(), 2) // only 2 utxos has been added
	assert.Equal(t, newTx.GetTotalInputSatoshis(), uint64(200000))
}

func TestTx_GetTxID(t *testing.T) {
	t.Parallel()

	t.Run("valid tx id", func(t *testing.T) {
		tx, err := bt.NewTxFromString("010000000193a35408b6068499e0d5abd799d3e827d9bfe70c9b75ebe209c91d2507232651000000006b483045022100c1d77036dc6cd1f3fa1214b0688391ab7f7a16cd31ea4e5a1f7a415ef167df820220751aced6d24649fa235132f1e6969e163b9400f80043a72879237dab4a1190ad412103b8b40a84123121d260f5c109bc5a46ec819c2e4002e5ba08638783bfb4e01435ffffffff02404b4c00000000001976a91404ff367be719efa79d76e4416ffb072cd53b208888acde94a905000000001976a91404d03f746652cfcb6cb55119ab473a045137d26588ac00000000")
		assert.NoError(t, err)
		assert.NotNil(t, tx)
		assert.Equal(t, "19dcf16ecc9286c3734fdae3d45d4fc4eb6b25f841131e06460f4939bba0026e", tx.GetTxID())
	})

	t.Run("new tx, no data, but has default tx id", func(t *testing.T) {
		tx := bt.NewTx()
		assert.NotNil(t, tx)
		assert.Equal(t, "d21633ba23f70118185227be58a63527675641ad37967e2aa461559f577aec43", tx.GetTxID())
	})
}

func TestTx_GetTotalOutputSatoshis(t *testing.T) {
	t.Parallel()

	t.Run("greater than zero", func(t *testing.T) {
		tx, err := bt.NewTxFromString("020000000180f1ada3ad8e861441d9ceab40b68ed98f13695b185cc516226a46697cc01f80010000006b483045022100fa3a0f8fa9fbf09c372b7a318fa6175d022c1d782f7b8bc5949a7c8f59ce3f35022005e0e84c26f26d892b484ff738d803a57626679389c8b302939460dab29a5308412103e46b62eea5db5898fb65f7dc840e8a1dbd8f08a19781a23f1f55914f9bedcd49feffffff02dec537b2000000001976a914ba11bcc46ecf8d88e0828ddbe87997bf759ca85988ac00943577000000001976a91418392a59fc1f76ad6a3c7ffcea20cfcb17bda9eb88ac6e000000")
		assert.NoError(t, err)
		assert.NotNil(t, tx)
		assert.Equal(t, uint64((29.89999582+20.00)*1e8), tx.GetTotalOutputSatoshis())
	})

	t.Run("zero outputs", func(t *testing.T) {
		tx := bt.NewTx()
		assert.NotNil(t, tx)
		assert.Equal(t, uint64(0), tx.GetTotalOutputSatoshis())
	})
}

func TestGetVersion(t *testing.T) {
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

	var wif *bsvutil.WIF
	wif, err = bsvutil.DecodeWIF("KznvCNc6Yf4iztSThoMH6oHWzH9EgjfodKxmeuUGPq5DEX5maspS")
	assert.NoError(t, err)
	assert.NotNil(t, wif)

	_, err = tx.SignAuto(&bt.InternalSigner{PrivateKey: wif.PrivKey, SigHashFlag: 0})
	assert.NoError(t, err)
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

func TestTx_PayTo(t *testing.T) {
	t.Run("missing pay to address", func(t *testing.T) {
		tx := bt.NewTx()
		assert.NotNil(t, tx)
		err := tx.From(
			"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
			0,
			"76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
			4000000)
		assert.NoError(t, err)

		err = tx.PayTo("", 100)
		assert.Error(t, err)
	})

	t.Run("invalid pay to address", func(t *testing.T) {
		tx := bt.NewTx()
		assert.NotNil(t, tx)
		err := tx.From(
			"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
			0,
			"76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
			4000000)
		assert.NoError(t, err)

		err = tx.PayTo("1234567", 100)
		assert.Error(t, err)
	})

	t.Run("valid pay to address", func(t *testing.T) {
		tx := bt.NewTx()
		assert.NotNil(t, tx)
		err := tx.From(
			"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
			0,
			"76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
			4000000)
		assert.NoError(t, err)

		err = tx.PayTo("1GHMW7ABrFma2NSwiVe9b9bZxkMB7tuPZi", 100)
		assert.NoError(t, err)
		assert.Equal(t, 1, tx.OutputCount())
	})
}

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

		err = tx.ChangeToAddress("1GHMW7ABrFma2NSwiVe9b9bZxkMB7tuPZi", bt.DefaultFees())
		assert.NoError(t, err)

		assert.Equal(t, 1, tx.OutputCount())
		assert.Equal(t, "76a914a7a1a7fd7d279b57b84e596cbbf82608efdb441a88ac", tx.Outputs[0].LockingScript.ToString())
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

		inputs := tx.GetInputs()
		assert.Equal(t, 1, len(inputs))
		assert.Equal(t, "07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b", inputs[0].PreviousTxID)
		assert.Equal(t, uint32(0), inputs[0].PreviousTxOutIndex)
		assert.Equal(t, uint64(4000000), inputs[0].PreviousTxSatoshis)
		assert.Equal(t, bt.DefaultSequenceNumber, inputs[0].SequenceNumber)
		assert.Equal(t, "76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac", inputs[0].PreviousTxScript.ToString())
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

		err = tx.ChangeToAddress("mwV3YgnowbJJB3LcyCuqiKpdivvNNFiK7M", bt.DefaultFees())
		assert.NoError(t, err)

		var wif *bsvutil.WIF
		wif, err = bsvutil.DecodeWIF("L3MhnEn1pLWcggeYLk9jdkvA2wUK1iWwwrGkBbgQRqv6HPCdRxuw")
		assert.NoError(t, err)
		assert.NotNil(t, wif)

		_, err = tx.SignAuto(&bt.InternalSigner{PrivateKey: wif.PrivKey, SigHashFlag: 0})
		assert.NoError(t, err)

		assert.Equal(t, expectedTx.ToString(), tx.ToString())
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

		err = tx.ChangeToAddress("mwV3YgnowbJJB3LcyCuqiKpdivvNNFiK7M", bt.DefaultFees())
		assert.NoError(t, err)

		var wif *bsvutil.WIF
		wif, err = bsvutil.DecodeWIF("L3MhnEn1pLWcggeYLk9jdkvA2wUK1iWwwrGkBbgQRqv6HPCdRxuw")
		assert.NoError(t, err)
		assert.NotNil(t, wif)

		_, err = tx.SignAuto(&bt.InternalSigner{PrivateKey: wif.PrivKey, SigHashFlag: 0})
		assert.NoError(t, err)

		// Correct fee for the tx
		assert.Equal(t, uint64(3999904), tx.Outputs[0].Satoshis)

		// Correct script hex string
		assert.Equal(t,
			"76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
			tx.Outputs[0].GetLockingScriptHexString(),
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
		err = tx.PayTo("1C8bzHM8XFBHZ2ZZVvFy2NSoAZbwCXAicL", 500)
		assert.NoError(t, err)

		// add some op return
		var outPut *bt.Output
		outPut, err = bt.NewOpReturnPartsOutput([][]byte{[]byte("hi"), []byte("how"), []byte("are"), []byte("you")})
		assert.NoError(t, err)
		assert.NotNil(t, outPut)
		tx.AddOutput(outPut)

		err = tx.ChangeToAddress("1D7gaZJo3vPn2Ks3PH694W9P8UVYLNh2jY", bt.DefaultFees())
		assert.NoError(t, err)

		var wif *bsvutil.WIF
		wif, err = bsvutil.DecodeWIF("L3MhnEn1pLWcggeYLk9jdkvA2wUK1iWwwrGkBbgQRqv6HPCdRxuw")
		assert.NoError(t, err)
		assert.NotNil(t, wif)

		_, err = tx.SignAuto(&bt.InternalSigner{PrivateKey: wif.PrivKey, SigHashFlag: 0})
		assert.NoError(t, err)

		assert.Equal(t,
			"0100000001760595866e99c1ce920197844740f5598b34763878696371d41b3a7c0a65b0b7000000006b483045022100c9f6e9c809885705221e019eeddd36be5b0472e42bb422a11152da7d7edf724902201679640f362859f1fd15db4d104acaf31caea6b15a6bc57e2bfc7a6af1be2d99412102c8803fdd437d902f08e3c2344cb33065c99d7c99982018ff9f7219c3dd352ff0ffffffff03f4010000000000001976a9147a1980655efbfec416b2b0c663a7b3ac0b6a25d288ac000000000000000011006a02686903686f770361726503796f7576010000000000001976a91484e50b300b009833b297dc671817c79b5459da1d88ac00000000",
			tx.ToString(),
		)

		feePaid := tx.GetTotalInputSatoshis() - tx.GetTotalOutputSatoshis()
		assert.Equal(t, uint64(126), feePaid)

		txSize := len(tx.ToBytes())
		assert.Equal(t, 252, txSize)

		feeRate := float64(feePaid) / float64(txSize)
		assert.Equal(t, 0.5, feeRate)
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

		err = tx.ChangeToAddress("mwV3YgnowbJJB3LcyCuqiKpdivvNNFiK7M", bt.DefaultFees())
		assert.NoError(t, err)

		var wif *bsvutil.WIF
		wif, err = bsvutil.DecodeWIF("L3MhnEn1pLWcggeYLk9jdkvA2wUK1iWwwrGkBbgQRqv6HPCdRxuw")
		assert.NoError(t, err)
		assert.NotNil(t, wif)

		_, err = tx.SignAuto(&bt.InternalSigner{PrivateKey: wif.PrivKey, SigHashFlag: 0})
		assert.NoError(t, err)

		assert.Equal(t, "01000000010b94a1ef0fb352aa2adc54207ce47ba55d5a1c1609afda58fe9520e472299107000000006a473044022049ee0c0f26c00e6a6b3af5990fc8296c66eab3e3e42ab075069b89b1be6fefec02206079e49dd8c9e1117ef06fbe99714d822620b1f0f5d19f32a1128f5d29b7c3c4412102c8803fdd437d902f08e3c2344cb33065c99d7c99982018ff9f7219c3dd352ff0ffffffff01a0083d00000000001976a914af2590a45ae401651fdbdf59a76ad43d1862534088ac00000000", tx.ToString())

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

		err = tx.PayTo("mxAoAyZFXX6LZBWhoam3vjm6xt9NxPQ15f", 1000000)
		assert.NoError(t, err)

		err = tx.PayTo("mxAoAyZFXX6LZBWhoam3vjm6xt9NxPQ15f", 3000000)
		assert.NoError(t, err)

		err = tx.ChangeToAddress("mwV3YgnowbJJB3LcyCuqiKpdivvNNFiK7M", bt.DefaultFees())
		assert.NoError(t, err)

		var wif *bsvutil.WIF
		wif, err = bsvutil.DecodeWIF("L3MhnEn1pLWcggeYLk9jdkvA2wUK1iWwwrGkBbgQRqv6HPCdRxuw")
		assert.NoError(t, err)
		assert.NotNil(t, wif)

		_, err = tx.SignAuto(&bt.InternalSigner{PrivateKey: wif.PrivKey, SigHashFlag: 0})
		assert.NoError(t, err)

		assert.Equal(t, "01000000010b94a1ef0fb352aa2adc54207ce47ba55d5a1c1609afda58fe9520e472299107000000006a47304402206bbb4b23349bdf86e6fbc9067226e9a7b15c977fa530999b39cd0a6d9c83360d02202dd8ffdc610e58b3fc92b44400d99e38c78866765f31acb40d98007a52e7a826412102c8803fdd437d902f08e3c2344cb33065c99d7c99982018ff9f7219c3dd352ff0ffffffff0240420f00000000001976a914b6aa34534d2b11e66b438c7525f819aee01e397c88acc0c62d00000000001976a914b6aa34534d2b11e66b438c7525f819aee01e397c88ac00000000", tx.ToString())

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

		err = tx.PayTo("mxAoAyZFXX6LZBWhoam3vjm6xt9NxPQ15f", 999995)
		assert.NoError(t, err)

		err = tx.PayTo("mxAoAyZFXX6LZBWhoam3vjm6xt9NxPQ15f", 3000000)
		assert.NoError(t, err)

		err = tx.ChangeToAddress("mwV3YgnowbJJB3LcyCuqiKpdivvNNFiK7M", bt.DefaultFees())
		assert.NoError(t, err)

		var wif *bsvutil.WIF
		wif, err = bsvutil.DecodeWIF("L3MhnEn1pLWcggeYLk9jdkvA2wUK1iWwwrGkBbgQRqv6HPCdRxuw")
		assert.NoError(t, err)
		assert.NotNil(t, wif)

		_, err = tx.SignAuto(&bt.InternalSigner{PrivateKey: wif.PrivKey, SigHashFlag: 0})
		assert.NoError(t, err)

		assert.Equal(t, "01000000010b94a1ef0fb352aa2adc54207ce47ba55d5a1c1609afda58fe9520e472299107000000006b483045022100fd07316603e9abf393e695192e8ce1e7f808d2735cc57039109a2210ad32d9a7022000e301e2a988b23ab3872b041df8b6eb0315238e0918944cbaf8b6abdde75cac412102c8803fdd437d902f08e3c2344cb33065c99d7c99982018ff9f7219c3dd352ff0ffffffff023b420f00000000001976a914b6aa34534d2b11e66b438c7525f819aee01e397c88acc0c62d00000000001976a914b6aa34534d2b11e66b438c7525f819aee01e397c88ac00000000", tx.ToString())

		// todo: expected the pay-to inputs to change based on the fee :P

		assert.Equal(t, uint64(999995), tx.Outputs[0].Satoshis)
		assert.Equal(t, uint64(3000000), tx.Outputs[1].Satoshis)
	})

	t.Run("multiple inputs, spend all", func(t *testing.T) {
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

		err = tx.ChangeToAddress("1BxGFoRPSFgYxoAStEncL6HuELqPkV3JVj", bt.DefaultFees())
		assert.NoError(t, err)

		var wif *bsvutil.WIF
		wif, err = bsvutil.DecodeWIF("5JXAjNX7cbiWvmkdnj1EnTKPChauttKAJibXLm8tqWtDhXrRbKz")
		assert.NoError(t, err)
		assert.NotNil(t, wif)

		is, err := tx.SignAuto(&bt.InternalSigner{PrivateKey: wif.PrivKey, SigHashFlag: 0})
		assert.NoError(t, err)

		assert.ElementsMatch(t, []int{0, 1}, is)
		assert.Equal(t, 2, len(is))

		assert.Equal(t, "01000000028ee20a442cdbcc9f9f927d9c2c9370e611675ebc24c064e8e94508ec8eca889e000000006b483045022100fa52a44cd8010ba646a8df6bac6e5e8aa93f24439521c2ce1c8fe6550e73c1750220636e30d757702a6777d8310090962d4bac2b3fd634127856d51b184f5c702c8f4121034aaeabc056f33fd960d1e43fc8a0672723af02f275e54c31381af66a334634caffffffff42eaf7bdddc797a0beb97717ff8846f03c963fb5fe15a2b555b9cbd477b0254e000000006b483045022100c201fd55ef33525b3eb0557fac77408b8ec7f6ea5b00d08512df105172f992d60220753b21519a416dcbeaf1a501d9c36de2aea9c83c6d258320500371819d0758e14121034aaeabc056f33fd960d1e43fc8a0672723af02f275e54c31381af66a334634caffffffff01c62b0000000000001976a9147824dec00be2c45dad83c9b5e9f5d7ef05ba3cf988ac00000000", tx.ToString())
	})
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

		var out *bt.Output
		out, err = bt.NewOpReturnPartsOutput(ops)
		assert.NoError(t, err)

		tx.AddOutput(out)

		var wif *bsvutil.WIF
		wif, err = bsvutil.DecodeWIF("KznvCNc6Yf4iztSThoMH6oHWzH9EgjfodKxmeuUGPq5DEX5maspS")
		assert.NoError(t, err)
		assert.NotNil(t, wif)

		_, err = tx.SignAuto(&bt.InternalSigner{PrivateKey: wif.PrivKey, SigHashFlag: 0})
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

		var wif *bsvutil.WIF
		wif, err = bsvutil.DecodeWIF("KznvCNc6Yf4iztSThoMH6oHWzH9EgjfodKxmeuUGPq5DEX5maspS")
		assert.NoError(t, err)
		assert.NotNil(t, wif)

		_, err = tx.SignAuto(&bt.InternalSigner{PrivateKey: wif.PrivKey, SigHashFlag: 0})
		assert.NoError(t, err)

		assert.Equal(t, false, tx.HasDataOutputs())
	})
}

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

		_, err = tx.SignAuto(&bt.InternalSigner{PrivateKey: wif.PrivKey, SigHashFlag: 0})
		assert.NoError(t, err)

		assert.NotEqual(t, rawTxBefore, tx.ToString())
	})

	t.Run("no input or output", func(t *testing.T) {
		tx := bt.NewTx()
		assert.NotNil(t, tx)

		rawTxBefore := tx.ToString()

		_, err := tx.SignAuto(&bt.InternalSigner{PrivateKey: nil, SigHashFlag: 0})
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
		_, err = tx.SignAuto(&bt.InternalSigner{PrivateKey: wif.PrivKey, SigHashFlag: 0})
		assert.NoError(t, err)
		assert.Equal(t, rawTxBefore, tx.ToString())
	})
}
