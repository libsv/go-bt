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
				t, _ := hex.DecodeString("272ccf2b83f24e3531b2327dfc8bb4c483e21ecb7620c3a41e75103327a2a8f9")
				return t
			}(),
			Vout:          uint32(0),
			LockingScript: fundingScript,
			Satoshis:      20,
			Unlocker:      &fundingUnlocker,
		},
		{
			TxID: func() []byte {
				t, _ := hex.DecodeString("272ccf2b83f24e3531b2327dfc8bb4c483e21ecb7620c3a41e75103327a2a8f9")
				return t
			}(),
			Vout:          uint32(1),
			LockingScript: fundingScript,
			Satoshis:      20,
			Unlocker:      &fundingUnlocker,
		},
		{
			TxID: func() []byte {
				t, _ := hex.DecodeString("272ccf2b83f24e3531b2327dfc8bb4c483e21ecb7620c3a41e75103327a2a8f9")
				return t
			}(),
			Vout:          uint32(2),
			LockingScript: fundingScript,
			Satoshis:      1520,
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
			t, _ := hex.DecodeString("976465589a195b70720666ddbf400b5a566bb829bcf1d6a1e076d3ab2c53cf75")
			return t
		}(),
		Vout: uint32(1),
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
