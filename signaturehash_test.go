package bt_test

import (
	"encoding/hex"
	"testing"

	"github.com/libsv/go-bt"
	"github.com/libsv/go-bt/bscript"
	"github.com/libsv/go-bt/sighash"
	"github.com/stretchr/testify/assert"
)

func TestTx_CalcInputPreimage(t *testing.T) {
	t.Parallel()

	var testVector = []struct {
		name               string
		unsignedTx         string
		index              int
		previousTxSatoshis uint64
		previousTxScript   string
		sigHashType        sighash.Flag
		expectedPreimage   string
	}{
		{
			"1 Input 2 Outputs - SIGHASH_ALL (FORKID)",
			"010000000193a35408b6068499e0d5abd799d3e827d9bfe70c9b75ebe209c91d25072326510000000000ffffffff02404b4c00000000001976a91404ff367be719efa79d76e4416ffb072cd53b208888acde94a905000000001976a91404d03f746652cfcb6cb55119ab473a045137d26588ac00000000",
			0,
			100000000,
			"76a914c0a3c167a28cabb9fbb495affa0761e6e74ac60d88ac",
			sighash.AllForkID,
			"010000007ced5b2e5cf3ea407b005d8b18c393b6256ea2429b6ff409983e10adc61d0ae83bb13029ce7b1f559ef5e747fcac439f1455a2ec7c5f09b72290795e7066504493a35408b6068499e0d5abd799d3e827d9bfe70c9b75ebe209c91d2507232651000000001976a914c0a3c167a28cabb9fbb495affa0761e6e74ac60d88ac00e1f50500000000ffffffff87841ab2b7a4133af2c58256edb7c3c9edca765a852ebe2d0dc962604a30f1030000000041000000",
		},
		{
			"2 Inputs 3 Outputs - SIGHASH_ALL (FORKID) - Index 0",
			"01000000027e2705da59f7112c7337d79840b56fff582b8f3a0e9df8eb19e282377bebb1bc0100000000ffffffffdebe6fe5ad8e9220a10fcf6340f7fca660d87aeedf0f74a142fba6de1f68d8490000000000ffffffff0300e1f505000000001976a9142987362cf0d21193ce7e7055824baac1ee245d0d88ac00e1f505000000001976a9143ca26faa390248b7a7ac45be53b0e4004ad7952688ac34657fe2000000001976a914eb0bd5edba389198e73f8efabddfc61666969ff788ac00000000",
			0,
			2000000000,
			"76a914eb0bd5edba389198e73f8efabddfc61666969ff788ac",
			sighash.AllForkID,
			"01000000eaef7a1b82f72f4097e63b0173906d690cc137221d221fc4150bae88570fa356752adad0a7b9ceca853768aebb6965eca126a62965f698a0c1bc43d83db632ad7e2705da59f7112c7337d79840b56fff582b8f3a0e9df8eb19e282377bebb1bc010000001976a914eb0bd5edba389198e73f8efabddfc61666969ff788ac0094357700000000ffffffff0cf3246582f4b1b5fd150b942916c7d5c78e80259cbab1a761a9e4ac3a66e0a70000000041000000",
		},
		{
			"2 Inputs 3 Outputs - SIGHASH_ALL (FORKID) - Index 1",
			"01000000027e2705da59f7112c7337d79840b56fff582b8f3a0e9df8eb19e282377bebb1bc0100000000ffffffffdebe6fe5ad8e9220a10fcf6340f7fca660d87aeedf0f74a142fba6de1f68d8490000000000ffffffff0300e1f505000000001976a9142987362cf0d21193ce7e7055824baac1ee245d0d88ac00e1f505000000001976a9143ca26faa390248b7a7ac45be53b0e4004ad7952688ac34657fe2000000001976a914eb0bd5edba389198e73f8efabddfc61666969ff788ac00000000",
			1,
			2000000000,
			"76a914eb0bd5edba389198e73f8efabddfc61666969ff788ac",
			sighash.AllForkID,
			"01000000eaef7a1b82f72f4097e63b0173906d690cc137221d221fc4150bae88570fa356752adad0a7b9ceca853768aebb6965eca126a62965f698a0c1bc43d83db632addebe6fe5ad8e9220a10fcf6340f7fca660d87aeedf0f74a142fba6de1f68d849000000001976a914eb0bd5edba389198e73f8efabddfc61666969ff788ac0094357700000000ffffffff0cf3246582f4b1b5fd150b942916c7d5c78e80259cbab1a761a9e4ac3a66e0a70000000041000000",
		},
		// TODO: add different SIGHASH flags
		// note: checking bsv.js - using different sighash flags gives same
		// sighash for some reason.. check later..
	}

	for _, test := range testVector {
		t.Run(test.name, func(t *testing.T) {
			tx, err := bt.NewTxFromString(test.unsignedTx)
			assert.NoError(t, err)
			assert.NotNil(t, tx)

			// Add the UTXO amount and script (PreviousTxScript already in unsiged tx)
			tx.InputIdx(test.index).PreviousTxSatoshis = test.previousTxSatoshis
			tx.InputIdx(test.index).PreviousTxScript, err = bscript.NewFromHexString(test.previousTxScript)
			assert.NoError(t, err)

			var actualSigHash []byte
			actualSigHash, err = tx.CalcInputPreimage(uint32(test.index), sighash.All|sighash.ForkID)
			assert.NoError(t, err)
			assert.Equal(t, test.expectedPreimage, hex.EncodeToString(actualSigHash))
		})
	}
}

