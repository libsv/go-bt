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

	sigs := getSignatures(bt, privKeys, sigtype)
	for _, sig := range sigs {
		pubkey := wif.PrivKey.PubKey().SerializeCompressed()
		buf := make([]byte, 0)
		buf = append(buf, cryptolib.VarInt(uint64(len(sig.Signature)))...)
		buf = append(buf, sig.Signature...)
		buf = append(buf, cryptolib.VarInt(uint64(len(pubkey)))...)
		buf = append(buf, pubkey...)
		bt.GetInputs()[0].script = NewScript(buf)
	}

	expected := "0200000001037ded84940e54c8c9e1ba73aa338a61d2ee4c4ac0d1faf2e8671896b0f8da63000000006a4730440220416154a5a117e89855397c6a7b2796d82107d20c1326bc917444e4ab84567b80022057de212dc0615ea1f4bbca817ed18be49acb96acff7760fc4d6447cbe772d1e8412103fc7c702eb7a03099ef01970b31ecbebe7ff77adc202d3749a8562ffc185a44a6ffffffff01806de729010000001976a91463ea0d776d45502d2226aed9ebdf5b676e232ca188ac00000000"
	if expected != hex.EncodeToString(bt.Hex()) {
		t.Errorf("Expected %q, got %q", expected, bt.Hex())
	}
}
