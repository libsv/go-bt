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

type UTXOs []*UTXO

func (u *UTXO) NodeJSON() interface{} {
	return &nodeUTXOWrapper{UTXO: u}
}

func (u *UTXOs) NodeJSON() interface{} {
	return (*nodeUTXOsWrapper)(u)
}

func (u *UTXO) TxIDStr() string {
	return hex.EncodeToString(u.TxID)
}

func (u *UTXO) LockingScriptHexString() string {
	return u.LockingScript.String()
}
