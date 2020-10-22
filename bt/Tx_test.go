package bt_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/bitcoinsv/bsvutil"
	"github.com/libsv/libsv/bt"
	"github.com/libsv/libsv/bt/fees"
	"github.com/libsv/libsv/bt/input"
	"github.com/libsv/libsv/bt/output"
	"github.com/libsv/libsv/bt/sig"

	"github.com/libsv/libsv/script"
)

func TestNew(t *testing.T) {
	bt := bt.New()

	// check version
	if bt.Version != 1 {
		t.Errorf("Expcted version be %v, but got %v", 1, bt.Version)
	}

	//	check locktime
	if bt.Locktime != 0 {
		t.Errorf("Expcted locktime be %v, but got %v", 2, bt.Locktime)
	}
}

func TestNewFromString(t *testing.T) {
	h := "02000000011ccba787d421b98904da3329b2c7336f368b62e89bc896019b5eadaa28145b9c000000004847304402205cc711985ce2a6d61eece4f9b6edd6337bad3b7eca3aa3ce59bc15620d8de2a80220410c92c48a226ba7d5a9a01105524097f673f31320d46c3b61d2378e6f05320041ffffffff01c0aff629010000001976a91418392a59fc1f76ad6a3c7ffcea20cfcb17bda9eb88ac00000000"
	bt, err := bt.NewFromString(h)
	if err != nil {
		t.Error(err)
		return
	}

	// check version
	if bt.Version != 2 {
		t.Errorf("Expcted version be %v, but got %v", 2, bt.Version)
	}

	//	check locktime
	if bt.Locktime != 0 {
		t.Errorf("Expcted locktime be %v, but got %v", 2, bt.Locktime)
	}

	//	 check input
	inputLen := len(bt.Inputs)
	if inputLen != 1 {
		t.Errorf("Expcted input be %v, but got %v", 1, inputLen)
	}

	i := input.Input{}
	i.PreviousTxID = "9c5b1428aaad5e9b0196c89be8628b366f33c7b22933da0489b921d487a7cb1c"
	i.PreviousTxOutIndex = 0
	i.SequenceNumber = uint32(0xffffffff)
	i.UnlockingScript, _ = script.NewFromHexString("47304402205cc711985ce2a6d61eece4f9b6edd6337bad3b7eca3aa3ce59bc15620d8de2a80220410c92c48a226ba7d5a9a01105524097f673f31320d46c3b61d2378e6f05320041")
	sameInput := reflect.DeepEqual(*bt.Inputs[0], i)
	if !sameInput {
		t.Errorf("Input did not match")
	}

	//	 check output
	outputLen := len(bt.Outputs)
	if outputLen != 1 {
		t.Errorf("Expcted output be %v, but got %v", 1, outputLen)
	}
	ls, _ := script.NewFromHexString("76a91418392a59fc1f76ad6a3c7ffcea20cfcb17bda9eb88ac")
	o := output.Output{
		Satoshis:      4999000000,
		LockingScript: ls,
	}
	sameOutput := reflect.DeepEqual(*bt.Outputs[0], o)
	if !sameOutput {
		t.Errorf("Output did not match")
	}

}
func TestToBytes(t *testing.T) {
	h := "02000000011ccba787d421b98904da3329b2c7336f368b62e89bc896019b5eadaa28145b9c0000000049483045022100c4df63202a9aa2bea5c24ebf4418d145e81712072ef744a4b108174f1ef59218022006eb54cf904707b51625f521f8ed2226f7d34b62492ebe4ddcb1c639caf16c3c41ffffffff0140420f00000000001976a91418392a59fc1f76ad6a3c7ffcea20cfcb17bda9eb88ac00000000"
	bt, err := bt.NewFromString(h)
	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("%s", bt.ToString())
	t.Logf("%x", bt.ToBytes())

}

func TestTxID(t *testing.T) {
	tx, err := bt.NewFromString("010000000193a35408b6068499e0d5abd799d3e827d9bfe70c9b75ebe209c91d2507232651000000006b483045022100c1d77036dc6cd1f3fa1214b0688391ab7f7a16cd31ea4e5a1f7a415ef167df820220751aced6d24649fa235132f1e6969e163b9400f80043a72879237dab4a1190ad412103b8b40a84123121d260f5c109bc5a46ec819c2e4002e5ba08638783bfb4e01435ffffffff02404b4c00000000001976a91404ff367be719efa79d76e4416ffb072cd53b208888acde94a905000000001976a91404d03f746652cfcb6cb55119ab473a045137d26588ac00000000")
	if err != nil {
		t.Error(err)
	}

	id := tx.GetTxID()
	expected := "19dcf16ecc9286c3734fdae3d45d4fc4eb6b25f841131e06460f4939bba0026e"

	if expected != id {
		t.Errorf("Bad TXID")
	}
}

func TestGetTotalOutputSatoshis(t *testing.T) {
	tx, err := bt.NewFromString("020000000180f1ada3ad8e861441d9ceab40b68ed98f13695b185cc516226a46697cc01f80010000006b483045022100fa3a0f8fa9fbf09c372b7a318fa6175d022c1d782f7b8bc5949a7c8f59ce3f35022005e0e84c26f26d892b484ff738d803a57626679389c8b302939460dab29a5308412103e46b62eea5db5898fb65f7dc840e8a1dbd8f08a19781a23f1f55914f9bedcd49feffffff02dec537b2000000001976a914ba11bcc46ecf8d88e0828ddbe87997bf759ca85988ac00943577000000001976a91418392a59fc1f76ad6a3c7ffcea20cfcb17bda9eb88ac6e000000")
	if err != nil {
		t.Error(err)
	}

	total := tx.GetTotalOutputSatoshis()
	expected := (29.89999582 + 20.00) * 1e8

	if uint64(expected) != total {
		t.Errorf("Expected %d, got %d", uint64(expected), total)
	}
}

