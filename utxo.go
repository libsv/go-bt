package bt

// UTXO an unspent transaction output, used for creating inputs
type UTXO struct {
	TxID          string
	Vout          uint32
	LockingScript string
	Satoshis      uint64
}
