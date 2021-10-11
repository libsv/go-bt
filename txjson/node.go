package txjson

import (
	"encoding/json"
	"errors"

	"github.com/libsv/go-bt/v2"
	"github.com/libsv/go-bt/v2/bscript"
)

type nodeWrapper struct {
	*bt.Tx
}

func NodeWrapper(tx *bt.Tx) *nodeWrapper {
	return &nodeWrapper{tx}
}

type nodeTxJSON struct {
	Version  uint32            `json:"version"`
	LockTime uint32            `json:"locktime"`
	TxID     string            `json:"txid"`
	Hash     string            `json:"hash"`
	Size     int               `json:"size"`
	Hex      string            `json:"hex"`
	Inputs   []*bt.Input       `json:"vin"`
	Outputs  []*nodeOutputJSON `json:"vout"`
}

type nodeOutputJSON struct {
	Value        float64 `json:"value"`
	Satoshis     uint64  `json:"satoshis"`
	Index        int     `json:"n"`
	ScriptPubKey *struct {
		Asm     string `json:"asm"`
		Hex     string `json:"hex"`
		ReqSigs int    `json:"reqSigs,omitempty"`
		Type    string `json:"type"`
	} `json:"scriptPubKey,omitempty"`
	LockingScript *struct {
		Asm     string `json:"asm"`
		Hex     string `json:"hex"`
		ReqSigs int    `json:"reqSigs,omitempty"`
		Type    string `json:"type"`
	} `json:"lockingScript,omitempty"`
}

func (n *nodeWrapper) MarshalJSON() ([]byte, error) {
	tx := n.Tx
	if n == nil {
		return nil, errors.New("tx is nil so cannot be marshalled")
	}
	oo := make([]*nodeOutputJSON, 0, len(tx.Outputs))
	for i, o := range tx.Outputs {
		out, err := outputToNodeJSON(o)
		if err != nil {
			return nil, err
		}
		out.Index = i
		oo = append(oo, out)
	}
	txj := nodeTxJSON{
		Version:  tx.Version,
		LockTime: tx.LockTime,
		Inputs:   tx.Inputs,
		Outputs:  oo,
		TxID:     tx.TxID(),
		Hash:     tx.TxID(),
		Size:     len(tx.Bytes()),
		Hex:      tx.String(),
	}
	return json.Marshal(txj)
}

// UnmarshalJSON will unmarshall a transaction that has been marshalled with this library.
func (n *nodeWrapper) UnmarshalJSON(b []byte) error {
	tx := n.Tx

	var txj nodeTxJSON
	if err := json.Unmarshal(b, &txj); err != nil {
		return err
	}
	// quick convert
	if txj.Hex != "" {
		t, err := bt.NewTxFromString(txj.Hex)
		if err != nil {
			return err
		}
		*tx = *t
		return nil
	}
	oo := make([]*bt.Output, 0, len(txj.Outputs))
	for _, o := range txj.Outputs {
		out, err := o.toOutput()
		if err != nil {
			return err
		}
		oo = append(oo, out)
	}
	tx.Inputs = txj.Inputs
	tx.Outputs = oo
	tx.LockTime = txj.LockTime
	tx.Version = txj.Version
	return nil
}

func outputToNodeJSON(o *bt.Output) (*nodeOutputJSON, error) {
	asm, err := o.LockingScript.ToASM()
	if err != nil {
		return nil, err
	}
	addresses, err := o.LockingScript.Addresses()
	if err != nil {
		return nil, err
	}

	return &nodeOutputJSON{
		Value:    float64(o.Satoshis) / 100000000,
		Satoshis: o.Satoshis,
		Index:    0,
		LockingScript: &struct {
			Asm     string `json:"asm"`
			Hex     string `json:"hex"`
			ReqSigs int    `json:"reqSigs,omitempty"`
			Type    string `json:"type"`
		}{
			Asm:     asm,
			Hex:     o.LockingScriptHexString(),
			ReqSigs: len(addresses),
			Type:    o.LockingScript.ScriptType(),
		},
	}, nil
}

func (o *nodeOutputJSON) fromOutput(out *bt.Output) error {
	asm, err := out.LockingScript.ToASM()
	if err != nil {
		return err
	}
	addresses, err := out.LockingScript.Addresses()
	if err != nil {
		return err
	}

	*o = nodeOutputJSON{
		Value:    float64(o.Satoshis) / 100000000,
		Satoshis: o.Satoshis,
		Index:    0,
		LockingScript: &struct {
			Asm     string `json:"asm"`
			Hex     string `json:"hex"`
			ReqSigs int    `json:"reqSigs,omitempty"`
			Type    string `json:"type"`
		}{
			Asm:     asm,
			Hex:     out.LockingScriptHexString(),
			ReqSigs: len(addresses),
			Type:    out.LockingScript.ScriptType(),
		},
	}

	return nil
}

func (o *nodeOutputJSON) toOutput() (*bt.Output, error) {
	out := &bt.Output{}
	script := o.LockingScript
	if script == nil {
		script = o.ScriptPubKey
	}
	s, err := bscript.NewFromHexString(script.Hex)
	if err != nil {
		return nil, err
	}
	if o.Satoshis > 0 {
		out.Satoshis = o.Satoshis
	} else {
		out.Satoshis = uint64(o.Value * 100000000)
	}
	out.LockingScript = s
	return out, nil
}
