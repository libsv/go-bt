package bt

import "github.com/libsv/go-bt/v2/bscript"

// UTXO an unspent transaction output, used for creating inputs
type UTXO struct {
	TxID          []byte
	Vout          uint32
	LockingScript *bscript.Script
	Satoshis      uint64
}
