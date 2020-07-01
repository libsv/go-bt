package signature_test

import (
	"encoding/hex"
	"testing"

	"github.com/bitcoinsv/bsvutil"
	"github.com/libsv/libsv/script"
	"github.com/libsv/libsv/transaction"
	"github.com/libsv/libsv/transaction/signature"
)

func TestSignAuto(t *testing.T) {
	unsignedTx := "010000000193a35408b6068499e0d5abd799d3e827d9bfe70c9b75ebe209c91d25072326510000000000ffffffff02404b4c00000000001976a91404ff367be719efa79d76e4416ffb072cd53b208888acde94a905000000001976a91404d03f746652cfcb6cb55119ab473a045137d26588ac00000000"
	tx, err := transaction.NewFromString(unsignedTx)

	if err != nil {
		t.Fatal("Failed to create transaction")
	}
	// Add the UTXO amount and script.
	tx.Inputs[0].PreviousTxSatoshis = 100000000
	tx.Inputs[0].PreviousTxScript, _ = script.NewFromHexString("76a914c0a3c167a28cabb9fbb495affa0761e6e74ac60d88ac")

	// Our private key.
	wif, err := bsvutil.DecodeWIF("cNGwGSc7KRrTmdLUZ54fiSXWbhLNDc2Eg5zNucgQxyQCzuQ5YRDq")
	if err != nil {
		t.Fatal(err)
	}
	signer := signature.InternalSigner{PrivateKey: wif.PrivKey, SigHashFlag: 0}
	err = tx.SignAuto(&signer)
	if err != nil {
		t.Fatal(err)
	}
	expectedSignedTx := "010000000193a35408b6068499e0d5abd799d3e827d9bfe70c9b75ebe209c91d2507232651000000006b483045022100c1d77036dc6cd1f3fa1214b0688391ab7f7a16cd31ea4e5a1f7a415ef167df820220751aced6d24649fa235132f1e6969e163b9400f80043a72879237dab4a1190ad412103b8b40a84123121d260f5c109bc5a46ec819c2e4002e5ba08638783bfb4e01435ffffffff02404b4c00000000001976a91404ff367be719efa79d76e4416ffb072cd53b208888acde94a905000000001976a91404d03f746652cfcb6cb55119ab473a045137d26588ac00000000"

	if hex.EncodeToString(tx.ToBytes()) != expectedSignedTx {
		t.Errorf("Expecting %s\n, got %s\n", expectedSignedTx, hex.EncodeToString(tx.ToBytes()))
	}

	// TODO: what is this for?
	//if unsignedTx == expectedSignedTx {
	//	t.Errorf("Expected and signed TX strings in code identical")
	//}
}
