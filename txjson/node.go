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

// Node a json wrapper for *bt.Tx, used for marshalling and unmarshalling json
// to/from node format.
//
// Marshal example:
//  bb, err := json.Marshal(txjson.Node(tx))
//
// Unmarshal example:
//  tx := bt.NewTx()
//  if err := json.Unmarshal(bb, txjson.Node(tx)); err != nil {}
func Node(tx *bt.Tx) *nodeWrapper { // nolint:revive // We don't want this type to be used outside of json formatting.
	return &nodeWrapper{tx}
}

type nodeTxJSON struct {
	Version  uint32            `json:"version"`
	LockTime uint32            `json:"locktime"`
	TxID     string            `json:"txid"`
	Hash     string            `json:"hash"`
	Size     int               `json:"size"`
	Hex      string            `json:"hex"`
	Inputs   []*nodeInputJSON  `json:"vin"`
	Outputs  []*nodeOutputJSON `json:"vout"`
}

type nodeInputJSON struct {
	ScriptSig *struct {
		Asm string `json:"asm"`
		Hex string `json:"hex"`
	} `json:"scriptSig,omitempty"`
	TxID     string `json:"txid"`
	Vout     uint32 `json:"vout"`
	Sequence uint32 `json:"sequence"`
}

type nodeOutputJSON struct {
	Value        float64 `json:"value"`
	Index        int     `json:"n"`
	ScriptPubKey *struct {
		Asm     string `json:"asm"`
		Hex     string `json:"hex"`
		ReqSigs int    `json:"reqSigs,omitempty"`
		Type    string `json:"type"`
	} `json:"scriptPubKey,omitempty"`
}

func (n *nodeWrapper) MarshalJSON() ([]byte, error) {
	if n == nil || n.Tx == nil {
		return nil, errors.New("tx is nil so cannot be marshalled")
	}
	tx := n.Tx
	oo := make([]*nodeOutputJSON, 0, len(tx.Outputs))
	for i, o := range tx.Outputs {
		out := &nodeOutputJSON{}
		if err := out.fromOutput(o); err != nil {
			return nil, err
		}
		out.Index = i
		oo = append(oo, out)
	}
	ii := make([]*nodeInputJSON, 0, len(tx.Inputs))
	for _, i := range tx.Inputs {
		in := &nodeInputJSON{}
		if err := in.fromInput(i); err != nil {
			return nil, err
		}
		ii = append(ii, in)
	}
	txj := nodeTxJSON{
		Version:  tx.Version,
		LockTime: tx.LockTime,
		Inputs:   ii,
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
	ii := make([]*bt.Input, 0, len(txj.Inputs))
	for _, i := range txj.Inputs {
		in, err := i.toInput()
		if err != nil {
			return err
		}
		ii = append(ii, in)
	}
	tx.Inputs = ii
	tx.Outputs = oo
	tx.LockTime = txj.LockTime
	tx.Version = txj.Version
	return nil
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
		Value: float64(out.Satoshis) / 100000000,
		Index: 0,
		ScriptPubKey: &struct {
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
	s, err := bscript.NewFromHexString(o.ScriptPubKey.Hex)
	if err != nil {
		return nil, err
	}
	out.Satoshis = uint64(o.Value * 100000000)
	out.LockingScript = s
	return out, nil
}

func (i *nodeInputJSON) toInput() (*bt.Input, error) {
	input := &bt.Input{}
	s, err := bscript.NewFromHexString(i.ScriptSig.Hex)
	if err != nil {
		return nil, err
	}

	input.UnlockingScript = s
	input.PreviousTxOutIndex = i.Vout
	input.SequenceNumber = i.Sequence
	if err = input.PreviousTxIDAddStr(i.TxID); err != nil {
		return nil, err
	}

	return input, nil
}

func (i *nodeInputJSON) fromInput(input *bt.Input) error {
	asm, err := input.UnlockingScript.ToASM()
	if err != nil {
		return err
	}

	i.ScriptSig = &struct {
		Asm string `json:"asm"`
		Hex string `json:"hex"`
	}{
		Asm: asm,
		Hex: input.UnlockingScript.String(),
	}

	i.Vout = input.PreviousTxOutIndex
	i.Sequence = input.SequenceNumber
	i.TxID = input.PreviousTxIDStr()

	return nil
}
