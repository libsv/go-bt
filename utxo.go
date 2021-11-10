package bt

import (
	"encoding/hex"

	"github.com/libsv/go-bt/v2/bscript"
)

// UTXO an unspent transaction output, used for creating inputs
type UTXO struct {
	TxID           []byte
	Vout           uint32
	LockingScript  *bscript.Script
	Satoshis       uint64
	SequenceNumber uint32
}

// UTXOs a collection of *bt.UTXO.
type UTXOs []*UTXO

// NodeJSON returns a wrapped *bt.UTXO for marshalling/unmarshalling into a node utxo format.
//
// Marshalling usage example:
//  bb, err := json.Marshal(utxo.NodeJSON())
//
// Unmarshalling usage example:
//  utxo := &bt.UTXO{}
//  if err := json.Unmarshal(bb, utxo.NodeJSON()); err != nil {}
func (u *UTXO) NodeJSON() interface{} {
	return &nodeUTXOWrapper{UTXO: u}
}

// NodeJSON returns a wrapped bt.UTXOs for marshalling/unmarshalling into a node utxo format.
//
// Marshalling usage example:
//  bb, err := json.Marshal(utxos.NodeJSON())
//
// Unmarshalling usage example:
//  var txs bt.UTXOs
//  if err := json.Unmarshal(bb, utxos.NodeJSON()); err != nil {}
func (u *UTXOs) NodeJSON() interface{} {
	return (*nodeUTXOsWrapper)(u)
}

// TxIDStr return the tx id as a string.
func (u *UTXO) TxIDStr() string {
	return hex.EncodeToString(u.TxID)
}

// LockingScriptHexString retur nthe locking script in hex format.
func (u *UTXO) LockingScriptHexString() string {
	return u.LockingScript.String()
}
