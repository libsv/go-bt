package bt

import (
	"encoding/hex"
	"encoding/json"

	"github.com/libsv/go-bt/v2/bscript"
	"github.com/pkg/errors"
)

type txJSON struct {
	TxID     string    `json:"txid"`
	Hex      string    `json:"hex"`
	Inputs   []*Input  `json:"inputs"`
	Outputs  []*Output `json:"outputs"`
	Version  uint32    `json:"version"`
	LockTime uint32    `json:"lockTime"`
}

type inputJSON struct {
	UnlockingScript string `json:"unlockingScript"`
	TxID            string `json:"txid"`
	Vout            uint32 `json:"vout"`
	Sequence        uint32 `json:"sequence"`
}

type outputJSON struct {
	Satoshis      uint64 `json:"satoshis"`
	LockingScript string `json:"lockingScript"`
}

// MarshalJSON will serialise a transaction to json.
func (tx *Tx) MarshalJSON() ([]byte, error) {
	if tx == nil {
		return nil, errors.Wrap(ErrTxNil, "cannot marshal tx")
	}
	return json.Marshal(txJSON{
		TxID:     tx.TxID(),
		Hex:      tx.String(),
		Inputs:   tx.Inputs,
		Outputs:  tx.Outputs,
		LockTime: tx.LockTime,
		Version:  tx.Version,
	})
}

// UnmarshalJSON will unmarshall a transaction that has been marshalled with this library.
func (tx *Tx) UnmarshalJSON(b []byte) error {
	var txj txJSON
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
	tx.LockTime = txj.LockTime
	tx.Version = txj.Version
	return nil
}

// MarshalJSON will convert an input to json, expanding upon the
// input struct to add additional fields.
func (i *Input) MarshalJSON() ([]byte, error) {
	return json.Marshal(&inputJSON{
		TxID:            hex.EncodeToString(i.previousTxID),
		Vout:            i.PreviousTxOutIndex,
		UnlockingScript: i.UnlockingScript.String(),
		Sequence:        i.SequenceNumber,
	})
}

// UnmarshalJSON will convert a JSON input to an input.
func (i *Input) UnmarshalJSON(b []byte) error {
	var ij inputJSON
	if err := json.Unmarshal(b, &ij); err != nil {
		return err
	}
	ptxID, err := hex.DecodeString(ij.TxID)
	if err != nil {
		return err
	}
	s, err := bscript.NewFromHexString(ij.UnlockingScript)
	if err != nil {
		return err
	}
	i.UnlockingScript = s
	i.previousTxID = ptxID
	i.PreviousTxOutIndex = ij.Vout
	i.SequenceNumber = ij.Sequence
	return nil
}

// MarshalJSON will serialise an output to json.
func (o *Output) MarshalJSON() ([]byte, error) {
	return json.Marshal(&outputJSON{
		Satoshis:      o.Satoshis,
		LockingScript: o.LockingScriptHexString(),
	})
}

// UnmarshalJSON will convert a json serialised output to a bt Output.
func (o *Output) UnmarshalJSON(b []byte) error {
	var oj outputJSON
	if err := json.Unmarshal(b, &oj); err != nil {
		return err
	}
	s, err := bscript.NewFromHexString(oj.LockingScript)
	if err != nil {
		return err
	}
	o.Satoshis = oj.Satoshis
	o.LockingScript = s
	return nil
}
