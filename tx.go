package bt

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/libsv/go-bt/crypto"
)

/*
General format of a Bitcoin transaction (inside a block)
--------------------------------------------------------
Field            Description                                                               Size

Version no	     currently 1	                                                           4 bytes

In-counter  	 positive integer VI = VarInt                                              1 - 9 bytes

list of inputs	 the first input of the first transaction is also called "coinbase"        <in-counter>-many inputs
                 (its content was ignored in earlier versions)

Out-counter    	 positive integer VI = VarInt                                              1 - 9 bytes

list of outputs  the outputs of the first transaction spend the mined                      <out-counter>-many outputs
								 bitcoins for the block

lock_time        if non-zero and sequence numbers are < 0xFFFFFFFF: block height or        4 bytes
                 timestamp when transaction is final
--------------------------------------------------------
*/

// Tx wraps a bitcoin transaction
//
// DO NOT CHANGE ORDER - Optimized memory via malign
//
type Tx struct {
	Inputs   []*Input  `json:"vin"`
	Outputs  []*Output `json:"vout"`
	Version  uint32    `json:"version"`
	LockTime uint32    `json:"locktime"`
}

func (tx *Tx) MarshalJSON() ([]byte, error) {
	for i, o := range tx.Outputs {
		o.index = i
	}
	m := struct {
		Version  uint32    `json:"version"`
		LockTime uint32    `json:"locktime"`
		TxID     string    `json:"txid"`
		Hash     string    `json:"hash"`
		Size     int       `json:"size"`
		Hex      string    `json:"hex"`
		Inputs   []*Input  `json:"vin"`
		Outputs  []*Output `json:"vout"`
	}{
		Version:  tx.Version,
		LockTime: tx.LockTime,
		Inputs:   tx.Inputs,
		Outputs:  tx.Outputs,
		TxID:     tx.TxID(),
		Hash:     tx.TxID(),
		Size:     len(tx.ToBytes()),
		Hex:      tx.ToString(),
	}
	return json.Marshal(m)
}

// NewTx creates a new transaction object with default values.
func NewTx() *Tx {
	return &Tx{Version: 1, LockTime: 0}
}

// NewTxFromString takes a toBytesHelper string representation of a bitcoin transaction
// and returns a Tx object.
func NewTxFromString(str string) (*Tx, error) {
	bytes, err := hex.DecodeString(str)
	if err != nil {
		return nil, err
	}

	return NewTxFromBytes(bytes)
}

// NewTxFromBytes takes an array of bytes, constructs a Tx and returns it.
// This function assumes that the byte slice contains exactly 1 transaction.
func NewTxFromBytes(b []byte) (*Tx, error) {
	tx, used, err := NewTxFromStream((b))
	if err != nil {
		return nil, err
	}

	if used != len(b) {
		return nil, fmt.Errorf("nLockTime length must be 4 bytes long")
	}

	return tx, nil
}

// NewTxFromStream takes an array of bytes and contructs a Tx from it, returning the Tx and the bytes used.
// Despite the name, this is not actually reading a stream in the true sense: it is a byte slice that contains
// many transactions one after another.
func NewTxFromStream(b []byte) (*Tx, int, error) {

	if len(b) < 10 {
		return nil, 0, fmt.Errorf("too short to be a tx - even an empty tx has 10 bytes")
	}

	var offset int
	t := Tx{
		Version: binary.LittleEndian.Uint32(b[offset:4]),
	}
	offset += 4

	inputCount, size := DecodeVarInt(b[offset:])
	offset += size

	// create inputs
	var i uint64
	var err error
	var input *Input
	for ; i < inputCount; i++ {
		input, size, err = NewInputFromBytes(b[offset:])
		if err != nil {
			return nil, 0, err
		}
		offset += size

		t.Inputs = append(t.Inputs, input)
	}

	// create outputs
	var outputCount uint64
	var output *Output
	outputCount, size = DecodeVarInt(b[offset:])
	offset += size
	for i = 0; i < outputCount; i++ {
		output, size, err = NewOutputFromBytes(b[offset:])
		if err != nil {
			return nil, 0, err
		}
		offset += size
		t.Outputs = append(t.Outputs, output)
	}

	t.LockTime = binary.LittleEndian.Uint32(b[offset:])
	offset += 4

	return &t, offset, nil
}

// HasDataOutputs returns true if the transaction has
// at least one data (OP_RETURN) output in it.
func (tx *Tx) HasDataOutputs() bool {
	for _, out := range tx.Outputs {
		if out.LockingScript.IsData() {
			return true
		}
	}
	return false
}

// IsCoinbase determines if this transaction is a coinbase by
// checking if the tx input is a standard coinbase input.
func (tx *Tx) IsCoinbase() bool {
	if len(tx.Inputs) != 1 {
		return false
	}

	cbi := make([]byte, 32)

	if !bytes.Equal(tx.Inputs[0].PreviousTxIDBytes, cbi) {
		return false
	}

	if tx.Inputs[0].PreviousTxOutIndex == DefaultSequenceNumber || tx.Inputs[0].SequenceNumber == DefaultSequenceNumber {
		return true
	}

	return false
}

// TxIDAsBytes returns the transaction ID of the transaction as bytes
// (which is also the transaction hash).
func (tx *Tx) TxIDAsBytes() []byte {
	return ReverseBytes(crypto.Sha256d(tx.ToBytes()))
}

// TxID returns the transaction ID of the transaction
// (which is also the transaction hash).
func (tx *Tx) TxID() string {
	return hex.EncodeToString(ReverseBytes(crypto.Sha256d(tx.ToBytes())))
}

// ToString encodes the transaction into a hex string.
func (tx *Tx) ToString() string {
	return hex.EncodeToString(tx.ToBytes())
}

// ToBytes encodes the transaction into a byte array.
// See https://chainquery.com/bitcoin-cli/decoderawtransaction
func (tx *Tx) ToBytes() []byte {
	return tx.toBytesHelper(0, nil)
}

// ToBytesWithClearedInputs encodes the transaction into a byte array but clears its inputs first.
// This is used when signing transactions.
func (tx *Tx) ToBytesWithClearedInputs(index int, lockingScript []byte) []byte {
	return tx.toBytesHelper(index, lockingScript)
}

func (tx *Tx) toBytesHelper(index int, lockingScript []byte) []byte {
	h := make([]byte, 0)

	h = append(h, LittleEndianBytes(tx.Version, 4)...)

	h = append(h, VarInt(uint64(len(tx.Inputs)))...)

	for i, in := range tx.Inputs {
		s := in.ToBytes(lockingScript != nil)
		if i == index && lockingScript != nil {
			h = append(h, VarInt(uint64(len(lockingScript)))...)
			h = append(h, lockingScript...)
		} else {
			h = append(h, s...)
		}
	}

	h = append(h, VarInt(uint64(len(tx.Outputs)))...)
	for _, out := range tx.Outputs {
		h = append(h, out.ToBytes()...)
	}

	lt := make([]byte, 4)
	binary.LittleEndian.PutUint32(lt, tx.LockTime)

	return append(h, lt...)
}
