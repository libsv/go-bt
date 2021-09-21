package bt_test

import (
	"context"
	"errors"
	"testing"

	"github.com/libsv/go-bk/bip32"
	"github.com/libsv/go-bt/v2"
	"github.com/stretchr/testify/assert"
)

type mockSignerDeriver struct {
	deriverFunc func(string) (bt.Signer, error)
}

func (m *mockSignerDeriver) DeriveBip32Signer(derivationPath string) (bt.Signer, error) {
	if m.deriverFunc == nil {
		return nil, errors.New("deriverFunc not defined for this test")
	}
	return m.deriverFunc(derivationPath)
}

func TestTx_SignAllBip32(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		tx             *bt.Tx
		pathGetterFunc bt.Bip32PathGetterFunc
		mockDeriver    bt.Bip32SignerDeriver
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
		"valid tx (empty path)": {
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
				return "", nil
			},
			expHex: "01000000014451cec659e29038c003b90d402ac5e62d377b8f78ac84af381dcd256af7710e000000006a47304402202517dfda25e61c80968d8c7005b461c30f88d299e9fd122e05b14d80fc6a280902205f096136825bae3ec4a084cd8c4445bfaf2ab234ef6478a61e5d05ec6bb76386412102016f7083a7f49a7df8f207e14c29e3d90c24f8327d321d70eb971e1d08bcc724ffffffff01e0410f00000000001976a914af2590a45ae401651fdbdf59a76ad43d1862534088ac00000000",
		},
		"error with path getter func is returned to user": {
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
				return "", errors.New("wat")
			},
			expErr: errors.New("wat"),
		},
		"nil path getter func errors": {
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
			expErr: errors.New("Bip32PathGetterFunc not provided"),
		},
		"derivation error is reported to the user": {
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
			mockDeriver: &mockSignerDeriver{
				deriverFunc: func(string) (bt.Signer, error) {
					return nil, errors.New("error creating derived signer")
				},
			},
			pathGetterFunc: func(context.Context, string) (string, error) {
				return "", nil
			},
			expErr: errors.New("error creating derived signer"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if test.mockDeriver == nil {
				key, err := bip32.NewKeyFromString("tprv8ZgxMBicQKsPfLGWCgh24MgovBADKbvqTmT9BUJ8zyXi2y541oSigK5SmWqq1wTKh4PnGgeEe7boGnocyXWwEjk88hjDdSgy8rvryyPdHzL")
				assert.NoError(t, err)

				test.mockDeriver = &bt.LocalBip32SignerDeriver{MasterPrivateKey: key}
			}

			err := test.tx.SignAllBip32(context.Background(), test.mockDeriver, test.pathGetterFunc)

			if test.expErr != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, test.expErr.Error())
				return
			}

			assert.Equal(t, test.expSigned, test.expSigned)
			assert.Equal(t, test.tx.String(), test.expHex)
		})
	}
}