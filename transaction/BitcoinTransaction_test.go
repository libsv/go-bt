package transaction

import (
	"encoding/hex"
	"fmt"
	"testing"

	"bitbucket.org/simon_ordish/cryptolib"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

func TestRegTestCoinbase(t *testing.T) {
	hex := "02000000010000000000000000000000000000000000000000000000000000000000000000ffffffff0e5101010a2f4542323030302e302fffffffff0100f2052a01000000232103db233bb9fc387d78b133ec904069d46e95ff17da657671b44afa0bc64e89ac18ac00000000"
	bt, err := NewFromString(hex)
	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("Iscoinbase %t", bt.IsCoinbase())
	t.Logf("input count %d", bt.InputCount())

}

func TestRegtestUnsigned(t *testing.T) {
	unsigned := "0200000001037ded84940e54c8c9e1ba73aa338a61d2ee4c4ac0d1faf2e8671896b0f8da630000000000ffffffff01806de729010000001976a91463ea0d776d45502d2226aed9ebdf5b676e232ca188ac00000000"

	bt, err := NewFromString(unsigned)
	if err != nil {
		t.Error(err)
		return
	}

	// Taking this unsigned transaction, we want to generate a signature of the tx that equals:
	// scriptSig := "4730440220416154a5a117e89855397c6a7b2796d82107d20c1326bc917444e4ab84567b80022057de212dc0615ea1f4bbca817ed18be49acb96acff7760fc4d6447cbe772d1e8412103fc7c702eb7a03099ef01970b31ecbebe7ff77adc202d3749a8562ffc185a44a6"

	scriptSig := []byte{}

	bt.GetInputs()[0].script = NewScript(scriptSig)

	t.Logf("%x", bt.Hex())

	// signed := "0200000001c78e6fda3658d39192d72aeb6aca80ff07cb1e41f375de8b4af850a03b7d8419000000006b483045022100b2d0657263ce1ece216b4411b597eb856b07d0e1e99082e4d4be00e0637411ab022044d090a2c0c0aa14517920bae94d1fed870ded61ac57f8dfc96d22408f67c51941210288e78dc896da65d8a96f8f7a16b2ae87378597b317931bfc1ccd89c88703c467ffffffff01806de729010000001976a914003ebbc2b6383e864b38abadad712e4e5add4fef88ac00000000"

	// xtxoTXID := "19847d3ba050f84a8bde75f3411ecb07ff80ca6aeb2ad79291d35836da6f8ec7"
	// utxoAddress := "mi1Mh7ENBnum1CnDAESXfCwikA2shwtdNN"
	// utxoPrivateKey := "cPjqbeH84Qq9VmWrURUEJNo7DaKnrPP428utXzZRcbBdXPx7kGe5"
	// utxoPublicKey := "0288e78dc896da65d8a96f8f7a16b2ae87378597b317931bfc1ccd89c88703c467"

	// pubKeyHash := "1b4f6e032a4da3b75fa685475ccfce51b2ad707e"

	// wif, err := btcutil.DecodeWIF(utxoPrivateKey)
	// if err != nil {
	// 	t.Error(err)
	// 	return
	// }
	// t.Logf("%x", wif.PrivKey.Serialize())

	// Get the publicKey from the private
	// privKey, publicKey := btcec.PrivKeyFromBytes(btcec.S256(), wif.PrivKey.Serialize())
	// t.Logf("%x", publicKey.SerializeCompressed())

	// t.Log(bt.HexWithClearedInputs(0, nil))

	// bt.Sign(privKey, 0) // 03ececf2d12a7f614aef4c82ecf13c303bd9975d
}

func TestGetVersion(t *testing.T) {
	const tx = "01000000014c6ec863cf3e0284b407a1a1b8138c76f98280812cb9653231f385a0305fc76f010000006b483045022100f01c1a1679c9437398d691c8497f278fa2d615efc05115688bf2c3335b45c88602201b54437e54fb53bc50545de44ea8c64e9e583952771fcc663c8687dc2638f7854121037e87bbd3b680748a74372640628a8f32d3a841ceeef6f75626ab030c1a04824fffffffff021d784500000000001976a914e9b62e25d4c6f97287dfe62f8063b79a9638c84688ac60d64f00000000001976a914bb4bca2306df66d72c6e44a470873484d8808b8888ac00000000"
	bt, err := NewFromString(tx)
	if err != nil {
		t.Error(err)
		return
	}

	res := bt.Version
	if res != 1 {
		t.Errorf("Expecting 1, got %d", res)
	}
}

func TestConvertXPriv(t *testing.T) {
	const xprv = "xprv9s21ZrQH143K2beTKhLXFRWWFwH8jkwUssjk3SVTiApgmge7kNC3jhVc4NgHW8PhW2y7BCDErqnKpKuyQMjqSePPJooPJowAz5BVLThsv6c"
	const expected = "5f86e4023a4e94f00463f81b70ff951f83f896a0a3e6ed89cf163c152f954f8b"

	r, _ := cryptolib.NewPrivateKey(xprv)

	t.Logf("%x", r.PrivateKey)
}

func TestSignRedeemScript(t *testing.T) {
	var redeemScript, _ = hex.DecodeString("524c53ff0488b21e000000000000000000362f7a9030543db8751401c387d6a71e870f1895b3a62569d455e8ee5f5f5e5f03036624c6df96984db6b4e625b6707c017eb0e0d137cd13a0c989bfa77a4473fd000000004c53ff0488b21e0000000000000000008b20425398995f3c866ea6ce5c1828a516b007379cf97b136bffbdc86f75df14036454bad23b019eae34f10aff8b8d6d8deb18cb31354e5a169ee09d8a4560e8250000000052ae")
	const expected = "3044022041682b268531cf6209577deae34b92fdc83d9ef6e3abc190d4952e927761efd502201696256fba4dd6b05e44ed3871abbd1bc11356aea5ddc36816ca779f68cca6fa"

	const xprv = "xprv9s21ZrQH143K2beTKhLXFRWWFwH8jkwUssjk3SVTiApgmge7kNC3jhVc4NgHW8PhW2y7BCDErqnKpKuyQMjqSePPJooPJowAz5BVLThsv6c"
	const privHex = "5f86e4023a4e94f00463f81b70ff951f83f896a0a3e6ed89cf163c152f954f8b"

	pkBytes, err := hex.DecodeString(privHex)
	if err != nil {
		fmt.Println(err)
		return
	}
	privKey, pubKey := btcec.PrivKeyFromBytes(btcec.S256(), pkBytes)

	// Sign a message using the private key.
	messageHash := chainhash.DoubleHashB(redeemScript)
	signature, err := privKey.Sign(messageHash)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Serialize and display the signature.
	fmt.Printf("Serialized Signature: %x\n", signature.Serialize())

	// Verify the signature for the message using the public key.
	verified := signature.Verify(messageHash, pubKey)
	fmt.Printf("Signature Verified? %v\n", verified)
}

func TestIsCoinbase(t *testing.T) {
	const coinbase = "01000000010000000000000000000000000000000000000000000000000000000000000000ffffffff4303bfea07322f53696d6f6e204f726469736820616e642053747561727420467265656d616e206d61646520746869732068617070656e2f9a46434790f7dbdea3430000ffffffff018a08ac4a000000001976a9148bf10d323ac757268eb715e613cb8e8e1d1793aa88ac00000000"
	bt1, err := NewFromString(coinbase)
	if err != nil {
		t.Error(err)
		return
	}

	cb1 := bt1.IsCoinbase()
	if cb1 == false {
		t.Errorf("Expecting true, got %t", cb1)
	}

	const tx = "01000000014c6ec863cf3e0284b407a1a1b8138c76f98280812cb9653231f385a0305fc76f010000006b483045022100f01c1a1679c9437398d691c8497f278fa2d615efc05115688bf2c3335b45c88602201b54437e54fb53bc50545de44ea8c64e9e583952771fcc663c8687dc2638f7854121037e87bbd3b680748a74372640628a8f32d3a841ceeef6f75626ab030c1a04824fffffffff021d784500000000001976a914e9b62e25d4c6f97287dfe62f8063b79a9638c84688ac60d64f00000000001976a914bb4bca2306df66d72c6e44a470873484d8808b8888ac00000000"
	bt2, err := NewFromString(tx)
	if err != nil {
		t.Error(err)
		return
	}

	cb2 := bt2.IsCoinbase()
	if cb2 == true {
		t.Errorf("Expecting false, got %t", cb2)
	}
}

/*
48
30
45
02
21
00f4de422896e461da647b21d800a4ca9ace98dbd08c2dc9b8e049c93197c314f5
02
20
68836c3dfa6650ebeff73b1e3caa8761cd107ed13d6cc713856ebde3f874dd41
41

21
02aea77c449eeeef2746562e56ad053202755f9844276e3f0c684f9d59cdb9458d
ac OP_CHECKSIG

*/
// func TestMyTransaction(t *testing.T) {
// 	fromTx, err := NewFromString("02000000010000000000000000000000000000000000000000000000000000000000000000ffffffff0d510101092f45423132382e302fffffffff0100f2052a01000000232102aea77c449eeeef2746562e56ad053202755f9844276e3f0c684f9d59cdb9458dac00000000")
// 	if err != nil {
// 		t.Error(err)
// 		return
// 	}

// 	toTx, err := NewFromString("02000000019bb2dea27bcff46bca60e46ba2fdce706a8eb9d22c9b05e54166b8f9ac57d6de0000000049483045022100f4de422896e461da647b21d800a4ca9ace98dbd08c2dc9b8e049c93197c314f5022068836c3dfa6650ebeff73b1e3caa8761cd107ed13d6cc713856ebde3f874dd4141feffffff0200ca9a3b000000001976a9143c134f3ccd097be40242efd6fb370fc62501afe788ac00196bee000000001976a914c3d737cb0d93ded96a35d240aa3f01b34edc4e5d88ac65000000")
// 	if err != nil {
// 		t.Error(err)
// 		return
// 	}

// 	t.Errorf("%x\n%x\n", fromTx.GetOutputs()[0].GetOutputScript(), toTx.GetInputs()[0].GetInputScript())
// }
