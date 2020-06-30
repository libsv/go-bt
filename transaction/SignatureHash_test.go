package transaction_test

import (
	"encoding/hex"
	"testing"

	"github.com/libsv/libsv/script"
	"github.com/libsv/libsv/transaction"
)

func TestGetInputSignatureHash(t *testing.T) {
	unsignedTx := "010000000193a35408b6068499e0d5abd799d3e827d9bfe70c9b75ebe209c91d25072326510000000000ffffffff02404b4c00000000001976a91404ff367be719efa79d76e4416ffb072cd53b208888acde94a905000000001976a91404d03f746652cfcb6cb55119ab473a045137d26588ac00000000"
	tx, err := transaction.NewFromString(unsignedTx)

	// Add the UTXO amount and script.
	tx.Inputs[0].PreviousTxSatoshis = uint64(100000000)

	tx.Inputs[0].PreviousTxScript, _ = script.NewFromHexString("76a914c0a3c167a28cabb9fbb495affa0761e6e74ac60d88ac")

	expectedSigHash := "b111212a304c8f3a84f6e3f41850bccb927266901263cd02efd72d2eef429abe"

	actualSigHash, err := tx.GetInputSignatureHash(0, transaction.SigHashAll|transaction.SigHashForkID)
	if err != nil {
		t.Error(err)
		return
	}
	if expectedSigHash != hex.EncodeToString(actualSigHash) {
		t.Errorf("Error expected %s got %s", expectedSigHash, hex.EncodeToString(actualSigHash))
	}
}

func Test2Inputs3Outputs(t *testing.T) {
	unsignedTx := "01000000027e2705da59f7112c7337d79840b56fff582b8f3a0e9df8eb19e282377bebb1bc0100000000ffffffffdebe6fe5ad8e9220a10fcf6340f7fca660d87aeedf0f74a142fba6de1f68d8490000000000ffffffff0300e1f505000000001976a9142987362cf0d21193ce7e7055824baac1ee245d0d88ac00e1f505000000001976a9143ca26faa390248b7a7ac45be53b0e4004ad7952688ac34657fe2000000001976a914eb0bd5edba389198e73f8efabddfc61666969ff788ac00000000"
	tx, err := transaction.NewFromString(unsignedTx)

	// Add the UTXO amount and script.
	tx.Inputs[0].PreviousTxSatoshis = uint64(2000000000)
	tx.Inputs[0].PreviousTxScript, _ = script.NewFromHexString("76a914eb0bd5edba389198e73f8efabddfc61666969ff788ac")

	tx.Inputs[1].PreviousTxSatoshis = uint64(2000000000)
	tx.Inputs[1].PreviousTxScript, _ = script.NewFromHexString("76a914eb0bd5edba389198e73f8efabddfc61666969ff788ac")

	expectedSigHash0 := "4c19a3288e2385767006c8438b0e63e029185d7b79195e4827e7d5b6cfee158b"

	actualSigHash0, err := tx.GetInputSignatureHash(0, transaction.SigHashAll|transaction.SigHashForkID)
	if err != nil {
		t.Error(err)
		return
	}
	if expectedSigHash0 != hex.EncodeToString(actualSigHash0) {
		t.Errorf("Error expected %s got %s", expectedSigHash0, hex.EncodeToString(actualSigHash0))
	}

	expectedSigHash1 := "dbba819f30c3cfbfe5d207bfb4eab0992079ee5ebd7fd939504a71a255c3727b"

	actualSigHash1, err := tx.GetInputSignatureHash(1, transaction.SigHashAll|transaction.SigHashForkID)
	if err != nil {
		t.Error(err)
		return
	}
	if expectedSigHash1 != hex.EncodeToString(actualSigHash1) {
		t.Errorf("Error expected %s got %s", expectedSigHash1, hex.EncodeToString(actualSigHash1))
	}
}

func TestGetInputSignatureHash2(t *testing.T) {
	unsignedTx := "01000000017e2705da59f7112c7337d79840b56fff582b8f3a0e9df8eb19e282377bebb1bc0100000000ffffffff01a0933577000000001976a914eb0bd5edba389198e73f8efabddfc61666969ff788ac00000000"
	tx, err := transaction.NewFromString(unsignedTx)

	// Add the UTXO amount and script.
	tx.Inputs[0].PreviousTxSatoshis = uint64(2000000000)

	tx.Inputs[0].PreviousTxScript, _ = script.NewFromHexString("76a914eb0bd5edba389198e73f8efabddfc61666969ff788ac")

	expectedSigHash := "2e4dfcf35598ef4e1c16bc16497f72907f9e1184ed71a6fd97082b21c88fc300"

	actualSigHash, err := tx.GetInputSignatureHash(0, transaction.SigHashAll|transaction.SigHashForkID)
	if err != nil {
		t.Error(err)
		return
	}
	if expectedSigHash != hex.EncodeToString(actualSigHash) {
		t.Errorf("Error expected %s got %s", expectedSigHash, hex.EncodeToString(actualSigHash))
	}
}

// func TestGetInputSignatureHashCoinbase(t *testing.T) {
// 	unsignedTx := "02000000010000000000000000000000000000000000000000000000000000000000000000ffffffff2e039b1e1304c0737c5b68747470733a2f2f6769746875622e636f6d2f62636578742f01000001c096020000000000ffffffff014a355009000000001976a91448b20e254c0677e760bab964aec16818d6b7134a88ac00000000"
// 	tx, err := transaction.NewFromString(unsignedTx)

// 	expectedSigHash := "6829f7d44dfd4654749b8027f44c9381527199f78ae9b0d58ffc03fdab3c82f1"

// 	actualSigHash, err := tx.GetInputSignatureHash(0, transaction.SigHashAll|transaction.SigHashForkID)
// 	if err != nil {
// 		t.Error(err)
// 		return
// 	}
// 	if expectedSigHash != hex.EncodeToString(actualSigHash) {
// 		t.Errorf("Error expected %s got %s", expectedSigHash, hex.EncodeToString(actualSigHash))
// 	}
// }
