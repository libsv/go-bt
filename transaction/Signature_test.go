package transaction

import (
	"encoding/hex"
	"testing"

	"bitbucket.org/simon_ordish/cryptolib"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcutil"
)

func TestSigHash(t *testing.T) {
	h := "0200000001037ded84940e54c8c9e1ba73aa338a61d2ee4c4ac0d1faf2e8671896b0f8da630000000000ffffffff01806de729010000001976a91463ea0d776d45502d2226aed9ebdf5b676e232ca188ac00000000"
	bt, err := NewFromString(h)
	if err != nil {
		t.Error(err)
		return
	}

	bt.Inputs[0].Script = NewScriptFromString("76a91403ececf2d12a7f614aef4c82ecf13c303bd9975d88ac")
	bt.Inputs[0].PreviousTxSatoshis = 4998000000
	// const utxo = bsv.Transaction.UnspentOutput({
	// 	txid: '63daf8b0961867e8f2fad1c04a4ceed2618a33aa73bae1c9c8540e9484ed7d03',
	// 	vout: 0,
	// 	scriptPubKey: '76a91403ececf2d12a7f614aef4c82ecf13c303bd9975d88ac',
	// 	amount: 49.98000000
	// })

	wif, err := btcutil.DecodeWIF("cPjqbeH84Qq9VmWrURUEJNo7DaKnrPP428utXzZRcbBdXPx7kGe5")
	if err != nil {
		t.Error(err)
		return
	}

	privKeys := make([]*btcec.PrivateKey, 0)
	privKeys = append(privKeys, wif.PrivKey)

	var sigtype uint32 = SighashAll | SighashForkID

	sigs, _ := GetSignatures(bt, privKeys, sigtype)
	for _, sig := range sigs {
		pubkey := wif.PrivKey.PubKey().SerializeCompressed()
		buf := make([]byte, 0)
		buf = append(buf, cryptolib.VarInt(uint64(len(sig.Signature)))...)
		buf = append(buf, sig.Signature...)
		buf = append(buf, cryptolib.VarInt(uint64(len(pubkey)))...)
		buf = append(buf, pubkey...)
		bt.GetInputs()[0].Script = NewScriptFromBytes(buf)
	}

	expected := "0200000001037ded84940e54c8c9e1ba73aa338a61d2ee4c4ac0d1faf2e8671896b0f8da63000000006a4730440220416154a5a117e89855397c6a7b2796d82107d20c1326bc917444e4ab84567b80022057de212dc0615ea1f4bbca817ed18be49acb96acff7760fc4d6447cbe772d1e8412103fc7c702eb7a03099ef01970b31ecbebe7ff77adc202d3749a8562ffc185a44a6ffffffff01806de729010000001976a91463ea0d776d45502d2226aed9ebdf5b676e232ca188ac00000000"
	if expected != hex.EncodeToString(bt.Hex()) {
		t.Errorf("Expected %q, got %q", expected, bt.Hex())
	}
}

func TestSigHashDave(t *testing.T) {

	// Unsigned TX generated in Moneybutton BSV library.
	unsignedTx := "010000000193a35408b6068499e0d5abd799d3e827d9bfe70c9b75ebe209c91d25072326510000000000ffffffff02404b4c00000000001976a91404ff367be719efa79d76e4416ffb072cd53b208888acde94a905000000001976a91404d03f746652cfcb6cb55119ab473a045137d26588ac00000000"
	bt, err := NewFromString(unsignedTx)
	if err != nil {
		t.Error(err)
		return
	}

	// Add the UTXO amount and script.
	bt.Inputs[0].PreviousTxSatoshis = 100000000
	bt.Inputs[0].Script = NewScriptFromString("76a914c0a3c167a28cabb9fbb495affa0761e6e74ac60d88ac")

	// Our private key.
	wif, err := btcutil.DecodeWIF("cNGwGSc7KRrTmdLUZ54fiSXWbhLNDc2Eg5zNucgQxyQCzuQ5YRDq")
	if err != nil {
		t.Error(err)
		return
	}

	privKeys := make([]*btcec.PrivateKey, 0)
	privKeys = append(privKeys, wif.PrivKey)

	var sigtype uint32 = SighashAll | SighashForkID

	sigs, _ := GetSignatures(bt, privKeys, sigtype)
	for _, sig := range sigs {
		pubkey := wif.PrivKey.PubKey().SerializeCompressed()
		buf := make([]byte, 0)
		buf = append(buf, cryptolib.VarInt(uint64(len(sig.Signature)))...)
		buf = append(buf, sig.Signature...)
		buf = append(buf, cryptolib.VarInt(uint64(len(pubkey)))...)
		buf = append(buf, pubkey...)
		bt.GetInputs()[0].Script = NewScriptFromBytes(buf)
	}

	expectedSignedTx := "010000000193a35408b6068499e0d5abd799d3e827d9bfe70c9b75ebe209c91d2507232651000000006b483045022100c1d77036dc6cd1f3fa1214b0688391ab7f7a16cd31ea4e5a1f7a415ef167df820220751aced6d24649fa235132f1e6969e163b9400f80043a72879237dab4a1190ad412103b8b40a84123121d260f5c109bc5a46ec819c2e4002e5ba08638783bfb4e01435ffffffff02404b4c00000000001976a91404ff367be719efa79d76e4416ffb072cd53b208888acde94a905000000001976a91404d03f746652cfcb6cb55119ab473a045137d26588ac00000000"
	if expectedSignedTx != hex.EncodeToString(bt.Hex()) {
		t.Errorf("Expected %q, got %x", expectedSignedTx, bt.Hex())
	}
}
