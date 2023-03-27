package bt_test

import (
	"context"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/libsv/go-bk/wif"
	"github.com/libsv/go-bt/v2"
	"github.com/libsv/go-bt/v2/bscript"
	"github.com/libsv/go-bt/v2/unlocker"
	"github.com/stretchr/testify/assert"
)

func TestBidToBuyPSBTNoErrors(t *testing.T) {
	fundingWif, _ := wif.DecodeWIF("L5W2nyKUCsDStVUBwZj2Q3Ph5vcae4bgdzprZDYqDpvZA8AFguFH") // 19NfKd8aTwvb5ngfP29RxgfQzZt8KAYtQo
	fundingAddr, _ := bscript.NewAddressFromPublicKeyString(hex.EncodeToString(fundingWif.SerialisePubKey()), true)
	fundingScript, _ := bscript.NewP2PKHFromAddress(fundingAddr.AddressString)
	fundingUnlockerGetter := unlocker.Getter{PrivateKey: fundingWif.PrivKey}
	fundingUnlocker, _ := fundingUnlockerGetter.Unlocker(context.Background(), fundingScript)

	bidAmount := 250

	us := []*bt.UTXO{
		{
			TxID: func() []byte {
				t, _ := hex.DecodeString("411084d83d4f380cfc331ed849946bd7f354ca17138dbd723a6420ec9f5f4bd1")
				return t
			}(),
			Vout:          uint32(0),
			LockingScript: fundingScript,
			Satoshis:      20,
			Unlocker:      &fundingUnlocker,
		},
		{
			TxID: func() []byte {
				t, _ := hex.DecodeString("411084d83d4f380cfc331ed849946bd7f354ca17138dbd723a6420ec9f5f4bd1")
				return t
			}(),
			Vout:          uint32(1),
			LockingScript: fundingScript,
			Satoshis:      20,
			Unlocker:      &fundingUnlocker,
		},
		{
			TxID: func() []byte {
				t, _ := hex.DecodeString("4d815adc39a740810cb438eb285f6e08ae3957fdc4e4806399babfa806dfc456")
				return t
			}(),
			Vout:          uint32(0),
			LockingScript: fundingScript,
			Satoshis:      100000000,
			Unlocker:      &fundingUnlocker,
		},
	}

	ordWif, _ := wif.DecodeWIF("KwQq67d4Jds3wxs3kQHB8PPwaoaBQfNKkzAacZeMesb7zXojVYpj") // 1HebepswCi6huw1KJ7LvkrgemAV63TyVUs
	ordPrefixAddr, _ := bscript.NewAddressFromPublicKeyString(hex.EncodeToString(ordWif.SerialisePubKey()), true)
	ordPrefixScript, _ := bscript.NewP2PKHFromAddress(ordPrefixAddr.AddressString)
	ordUnlockerGetter := unlocker.Getter{PrivateKey: ordWif.PrivKey}
	ordUnlocker, _ := ordUnlockerGetter.Unlocker(context.Background(), ordPrefixScript)

	ordUTXO := &bt.UTXO{
		TxID: func() []byte {
			t, _ := hex.DecodeString("e17d7856c375640427943395d2341b6ed75f73afc8b22bb3681987278978a584")
			return t
		}(),
		Vout: uint32(81),
		LockingScript: func() *bscript.Script {
			s, _ := bscript.NewFromHexString("76a914b69e544cbf33c4eabdd5cf8792cd4e53f5ed6d1788ac")
			return s
		}(),
		Satoshis: 1,
	}

	mba := &bt.MakeBidArgs{
		BidAmount:   uint64(bidAmount),
		OrdinalTxID: ordUTXO.TxIDStr(),
		OrdinalVOut: ordUTXO.Vout,
		BidderUTXOs: us,
		BuyerReceiveOrdinalScript: func() *bscript.Script {
			s, _ := bscript.NewP2PKHFromAddress("12R2qFEoUtWwwVecgrkxwMZNnMq6GB8pQW") // L3kLQ9rpDBLgbh3GfPSbXDGwxgmK2Dcb6Qrp4JZRRcne8FMDZWDc
			return s
		}(),
		DummyOutputScript: func() *bscript.Script {
			s, _ := bscript.NewP2PKHFromAddress("19NfKd8aTwvb5ngfP29RxgfQzZt8KAYtQo") // L1JWiLZtCkkqin41XtQ2Jxo1XGxj1R4ydT2zmxPiaeQfuyUK631D
			return s
		}(),
		ChangeScript: func() *bscript.Script {
			s, _ := bscript.NewP2PKHFromAddress("19NfKd8aTwvb5ngfP29RxgfQzZt8KAYtQo") // L1JWiLZtCkkqin41XtQ2Jxo1XGxj1R4ydT2zmxPiaeQfuyUK631D
			return s
		}(),
		FQ: bt.NewFeeQuote(),
	}

	pstx, CreateBidError := bt.MakeBidToBuy1SatOrdinal(context.Background(), mba)
	fmt.Println(pstx.String())

	t.Run("no errors creating bid to buy ordinal", func(t *testing.T) {
		assert.NoError(t, CreateBidError)
	})

	t.Run("validate PSBT bid to buy ordinal", func(t *testing.T) {
		vba := &bt.ValidateBidArgs{
			BidAmount:  uint64(bidAmount),
			ExpectedFQ: bt.NewFeeQuote(),
			// insert ordinal utxo at index 2
			PreviousUTXOs: append(us[:2], append([]*bt.UTXO{ordUTXO}, us[2:]...)...),
		}
		assert.True(t, vba.Validate(pstx))
	})

	t.Run("no errors when accepting bid", func(t *testing.T) {
		tx, err := bt.AcceptBidToBuy1SatOrdinal(context.Background(), &bt.ValidateBidArgs{
			BidAmount:     uint64(bidAmount),
			ExpectedFQ:    bt.NewFeeQuote(),
			PreviousUTXOs: append(us[:2], append([]*bt.UTXO{ordUTXO}, us[2:]...)...),
		},
			&bt.AcceptBidArgs{
				PSTx: pstx,
				SellerReceiveOrdinalScript: func() *bscript.Script {
					s, _ := bscript.NewP2PKHFromAddress("1C3V9TTJefP8Hft96sVf54mQyDJh8Ze4w4") // L1JWiLZtCkkqin41XtQ2Jxo1XGxj1R4ydT2zmxPiaeQfuyUK631D
					return s
				}(),
				OrdinalUnlocker: ordUnlocker,
			})

		assert.NoError(t, err)
		fmt.Println(tx.String())
	})
}
