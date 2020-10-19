package bt_test

import (
	"encoding/hex"
	"testing"

	"github.com/libsv/libsv/bt"
	"github.com/libsv/libsv/bt/sig/sighash"
	"github.com/libsv/libsv/script"
)

var testVector = []struct {
	name               string
	unsignedTx         string
	index              uint32
	previousTxSatoshis uint64
	previousTxScript   string
	sigHashType        sighash.Flag
	expectedSigHash    string
}{
	{
		"1 Input 2 Outputs - SIGHASH_ALL (FORKID)",
		"010000000193a35408b6068499e0d5abd799d3e827d9bfe70c9b75ebe209c91d25072326510000000000ffffffff02404b4c00000000001976a91404ff367be719efa79d76e4416ffb072cd53b208888acde94a905000000001976a91404d03f746652cfcb6cb55119ab473a045137d26588ac00000000",
		0,
		100000000,
		"76a914c0a3c167a28cabb9fbb495affa0761e6e74ac60d88ac",
		sighash.AllForkID,
		"b111212a304c8f3a84f6e3f41850bccb927266901263cd02efd72d2eef429abe",
	},
	{
		"2 Inputs 3 Outputs - SIGHASH_ALL (FORKID) - Index 0",
		"01000000027e2705da59f7112c7337d79840b56fff582b8f3a0e9df8eb19e282377bebb1bc0100000000ffffffffdebe6fe5ad8e9220a10fcf6340f7fca660d87aeedf0f74a142fba6de1f68d8490000000000ffffffff0300e1f505000000001976a9142987362cf0d21193ce7e7055824baac1ee245d0d88ac00e1f505000000001976a9143ca26faa390248b7a7ac45be53b0e4004ad7952688ac34657fe2000000001976a914eb0bd5edba389198e73f8efabddfc61666969ff788ac00000000",
		0,
		2000000000,
		"76a914eb0bd5edba389198e73f8efabddfc61666969ff788ac",
		sighash.AllForkID,
		"4c19a3288e2385767006c8438b0e63e029185d7b79195e4827e7d5b6cfee158b",
	},
	{
		"2 Inputs 3 Outputs - SIGHASH_ALL (FORKID) - Index 1",
		"01000000027e2705da59f7112c7337d79840b56fff582b8f3a0e9df8eb19e282377bebb1bc0100000000ffffffffdebe6fe5ad8e9220a10fcf6340f7fca660d87aeedf0f74a142fba6de1f68d8490000000000ffffffff0300e1f505000000001976a9142987362cf0d21193ce7e7055824baac1ee245d0d88ac00e1f505000000001976a9143ca26faa390248b7a7ac45be53b0e4004ad7952688ac34657fe2000000001976a914eb0bd5edba389198e73f8efabddfc61666969ff788ac00000000",
		1,
		2000000000,
		"76a914eb0bd5edba389198e73f8efabddfc61666969ff788ac",
		sighash.AllForkID,
		"dbba819f30c3cfbfe5d207bfb4eab0992079ee5ebd7fd939504a71a255c3727b",
	},
	// TODO: add different SIGHASH flags
	// note: checking bsv.js - using different sighash flags gives same
	// sighash for some reason.. check later..
}

func TestSignatureHashes(t *testing.T) {

	for _, test := range testVector {
		t.Run(test.name, func(t *testing.T) {
			tx, err := bt.NewFromString(test.unsignedTx)
			if err != nil {
				t.Error(err)
			}

			// Add the UTXO amount and script (PreviousTxScript already in unsiged tx)
			tx.Inputs[test.index].PreviousTxSatoshis = test.previousTxSatoshis
			tx.Inputs[test.index].PreviousTxScript, _ = script.NewFromHexString(test.previousTxScript)

			actualSigHash, err := tx.GetInputSignatureHash(test.index, sighash.All|sighash.ForkID)
			if err != nil {
				t.Error(err)
				return
			}

			if test.expectedSigHash != hex.EncodeToString(actualSigHash) {
				t.Errorf("Error expected %s got %x", test.expectedSigHash, actualSigHash)
			}
		})
	}
}