func TestRegTestCoinbase(t *testing.T) {
	h := "02000000010000000000000000000000000000000000000000000000000000000000000000ffffffff0e5101010a2f4542323030302e302fffffffff0100f2052a01000000232103db233bb9fc387d78b133ec904069d46e95ff17da657671b44afa0bc64e89ac18ac00000000"
	bt, err := bt.NewFromString(h)
	if err != nil {
		t.Error(err)
		return
	}
	// check if coinbase transaction
	if !bt.IsCoinbase() {
		t.Error("Tx is not Coinbase transaction")
	}
	// check input count
	expectedInputCount := bt.InputCount()
	if expectedInputCount != 1 {
		t.Errorf("Expcted input count to be %v, but got %v", 1, expectedInputCount)
	}

}

func TestGetVersion(t *testing.T) {
	const tx = "01000000014c6ec863cf3e0284b407a1a1b8138c76f98280812cb9653231f385a0305fc76f010000006b483045022100f01c1a1679c9437398d691c8497f278fa2d615efc05115688bf2c3335b45c88602201b54437e54fb53bc50545de44ea8c64e9e583952771fcc663c8687dc2638f7854121037e87bbd3b680748a74372640628a8f32d3a841ceeef6f75626ab030c1a04824fffffffff021d784500000000001976a914e9b62e25d4c6f97287dfe62f8063b79a9638c84688ac60d64f00000000001976a914bb4bca2306df66d72c6e44a470873484d8808b8888ac00000000"
	bt, err := bt.NewFromString(tx)
	if err != nil {
		t.Error(err)
		return
	}

	res := bt.Version
	if res != 1 {
		t.Errorf("Expecting 1, got %d", res)
	}
}

func TestIsCoinbase(t *testing.T) {
	const coinbase = "01000000010000000000000000000000000000000000000000000000000000000000000000ffffffff4303bfea07322f53696d6f6e204f726469736820616e642053747561727420467265656d616e206d61646520746869732068617070656e2f9a46434790f7dbdea3430000ffffffff018a08ac4a000000001976a9148bf10d323ac757268eb715e613cb8e8e1d1793aa88ac00000000"
	bt1, err := bt.NewFromString(coinbase)
	if err != nil {
		t.Error(err)
		return
	}

	cb1 := bt1.IsCoinbase()
	if cb1 == false {
		t.Errorf("Expecting true, got %t", cb1)
	}

	const tx = "01000000014c6ec863cf3e0284b407a1a1b8138c76f98280812cb9653231f385a0305fc76f010000006b483045022100f01c1a1679c9437398d691c8497f278fa2d615efc05115688bf2c3335b45c88602201b54437e54fb53bc50545de44ea8c64e9e583952771fcc663c8687dc2638f7854121037e87bbd3b680748a74372640628a8f32d3a841ceeef6f75626ab030c1a04824fffffffff021d784500000000001976a914e9b62e25d4c6f97287dfe62f8063b79a9638c84688ac60d64f00000000001976a914bb4bca2306df66d72c6e44a470873484d8808b8888ac00000000"
	bt2, err := bt.NewFromString(tx)
	if err != nil {
		t.Error(err)
		return
	}

	cb2 := bt2.IsCoinbase()
	if cb2 == true {
		t.Errorf("Expecting false, got %t", cb2)
	}
}

func TestCreateTx(t *testing.T) {
	tx := bt.New()

	tx.From(
		"3c8edde27cb9a9132c22038dac4391496be9db16fd21351565cc1006966fdad5",
		0,
		"76a914eb0bd5edba389198e73f8efabddfc61666969ff788ac",
		2000000)

	tx.PayTo("n2wmGVP89x3DsLNqk3NvctfQy9m9pvt7mk", 1999942)

	wif, _ := bsvutil.DecodeWIF("KznvCNc6Yf4iztSThoMH6oHWzH9EgjfodKxmeuUGPq5DEX5maspS")

	signer := sig.InternalSigner{PrivateKey: wif.PrivKey, SigHashFlag: 0}
	err := tx.SignAuto(&signer)
	if err != nil {
		t.Errorf(err.Error())
	}

	fmt.Println(tx.ToString())
}

func TestChange(t *testing.T) {
	expectedTx, _ := bt.NewFromString("01000000010b94a1ef0fb352aa2adc54207ce47ba55d5a1c1609afda58fe9520e472299107000000006a473044022049ee0c0f26c00e6a6b3af5990fc8296c66eab3e3e42ab075069b89b1be6fefec02206079e49dd8c9e1117ef06fbe99714d822620b1f0f5d19f32a1128f5d29b7c3c4412102c8803fdd437d902f08e3c2344cb33065c99d7c99982018ff9f7219c3dd352ff0ffffffff01a0083d00000000001976a914af2590a45ae401651fdbdf59a76ad43d1862534088ac00000000")

	tx := bt.New()

	tx.From(
		"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
		0,
		"76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
		4000000)

	tx.ChangeToAddress("mwV3YgnowbJJB3LcyCuqiKpdivvNNFiK7M", fees.Default())

	wif, _ := bsvutil.DecodeWIF("L3MhnEn1pLWcggeYLk9jdkvA2wUK1iWwwrGkBbgQRqv6HPCdRxuw")

	signer := sig.InternalSigner{PrivateKey: wif.PrivKey, SigHashFlag: 0}
	err := tx.SignAuto(&signer)
	if err != nil {
		t.Errorf(err.Error())
	}

	if tx.ToString() != expectedTx.ToString() {
		t.Errorf("Expected %s, got %s", tx.ToString(), expectedTx.ToString())
	}

}
