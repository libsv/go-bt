package txjson_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/libsv/go-bk/wif"
	"github.com/libsv/go-bt/v2"
	"github.com/libsv/go-bt/v2/bscript"
	"github.com/libsv/go-bt/v2/txjson"
	"github.com/stretchr/testify/assert"
)

func TestNode_Marshal(t *testing.T) {

}

func TestTx_JSON(t *testing.T) {
	tests := map[string]struct {
		tx  *bt.Tx
		err error
	}{
		"node standard tx should marshal and unmarshall correctly": {
			tx: func() *bt.Tx {
				tx := bt.NewTx()
				assert.NoError(t, tx.From(
					"3c8edde27cb9a9132c22038dac4391496be9db16fd21351565cc1006966fdad5",
					0,
					"76a914eb0bd5edba389198e73f8efabddfc61666969ff788ac",
					2000000,
				))
				assert.NoError(t, tx.PayToAddress("n2wmGVP89x3DsLNqk3NvctfQy9m9pvt7mk", 1000))
				var w *wif.WIF
				w, err := wif.DecodeWIF("KznvCNc6Yf4iztSThoMH6oHWzH9EgjfodKxmeuUGPq5DEX5maspS")
				assert.NoError(t, err)
				assert.NotNil(t, w)

				err = tx.SignAll(context.Background(), &bt.LocalSignerGetter{PrivateKey: w.PrivKey})
				assert.NoError(t, err)
				return tx
			}(),
		}, "node data tx should marshall correctly": {
			tx: func() *bt.Tx {
				tx := bt.NewTx()
				assert.NoError(t, tx.From(
					"3c8edde27cb9a9132c22038dac4391496be9db16fd21351565cc1006966fdad5",
					0,
					"76a914eb0bd5edba389198e73f8efabddfc61666969ff788ac",
					2000000,
				))
				assert.NoError(t, tx.PayToAddress("n2wmGVP89x3DsLNqk3NvctfQy9m9pvt7mk", 1000))
				var w *wif.WIF
				w, err := wif.DecodeWIF("KznvCNc6Yf4iztSThoMH6oHWzH9EgjfodKxmeuUGPq5DEX5maspS")
				assert.NoError(t, err)
				assert.NotNil(t, w)
				s := &bscript.Script{}
				assert.NoError(t, s.AppendPushDataString("test"))
				tx.AddOutput(&bt.Output{
					LockingScript: s,
				})
				err = tx.SignAll(context.Background(), &bt.LocalSignerGetter{PrivateKey: w.PrivKey})
				assert.NoError(t, err)
				return tx
			}(),
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			bb, err := json.Marshal(txjson.NodeWrapper(test.tx))
			assert.NoError(t, err)
			if err != nil {
				return
			}
			tx := &bt.Tx{}
			assert.NoError(t, json.Unmarshal(bb, txjson.NodeWrapper(tx)))
			assert.Equal(t, test.tx.String(), tx.String())
		})
	}
}
