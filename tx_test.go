package bt_test

import (
	"reflect"
	"testing"

	"github.com/bitcoinsv/bsvutil"
	"github.com/libsv/go-bt"
	"github.com/libsv/go-bt/bscript"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	tx := bt.NewTx()
	assert.NotNil(t, tx)

	// check version
	assert.Equal(t, uint32(1), tx.Version)
	assert.Equal(t, uint32(0), tx.Locktime)
}

func TestNewFromString(t *testing.T) {
	h := "02000000011ccba787d421b98904da3329b2c7336f368b62e89bc896019b5eadaa28145b9c000000004847304402205cc711985ce2a6d61eece4f9b6edd6337bad3b7eca3aa3ce59bc15620d8de2a80220410c92c48a226ba7d5a9a01105524097f673f31320d46c3b61d2378e6f05320041ffffffff01c0aff629010000001976a91418392a59fc1f76ad6a3c7ffcea20cfcb17bda9eb88ac00000000"
	tx, err := bt.NewTxFromString(h)
	assert.NoError(t, err)
	assert.NotNil(t, tx)

	// Check version, locktime, inputs
	assert.Equal(t, uint32(2), tx.Version)
	assert.Equal(t, uint32(0), tx.Locktime)
	assert.Equal(t, 1, len(tx.Inputs))

	// Create a new unlocking script
	i := bt.Input{}
	i.PreviousTxID = "9c5b1428aaad5e9b0196c89be8628b366f33c7b22933da0489b921d487a7cb1c"
	i.PreviousTxOutIndex = 0
	i.SequenceNumber = uint32(0xffffffff)
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
}

func TestToBytes(t *testing.T) {
	h := "02000000011ccba787d421b98904da3329b2c7336f368b62e89bc896019b5eadaa28145b9c0000000049483045022100c4df63202a9aa2bea5c24ebf4418d145e81712072ef744a4b108174f1ef59218022006eb54cf904707b51625f521f8ed2226f7d34b62492ebe4ddcb1c639caf16c3c41ffffffff0140420f00000000001976a91418392a59fc1f76ad6a3c7ffcea20cfcb17bda9eb88ac00000000"
	tx, err := bt.NewTxFromString(h)
	assert.NoError(t, err)
	assert.NotNil(t, tx)
}

func TestTxID(t *testing.T) {
	tx, err := bt.NewTxFromString("010000000193a35408b6068499e0d5abd799d3e827d9bfe70c9b75ebe209c91d2507232651000000006b483045022100c1d77036dc6cd1f3fa1214b0688391ab7f7a16cd31ea4e5a1f7a415ef167df820220751aced6d24649fa235132f1e6969e163b9400f80043a72879237dab4a1190ad412103b8b40a84123121d260f5c109bc5a46ec819c2e4002e5ba08638783bfb4e01435ffffffff02404b4c00000000001976a91404ff367be719efa79d76e4416ffb072cd53b208888acde94a905000000001976a91404d03f746652cfcb6cb55119ab473a045137d26588ac00000000")
	assert.NoError(t, err)
	assert.NotNil(t, tx)
	assert.Equal(t, "19dcf16ecc9286c3734fdae3d45d4fc4eb6b25f841131e06460f4939bba0026e", tx.GetTxID())
}

func TestGetTotalOutputSatoshis(t *testing.T) {
	tx, err := bt.NewTxFromString("020000000180f1ada3ad8e861441d9ceab40b68ed98f13695b185cc516226a46697cc01f80010000006b483045022100fa3a0f8fa9fbf09c372b7a318fa6175d022c1d782f7b8bc5949a7c8f59ce3f35022005e0e84c26f26d892b484ff738d803a57626679389c8b302939460dab29a5308412103e46b62eea5db5898fb65f7dc840e8a1dbd8f08a19781a23f1f55914f9bedcd49feffffff02dec537b2000000001976a914ba11bcc46ecf8d88e0828ddbe87997bf759ca85988ac00943577000000001976a91418392a59fc1f76ad6a3c7ffcea20cfcb17bda9eb88ac6e000000")
	assert.NoError(t, err)
	assert.NotNil(t, tx)
	assert.Equal(t, uint64((29.89999582+20.00)*1e8), tx.GetTotalOutputSatoshis())
}

