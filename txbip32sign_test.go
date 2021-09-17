package bt_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/libsv/go-bk/bip32"
	"github.com/libsv/go-bt/v2"
	"github.com/stretchr/testify/assert"
)

func TestTx_Bip32SignAuto(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		tx             *bt.Tx
		pathGetterFunc bt.Bip32PathGetterFunc
		expHex         string
		expSigned      int
		expErr         error
	}{
		"valid tx (basic)": {
			tx: func() *bt.Tx {
				tx := bt.NewTx()
				assert.NotNil(t, tx)

				err := tx.From(
					"0e71f76a25cd1d38af84ac788f7b372de6c52a400db903c03890e259c6ce5144",
					0,
					"76a91407cfae7ee743646d148c565d951277a1607d6b4688ac",
					1000000)
				assert.NoError(t, err)

				err = tx.ChangeToAddress("mwV3YgnowbJJB3LcyCuqiKpdivvNNFiK7M", bt.NewFeeQuote())
				assert.NoError(t, err)

				return tx
			}(),
			pathGetterFunc: func(context.Context, string) (string, error) {
				return "2707843355/3214598338/2147483649", nil
			},
			expHex:    "01000000014451cec659e29038c003b90d402ac5e62d377b8f78ac84af381dcd256af7710e000000006a4730440220349d591e2334cbd6e05c108cd3ddca96fe3c770c56bdceb8aceb4783881b7a6f022029ea973689bc43251dba17861e70887801b188b804fd978469d9eb74ff9c501b41210289e621710008c545f7b900db03bb26be8fe08a11160cdd959a3312fa5bbebef3ffffffff01e0410f00000000001976a914af2590a45ae401651fdbdf59a76ad43d1862534088ac00000000",
			expSigned: 1,
		},
		"empty tx": {
			tx:     bt.NewTx(),
			expHex: bt.NewTx().String(),
		},
		"valid tx (wrong path)": {
			tx: func() *bt.Tx {
				tx := bt.NewTx()
				assert.NotNil(t, tx)

				err := tx.From(
					"0e71f76a25cd1d38af84ac788f7b372de6c52a400db903c03890e259c6ce5144",
					0,
					"76a91407cfae7ee743646d148c565d951277a1607d6b4688ac",
					1000000)
				assert.NoError(t, err)

				err = tx.ChangeToAddress("mwV3YgnowbJJB3LcyCuqiKpdivvNNFiK7M", bt.NewFeeQuote())
				assert.NoError(t, err)

				return tx
			}(),
			pathGetterFunc: func(context.Context, string) (string, error) {
				return "2707843355/3214598338/2147483650", nil
			},
			expHex:    "01000000014451cec659e29038c003b90d402ac5e62d377b8f78ac84af381dcd256af7710e0000000000ffffffff01e0410f00000000001976a914af2590a45ae401651fdbdf59a76ad43d1862534088ac00000000",
			expSigned: 0,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			tx := test.tx

			key, err := bip32.NewKeyFromString("tprv8ZgxMBicQKsPfLGWCgh24MgovBADKbvqTmT9BUJ8zyXi2y541oSigK5SmWqq1wTKh4PnGgeEe7boGnocyXWwEjk88hjDdSgy8rvryyPdHzL")
			assert.NoError(t, err)

			n, err := tx.Bip32SignAuto(context.Background(), &bt.LocalBip32SignerDeriver{MasterPrivateKey: key}, test.pathGetterFunc)

			if test.expErr != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, test.expErr.Error())
				return
			}

			fmt.Println(tx.String())

			assert.Equal(t, test.expSigned, len(n))
			assert.Equal(t, tx.String(), test.expHex)
		})
	}
}
