package bt_test

import (
	"encoding/hex"
	"encoding/json"
	"testing"

	"github.com/libsv/go-bt/v2"
	"github.com/libsv/go-bt/v2/bscript"
	"github.com/stretchr/testify/assert"
)

func TestUTXO_JSON(t *testing.T) {
	tests := map[string]struct {
		utxo *bt.UTXO
	}{
		"standard utxo should marshal and unmarshal correctly": {
			utxo: func() *bt.UTXO {
				txID, err := hex.DecodeString("31ad4b5ef1d0d48340e063087cbfa6a3f3dea3cd5d34c983e0028c18daf3d2a7")
				assert.NoError(t, err)
				script, err := bscript.NewFromHexString("2102076ad7c107f82ae973fbdaa1d84532c8d69e3838bcbee1570efe0fa30b3cb25bac")
				assert.NoError(t, err)
				return &bt.UTXO{
					TxID:          txID,
					LockingScript: script,
					Satoshis:      1250000000,
					Vout:          0,
				}
			}(),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			bb, err := json.Marshal(test.utxo)
			assert.NoError(t, err)

			var utxo *bt.UTXO
			assert.NoError(t, json.Unmarshal(bb, &utxo))

			bb2, err := json.Marshal(utxo)
			assert.NoError(t, err)
			assert.Equal(t, bb, bb2)
		})
	}
}

func TestUTXO_MarshalJSON(t *testing.T) {
	tests := map[string]struct {
		utxo *bt.UTXO
		exp  string
	}{
		"standard utxo should marshal correctly": {
			utxo: func() *bt.UTXO {
				txID, err := hex.DecodeString("31ad4b5ef1d0d48340e063087cbfa6a3f3dea3cd5d34c983e0028c18daf3d2a7")
				assert.NoError(t, err)
				script, err := bscript.NewFromHexString("2102076ad7c107f82ae973fbdaa1d84532c8d69e3838bcbee1570efe0fa30b3cb25bac")
				assert.NoError(t, err)
				return &bt.UTXO{
					TxID:          txID,
					LockingScript: script,
					Satoshis:      1250000000,
					Vout:          0,
				}
			}(),
			exp: `{
    "txid": "31ad4b5ef1d0d48340e063087cbfa6a3f3dea3cd5d34c983e0028c18daf3d2a7",
    "vout": 0,
    "lockingScript": "2102076ad7c107f82ae973fbdaa1d84532c8d69e3838bcbee1570efe0fa30b3cb25bac",
    "satoshis": 1250000000
}`,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			bb, err := json.MarshalIndent(test.utxo, "", "    ")
			assert.NoError(t, err)

			assert.Equal(t, test.exp, string(bb))
		})
	}
}

func TestUTXO_Node_JSON(t *testing.T) {
	tests := map[string]struct {
		utxo *bt.UTXO
	}{
		"node utxo should marshal and unmarshal correctly": {
			utxo: func() *bt.UTXO {
				txID, err := hex.DecodeString("31ad4b5ef1d0d48340e063087cbfa6a3f3dea3cd5d34c983e0028c18daf3d2a7")
				assert.NoError(t, err)
				script, err := bscript.NewFromHexString("2102076ad7c107f82ae973fbdaa1d84532c8d69e3838bcbee1570efe0fa30b3cb25bac")
				assert.NoError(t, err)
				return &bt.UTXO{
					TxID:          txID,
					LockingScript: script,
					Satoshis:      1250000000,
					Vout:          0,
				}
			}(),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			bb, err := json.Marshal(test.utxo.NodeJSON())
			assert.NoError(t, err)

			utxo := &bt.UTXO{}
			assert.NoError(t, json.Unmarshal(bb, utxo.NodeJSON()))

			bb2, err := json.Marshal(utxo.NodeJSON())
			assert.NoError(t, err)
			assert.Equal(t, bb, bb2)
		})
	}
}

func TestUTXO_Node_MarshalJSON(t *testing.T) {
	tests := map[string]struct {
		utxo *bt.UTXO
		exp  string
	}{
		"standard utxo should marshal correctly": {
			utxo: func() *bt.UTXO {
				txID, err := hex.DecodeString("31ad4b5ef1d0d48340e063087cbfa6a3f3dea3cd5d34c983e0028c18daf3d2a7")
				assert.NoError(t, err)
				script, err := bscript.NewFromHexString("2102076ad7c107f82ae973fbdaa1d84532c8d69e3838bcbee1570efe0fa30b3cb25bac")
				assert.NoError(t, err)
				return &bt.UTXO{
					TxID:          txID,
					LockingScript: script,
					Satoshis:      1250000000,
					Vout:          0,
				}
			}(),
			exp: `{
    "txid": "31ad4b5ef1d0d48340e063087cbfa6a3f3dea3cd5d34c983e0028c18daf3d2a7",
    "vout": 0,
    "scriptPubKey": "2102076ad7c107f82ae973fbdaa1d84532c8d69e3838bcbee1570efe0fa30b3cb25bac",
    "amount": 12.5
}`,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			bb, err := json.MarshalIndent(test.utxo.NodeJSON(), "", "    ")
			assert.NoError(t, err)

			assert.Equal(t, test.exp, string(bb))
		})
	}
}

func TestUTXO_Node_UnmarshalJSON(t *testing.T) {
	tests := map[string]struct {
		utxoJSON string
		expUTXO  *bt.UTXO
	}{
		"node json can be unmarshalled": {
			utxoJSON: `{
    "txid": "31ad4b5ef1d0d48340e063087cbfa6a3f3dea3cd5d34c983e0028c18daf3d2a7",
    "vout": 0,
    "scriptPubKey": "2102076ad7c107f82ae973fbdaa1d84532c8d69e3838bcbee1570efe0fa30b3cb25bac",
    "amount": 12.5
}`,
			expUTXO: func() *bt.UTXO {
				txID, err := hex.DecodeString("31ad4b5ef1d0d48340e063087cbfa6a3f3dea3cd5d34c983e0028c18daf3d2a7")
				assert.NoError(t, err)
				script, err := bscript.NewFromHexString("2102076ad7c107f82ae973fbdaa1d84532c8d69e3838bcbee1570efe0fa30b3cb25bac")
				assert.NoError(t, err)
				return &bt.UTXO{
					TxID:          txID,
					LockingScript: script,
					Satoshis:      1250000000,
					Vout:          0,
				}
			}(),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			var utxo bt.UTXO
			assert.NoError(t, json.Unmarshal([]byte(test.utxoJSON), utxo.NodeJSON()))

			assert.Equal(t, *test.expUTXO, utxo)
		})
	}
}
