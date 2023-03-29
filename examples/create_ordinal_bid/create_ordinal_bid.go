package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"

	"github.com/libsv/go-bk/wif"
	"github.com/libsv/go-bt/v2"
	"github.com/libsv/go-bt/v2/bscript"
	"github.com/libsv/go-bt/v2/ord"
	"github.com/libsv/go-bt/v2/unlocker"
)

func main() {
	fundingWif, _ := wif.DecodeWIF("L5W2nyKUCsDStVUBwZj2Q3Ph5vcae4bgdzprZDYqDpvZA8AFguFH") // 19NfKd8aTwvb5ngfP29RxgfQzZt8KAYtQo
	fundingAddr, _ := bscript.NewAddressFromPublicKeyString(hex.EncodeToString(fundingWif.SerialisePubKey()), true)
	fundingScript, _ := bscript.NewP2PKHFromAddress(fundingAddr.AddressString)
	fundingUnlockerGetter := unlocker.Getter{PrivateKey: fundingWif.PrivKey}
	fundingUnlocker, _ := fundingUnlockerGetter.Unlocker(context.Background(), fundingScript)

	bidAmount := 100000000

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
				t, _ := hex.DecodeString("fc136d44114bdaa99f2d7d06a0fee514d376d974af53a3909fc43a79a3644653")
				return t
			}(),
			Vout:          uint32(0),
			LockingScript: fundingScript,
			Satoshis:      100027971,
			Unlocker:      &fundingUnlocker,
		},
	}

	mba := &ord.MakeBidArgs{
		BidAmount:   uint64(bidAmount),
		OrdinalTxID: "e17d7856c375640427943395d2341b6ed75f73afc8b22bb3681987278978a584",
		OrdinalVOut: 81,
		BidderUTXOs: us,
		BuyerReceiveOrdinalScript: func() *bscript.Script {
			s, _ := bscript.NewP2PKHFromAddress("1JPxYgWSYCb3ZEBBkcum84AHHdPWQzHGXj")
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

	pstx, err := ord.MakeBidToBuy1SatOrdinal(context.Background(), mba)
	if err != nil {
		log.Fatal(err.Error())
	}
	fmt.Println(pstx.String())
}
