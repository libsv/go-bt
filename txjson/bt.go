package txjson

import (
	"encoding/json"

	"github.com/libsv/go-bt/v2"
)

func BTWrapper(tx *bt.Tx) *txJSON {
	return &txJSON{Tx: tx}
}

type txJSON struct {
	*bt.Tx
	TxID string `json:"txid"`
	Hex  string `json:"hex"`
}

func (t *txJSON) MarshalJSON() ([]byte, error) {
	t.TxID = t.Tx.TxID()
	t.Hex = t.Tx.String()
	return json.Marshal(t)
}

func (t *txJSON) UnmarshalJSON(b []byte) error {
	return json.Unmarshal(b, &t.Tx)
}
