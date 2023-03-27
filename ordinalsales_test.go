package bt_test

import (
	"context"
	"encoding/hex"
	"testing"

	"github.com/libsv/go-bk/wif"
	"github.com/libsv/go-bt/v2"
	"github.com/libsv/go-bt/v2/bscript"
	"github.com/libsv/go-bt/v2/unlocker"
	"github.com/stretchr/testify/assert"
)

func TestOfferToSellPSBTNoErrors(t *testing.T) {
	t.Run("create PSBT to make an offer to sell ordinal", func(t *testing.T) {

		ordWif, _ := wif.DecodeWIF("L42PyNwEKE4XRaa8PzPh7JZurSAWJmx49nbVfaXYuiQg3RCubwn7") // 1JijRHzVfub38S2hizxkxEcVKQwuCTZmxJ
		ordPrefixAddr, _ := bscript.NewAddressFromPublicKeyString(hex.EncodeToString(ordWif.SerialisePubKey()), true)
		ordPrefixScript, _ := bscript.NewP2PKHFromAddress(ordPrefixAddr.AddressString)

		ordUnlockerGetter := unlocker.Getter{PrivateKey: ordWif.PrivKey}
		ordUnlocker, _ := ordUnlockerGetter.Unlocker(context.Background(), ordPrefixScript)

		pstx, err := bt.ListOrdinalForSale(context.Background(), &bt.ListOrdinalArgs{
			SellerReceiveOutput: &bt.Output{
				Satoshis: 500,
				LockingScript: func() *bscript.Script {
					s, _ := bscript.NewP2PKHFromAddress("1C3V9TTJefP8Hft96sVf54mQyDJh8Ze4w4") // L1JWiLZtCkkqin41XtQ2Jxo1XGxj1R4ydT2zmxPiaeQfuyUK631D
					return s
				}(),
			},
			OrdinalUTXO: &bt.UTXO{
				TxID: func() []byte {
					t, _ := hex.DecodeString("8f027fb1361ae46ac165e1d90e5436ed9c11d4eeaa60669ab90386a3abd9ce6a")
					return t
				}(),
				Vout: uint32(0),
				LockingScript: func() *bscript.Script {
					// hello world (text/plain) test inscription
					s, _ := bscript.NewFromHexString("76a914c25e9a2b70ec83d7b4fbd0f36f00a86723a48e6b88ac0063036f72645118746578742f706c61696e3b636861727365743d7574662d38000d48656c6c6f2c20776f726c642168")
					return s
				}(),
				Satoshis: 1,
			},
			OrdinalUnlocker: ordUnlocker,
		})

		assert.NoError(t, err)

		us := []*bt.UTXO{
			{
				TxID: func() []byte {
					t, _ := hex.DecodeString("61dfcc313763eb5332c036131facdf92c2ca9d663ffb96e4b997086a0643d635")
					return t
				}(),
				Vout:          uint32(0),
				LockingScript: ordPrefixScript,
				Satoshis:      10,
				Unlocker:      &ordUnlocker,
			},
			{
				TxID: func() []byte {
					t, _ := hex.DecodeString("61dfcc313763eb5332c036131facdf92c2ca9d663ffb96e4b997086a0643d635")
					return t
				}(),
				Vout:          uint32(1),
				LockingScript: ordPrefixScript,
				Satoshis:      10,
				Unlocker:      &ordUnlocker,
			},
			{
				TxID: func() []byte {
					t, _ := hex.DecodeString("8f027fb1361ae46ac165e1d90e5436ed9c11d4eeaa60669ab90386a3abd9ce6a")
					return t
				}(),
				Vout:          uint32(1),
				LockingScript: ordPrefixScript,
				Satoshis:      953,
				Unlocker:      &ordUnlocker,
			},
		}

		buyerOrdS, _ := bscript.NewP2PKHFromAddress("1HebepswCi6huw1KJ7LvkrgemAV63TyVUs") // KwQq67d4Jds3wxs3kQHB8PPwaoaBQfNKkzAacZeMesb7zXojVYpj
		dummyS, _ := bscript.NewP2PKHFromAddress("19NfKd8aTwvb5ngfP29RxgfQzZt8KAYtQo")    // L5W2nyKUCsDStVUBwZj2Q3Ph5vcae4bgdzprZDYqDpvZA8AFguFH
		changeS, _ := bscript.NewP2PKHFromAddress("19NfKd8aTwvb5ngfP29RxgfQzZt8KAYtQo")   // L5W2nyKUCsDStVUBwZj2Q3Ph5vcae4bgdzprZDYqDpvZA8AFguFH

		_, err = bt.AcceptOrdinalSaleListing(context.Background(), &bt.AcceptListingArgs{
			PSTx:                      pstx,
			Utxos:                     us,
			BuyerReceiveOrdinalScript: buyerOrdS,
			DummyOutputScript:         dummyS,
			ChangeScript:              changeS,
			FQ:                        bt.NewFeeQuote(),
		})
		assert.NoError(t, err)
	})
}