func TestTx_CalcInputSignatureHash(t *testing.T) {
	t.Parallel()

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
			"be9a42ef2e2dd7ef02cd631290667292cbbc5018f4e3f6843a8f4c302a2111b1",
		},
		{
			"2 Inputs 3 Outputs - SIGHASH_ALL (FORKID) - Index 0",
			"01000000027e2705da59f7112c7337d79840b56fff582b8f3a0e9df8eb19e282377bebb1bc0100000000ffffffffdebe6fe5ad8e9220a10fcf6340f7fca660d87aeedf0f74a142fba6de1f68d8490000000000ffffffff0300e1f505000000001976a9142987362cf0d21193ce7e7055824baac1ee245d0d88ac00e1f505000000001976a9143ca26faa390248b7a7ac45be53b0e4004ad7952688ac34657fe2000000001976a914eb0bd5edba389198e73f8efabddfc61666969ff788ac00000000",
			0,
			2000000000,
			"76a914eb0bd5edba389198e73f8efabddfc61666969ff788ac",
			sighash.AllForkID,
			"8b15eecfb6d5e727485e19797b5d1829e0630e8b43c806707685238e28a3194c",
		},
		{
			"2 Inputs 3 Outputs - SIGHASH_ALL (FORKID) - Index 1",
			"01000000027e2705da59f7112c7337d79840b56fff582b8f3a0e9df8eb19e282377bebb1bc0100000000ffffffffdebe6fe5ad8e9220a10fcf6340f7fca660d87aeedf0f74a142fba6de1f68d8490000000000ffffffff0300e1f505000000001976a9142987362cf0d21193ce7e7055824baac1ee245d0d88ac00e1f505000000001976a9143ca26faa390248b7a7ac45be53b0e4004ad7952688ac34657fe2000000001976a914eb0bd5edba389198e73f8efabddfc61666969ff788ac00000000",
			1,
			2000000000,
			"76a914eb0bd5edba389198e73f8efabddfc61666969ff788ac",
			sighash.AllForkID,
			"7b72c355a2714a5039d97fbd5eee792099b0eab4bf07d2e5bfcfc3309f81badb",
		},
		// TODO: add different SIGHASH flags
		// note: checking bsv.js - using different sighash flags gives same
		// sighash for some reason.. check later..
	}

	for _, test := range testVector {
		t.Run(test.name, func(t *testing.T) {
			tx, err := bt.NewTxFromString(test.unsignedTx)
			assert.NoError(t, err)
			assert.NotNil(t, tx)

			// Add the UTXO amount and script (PreviousTxScript already in unsiged tx)
			tx.Inputs()[test.index].PreviousTxSatoshis = test.previousTxSatoshis
			tx.Inputs()[test.index].PreviousTxScript, err = bscript.NewFromHexString(test.previousTxScript)
			assert.NoError(t, err)

			var actualSigHash []byte
			actualSigHash, err = tx.CalcInputSignatureHash(test.index, sighash.All|sighash.ForkID)
			assert.NoError(t, err)
			assert.Equal(t, test.expectedSigHash, hex.EncodeToString(actualSigHash))
		})
	}
}