func TestRegTestCoinbase(t *testing.T) {
	rawTx := "02000000010000000000000000000000000000000000000000000000000000000000000000ffffffff0e5101010a2f4542323030302e302fffffffff0100f2052a01000000232103db233bb9fc387d78b133ec904069d46e95ff17da657671b44afa0bc64e89ac18ac00000000"
	tx, err := bt.NewTxFromString(rawTx)
	assert.NoError(t, err)
	assert.NotNil(t, tx)

	// Check if coinbase transaction
	assert.Equal(t, true, tx.IsCoinbase())

	// Check input count
	assert.Equal(t, 1, tx.InputCount())
}

func TestGetVersion(t *testing.T) {
	const rawTx = "01000000014c6ec863cf3e0284b407a1a1b8138c76f98280812cb9653231f385a0305fc76f010000006b483045022100f01c1a1679c9437398d691c8497f278fa2d615efc05115688bf2c3335b45c88602201b54437e54fb53bc50545de44ea8c64e9e583952771fcc663c8687dc2638f7854121037e87bbd3b680748a74372640628a8f32d3a841ceeef6f75626ab030c1a04824fffffffff021d784500000000001976a914e9b62e25d4c6f97287dfe62f8063b79a9638c84688ac60d64f00000000001976a914bb4bca2306df66d72c6e44a470873484d8808b8888ac00000000"
	tx, err := bt.NewTxFromString(rawTx)
	assert.NoError(t, err)
	assert.NotNil(t, tx)
	assert.Equal(t, uint32(1), tx.Version)
}

func TestIsCoinbase(t *testing.T) {
	const coinbase = "01000000010000000000000000000000000000000000000000000000000000000000000000ffffffff4303bfea07322f53696d6f6e204f726469736820616e642053747561727420467265656d616e206d61646520746869732068617070656e2f9a46434790f7dbdea3430000ffffffff018a08ac4a000000001976a9148bf10d323ac757268eb715e613cb8e8e1d1793aa88ac00000000"
	bt1, err := bt.NewTxFromString(coinbase)
	assert.NoError(t, err)
	assert.NotNil(t, bt1)

	assert.Equal(t, true, bt1.IsCoinbase())

	const tx = "01000000014c6ec863cf3e0284b407a1a1b8138c76f98280812cb9653231f385a0305fc76f010000006b483045022100f01c1a1679c9437398d691c8497f278fa2d615efc05115688bf2c3335b45c88602201b54437e54fb53bc50545de44ea8c64e9e583952771fcc663c8687dc2638f7854121037e87bbd3b680748a74372640628a8f32d3a841ceeef6f75626ab030c1a04824fffffffff021d784500000000001976a914e9b62e25d4c6f97287dfe62f8063b79a9638c84688ac60d64f00000000001976a914bb4bca2306df66d72c6e44a470873484d8808b8888ac00000000"

	var bt2 *bt.Tx
	bt2, err = bt.NewTxFromString(tx)
	assert.NoError(t, err)
	assert.NotNil(t, bt2)

	assert.Equal(t, false, bt2.IsCoinbase())
}

func TestCreateTx(t *testing.T) {
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

	signer := bt.InternalSigner{PrivateKey: wif.PrivKey, SigHashFlag: 0}
	err = tx.SignAuto(&signer)
	assert.NoError(t, err)
}

func TestChange(t *testing.T) {
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

	err = tx.ChangeToAddress("mwV3YgnowbJJB3LcyCuqiKpdivvNNFiK7M", bt.Default())
	assert.NoError(t, err)

	var wif *bsvutil.WIF
	wif, err = bsvutil.DecodeWIF("L3MhnEn1pLWcggeYLk9jdkvA2wUK1iWwwrGkBbgQRqv6HPCdRxuw")
	assert.NoError(t, err)
	assert.NotNil(t, wif)

	signer := bt.InternalSigner{PrivateKey: wif.PrivKey, SigHashFlag: 0}
	err = tx.SignAuto(&signer)
	assert.NoError(t, err)

	assert.Equal(t, expectedTx.ToString(), tx.ToString())
}
