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

func TestMakeOfferToSellPSBT(t *testing.T) {
	t.Run("create PSBT to make an offer to sell ordinal", func(t *testing.T) {

		w, _ := wif.DecodeWIF("KzTeKvX5VCWkcaB1DCC2WeU7PML93XpKnqMRhXsyrHqFUhJvJyE9")
		pubkey := w.SerialisePubKey()
		addr, _ := bscript.NewAddressFromPublicKeyString(hex.EncodeToString(pubkey), true)
		s, _ := bscript.NewP2PKHFromAddress(addr.AddressString)

		unlocker := unlocker.Getter{PrivateKey: w.PrivKey}
		u, _ := unlocker.Unlocker(context.Background(), s)

		tx, err := bt.MakeOfferToSellOrdinal(context.Background(),
			&bt.Output{
				Satoshis: 100000,
				LockingScript: func() *bscript.Script {
					s, _ := bscript.NewP2PKHFromAddress("1PyWzkfKrq1kakvLTeaCdAL8y8UJAcZAqU")
					return s
				}(),
			},
			&bt.UTXO{
				TxID: func() []byte {
					t, _ := hex.DecodeString("5a2814071ef18c97248cf6050057cc5df09b34ae9a3105b079c002494b6e277d")
					return t
				}(),
				Vout:          uint32(0),
				LockingScript: s,
				Satoshis:      2,
			},
			u)

		assert.NoError(t, err)
		fmt.Println(tx.String())
		// TODO: add checks
	})
}
