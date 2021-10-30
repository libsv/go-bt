package bt

import (
	"encoding/json"
	"errors"

	"github.com/libsv/go-bt/v2/bscript"
)

type nodeTxWrapper struct {
	*Tx
}

type nodeTxsWrapper Txs

type nodeOutputWrapper struct {
	*Output
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

func (n *nodeTxWrapper) MarshalJSON() ([]byte, error) {
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
func (n *nodeTxWrapper) UnmarshalJSON(b []byte) error {
	tx := n.Tx

	var txj nodeTxJSON
	if err := json.Unmarshal(b, &txj); err != nil {
		return err
	}
	// quick convert
	if txj.Hex != "" {
		t, err := NewTxFromString(txj.Hex)
		if err != nil {
			return err
		}
		*tx = *t
		return nil
	}
	oo := make([]*Output, 0, len(txj.Outputs))
	for _, o := range txj.Outputs {
		out, err := o.toOutput()
		if err != nil {
			return err
		}
		oo = append(oo, out)
	}
	ii := make([]*Input, 0, len(txj.Inputs))
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

func (o *nodeOutputJSON) fromOutput(out *Output) error {
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

func (o *nodeOutputJSON) toOutput() (*Output, error) {
	out := &Output{}
	s, err := bscript.NewFromHexString(o.ScriptPubKey.Hex)
	if err != nil {
		return nil, err
	}
	out.Satoshis = uint64(o.Value * 100000000)
	out.LockingScript = s
	return out, nil
}

func (i *nodeInputJSON) toInput() (*Input, error) {
	input := &Input{}
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

func (i *nodeInputJSON) fromInput(input *Input) error {
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

// MarshalJSON will marshal a transaction that has been marshalled with this library.
func (nn nodeTxsWrapper) MarshalJSON() ([]byte, error) {
	txs := make([]*nodeTxWrapper, len(nn))
	for i, n := range nn {
		txs[i] = n.NodeJSON().(*nodeTxWrapper)
	}
	return json.Marshal(txs)
}

// UnmarshalJSON will unmarshal a transaction that has been marshalled with this library.
func (nn *nodeTxsWrapper) UnmarshalJSON(b []byte) error {
	var jj []json.RawMessage
	if err := json.Unmarshal(b, &jj); err != nil {
		return err
	}

	*nn = make(nodeTxsWrapper, 0)
	for _, j := range jj {
		tx := NewTx()
		if err := json.Unmarshal(j, tx.NodeJSON()); err != nil {
			return err
		}
		*nn = append(*nn, tx)
	}
	return nil
}

func (n *nodeOutputWrapper) MarshalJSON() ([]byte, error) {
	oj := &nodeOutputJSON{}
	oj.fromOutput(n.Output)
	return json.Marshal(oj)
}

func (n *nodeOutputWrapper) UnmarshalJSON(b []byte) error {
	oj := &nodeOutputJSON{}
	if err := json.Unmarshal(b, &oj); err != nil {
		return nil
	}

	o, err := oj.toOutput()
	if err != nil {
		return err
	}

	*n.Output = *o

	return nil
}
