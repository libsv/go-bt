package bt

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/libsv/go-bk/crypto"
)

/*
General format of a Bitcoin transaction (inside a block)
--------------------------------------------------------
Field            Description                                                               Size

Version no	     currently 1	                                                           4 bytes

In-counter  	 positive integer VI = VarInt                                              1 - 9 bytes

list of Inputs	 the first input of the first transaction is also called "coinbase"        <in-counter>-many Inputs
                 (its content was ignored in earlier versions)

Out-counter    	 positive integer VI = VarInt                                              1 - 9 bytes

list of Outputs  the Outputs of the first transaction spend the mined                      <out-counter>-many Outputs
								 bitcoins for the block

lock_time        if non-zero and sequence numbers are < 0xFFFFFFFF: block height or        4 bytes
                 timestamp when transaction is final
--------------------------------------------------------
*/

// Sentinel errors for transactions.
var (
	ErrInvalidTxID = errors.New("invalid TxID")
)

// Tx wraps a bitcoin transaction
//
// DO NOT CHANGE ORDER - Optimised memory via malign
//
type Tx struct {
	Inputs   []*Input
	Outputs  []*Output
	Version  uint32
	LockTime uint32
}

type txJSON struct {
	Version  uint32    `json:"version"`
	LockTime uint32    `json:"locktime"`
	TxID     string    `json:"txid"`
	Hash     string    `json:"hash"`
	Size     int       `json:"size"`
	Hex      string    `json:"hex"`
	Inputs   []*Input  `json:"vin"`
	Outputs  []*Output `json:"vout"`
}

// MarshalJSON will serialise a transaction to json.
func (tx *Tx) MarshalJSON() ([]byte, error) {
	if tx == nil {
		return nil, errors.New("tx is nil so cannot be marshalled")
	}
	for i, o := range tx.Outputs {
		o.index = i
	}
	txj := txJSON{
		Version:  tx.Version,
		LockTime: tx.LockTime,
		Inputs:   tx.Inputs,
		Outputs:  tx.Outputs,
		TxID:     tx.TxID(),
		Hash:     tx.TxID(),
		Size:     len(tx.Bytes()),
		Hex:      tx.String(),
	}
	return json.Marshal(txj)
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
	tx.Inputs = txj.Inputs
	tx.Outputs = txj.Outputs
	tx.LockTime = txj.LockTime
	tx.Version = txj.Version
	return nil
}

// NewTx creates a new transaction object with default values.
func NewTx() *Tx {
	return &Tx{Version: 1, LockTime: 0}
}

// NewTxFromString takes a toBytesHelper string representation of a bitcoin transaction
// and returns a Tx object.
func NewTxFromString(str string) (*Tx, error) {
	bb, err := hex.DecodeString(str)
	if err != nil {
		return nil, err
	}

	return NewTxFromBytes(bb)
}

// NewTxFromBytes takes an array of bytes, constructs a Tx and returns it.
// This function assumes that the byte slice contains exactly 1 transaction.
func NewTxFromBytes(b []byte) (*Tx, error) {
	tx, used, err := NewTxFromStream(b)
	if err != nil {
		return nil, err
	}

	if used != len(b) {
		return nil, fmt.Errorf("nLockTime length must be 4 bytes long")
	}

	return tx, nil
}

// NewTxFromStream takes an array of bytes and constructs a Tx from it, returning the Tx and the bytes used.
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

	// create Inputs
	var i uint64
	var err error
	var input *Input
	for ; i < inputCount; i++ {
		input, size, err = NewInputFromBytes(b[offset:])
		if err != nil {
			return nil, 0, err
		}
		offset += size
		t.addInput(input)
	}

	// create Outputs
	var outputCount uint64
	var output *Output
	outputCount, size = DecodeVarInt(b[offset:])
	offset += size
	for i = 0; i < outputCount; i++ {
		output, size, err = NewOutputFromBytes(b[offset:])
		if err != nil {
			return nil, 0, err
		}
		output.index = int(i)
		offset += size
		t.AddOutput(output)
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

// InputIdx will return the input at the specified index.
//
// This will consume an overflow error and simply return nil if the input
// isn't found at the index.
func (tx *Tx) InputIdx(i int) *Input {
	if i > tx.InputCount()-1 {
		return nil
	}
	return tx.Inputs[i]
}

// OutputIdx will return the output at the specified index.
//
// This will consume an overflow error and simply return nil if the output
// isn't found at the index.
func (tx *Tx) OutputIdx(i int) *Output {
	if i > tx.OutputCount()-1 {
		return nil
	}
	return tx.Outputs[i]
}

// IsCoinbase determines if this transaction is a coinbase by
// checking if the tx input is a standard coinbase input.
func (tx *Tx) IsCoinbase() bool {
	if len(tx.Inputs) != 1 {
		return false
	}

	cbi := make([]byte, 32)

	if !bytes.Equal(tx.Inputs[0].PreviousTxID(), cbi) {
		return false
	}

	if tx.Inputs[0].PreviousTxOutIndex == DefaultSequenceNumber || tx.Inputs[0].SequenceNumber == DefaultSequenceNumber {
		return true
	}

	return false
}

// TxIDBytes returns the transaction ID of the transaction as bytes
// (which is also the transaction hash).
func (tx *Tx) TxIDBytes() []byte {
	return ReverseBytes(crypto.Sha256d(tx.Bytes()))
}

// TxID returns the transaction ID of the transaction
// (which is also the transaction hash).
func (tx *Tx) TxID() string {
	return hex.EncodeToString(ReverseBytes(crypto.Sha256d(tx.Bytes())))
}

// String encodes the transaction into a hex string.
func (tx *Tx) String() string {
	return hex.EncodeToString(tx.Bytes())
}

// IsValidTxID will check that the txid bytes are valid.
//
// A txid should be of 32 bytes length.
func IsValidTxID(txid []byte) bool {
	return len(txid) == 32
}

// Bytes encodes the transaction into a byte array.
// See https://chainquery.com/bitcoin-cli/decoderawtransaction
func (tx *Tx) Bytes() []byte {
	return tx.toBytesHelper(0, nil)
}

// BytesWithClearedInputs encodes the transaction into a byte array but clears its Inputs first.
// This is used when signing transactions.
func (tx *Tx) BytesWithClearedInputs(index int, lockingScript []byte) []byte {
	return tx.toBytesHelper(index, lockingScript)
}

// Clone returns a clone of the tx
func (tx *Tx) Clone() *Tx {
	// Ignore err as byte slice passed in is created from valid tx
	clone, _ := NewTxFromBytes(tx.Bytes())

	for i, input := range tx.Inputs {
		clone.Inputs[i].PreviousTxSatoshis = input.PreviousTxSatoshis
		clone.Inputs[i].PreviousTxScript = input.PreviousTxScript
	}

	return clone
}

func (tx *Tx) toBytesHelper(index int, lockingScript []byte) []byte {
	h := make([]byte, 0)

	h = append(h, LittleEndianBytes(tx.Version, 4)...)

	h = append(h, VarInt(uint64(len(tx.Inputs)))...)

	for i, in := range tx.Inputs {
		s := in.Bytes(lockingScript != nil)
		if i == index && lockingScript != nil {
			h = append(h, VarInt(uint64(len(lockingScript)))...)
			h = append(h, lockingScript...)
		} else {
			h = append(h, s...)
		}
	}

	h = append(h, VarInt(uint64(len(tx.Outputs)))...)
	for _, out := range tx.Outputs {
		h = append(h, out.Bytes()...)
	}

	lt := make([]byte, 4)
	binary.LittleEndian.PutUint32(lt, tx.LockTime)

	return append(h, lt...)
}

// TxFees is returned when CalculateFee is called and contains
// a breakdown of the fees including the total.
type TxFees struct {
	Total       uint64
	StdFeeSats  uint64
	StdBytes    uint64
	DataFeeSats uint64
	DataBytes   uint64
}

// CalculateFees will calculate the fees required to cover this transaction and
// return with total and the individual fee types.
//
// There are a few conditions to be aware of:
//  - if the tx has not been signed, we will add 107 bytes for each unsigned input for the unlocking script
//  - if the tx has not yet had change added, we will check if change can be added, then add the bytes for the change output
func (tx *Tx) CalculateFees(fees *FeeQuote) (*TxFees, error) {
	inputAmount := tx.TotalInputSatoshis()
	outputAmount := tx.TotalOutputSatoshis()
	if inputAmount < outputAmount {
		return nil, errors.New("satoshis inputted to the tx are less than the outputted satoshis")
	}
	// calculate the base tx fees (before change)
	resp, err := tx.calculateTxFees(fees)
	if err != nil {
		return nil, err
	}
	stdFee, err := fees.Fee(FeeTypeStandard)
	if err != nil {
		return nil, err
	}
	// add change output if change can be added and a change output has not yet been added
	// if change has already been added, or we don't need change
	// as the inputs are exactly fee + outputs then we won't need
	// to add the bytes on. If we do need change added, we add the bytes.
	if change, ok := tx.canAddChange(resp, stdFee); ok {
		resp.StdBytes += 8 + 25 + 1 // 8 bytes for satoshi value + 25 bytes for p2pkh script (e.g. 76a914cc...05388ac) + 1 for varint
		resp.StdFeeSats += change
	}

	resp.Total = resp.DataFeeSats + resp.StdFeeSats
	return resp, nil
}

// calculateTxFees will calculate the base tx fees (without change)
func (tx *Tx) calculateTxFees(fees *FeeQuote) (*TxFees, error) {
	totBytes := len(tx.Bytes())
	// add (p2pkh) unlockingscript bytes for any inputs that haven't yet been signed.
	for _, in := range tx.Inputs {
		if !in.PreviousTxScript.IsP2PKH() {
			return nil, errors.New("non-P2PKH input used in the tx - unsupported")
		}
		if in.UnlockingScript == nil {
			totBytes += 107 // = 1 oppushdata + 70-71 sig + 1 sighash + 1 oppushdata + 33 public key
		}
	}

	// calculate data outputs
	dataLen := 0
	for _, d := range tx.Outputs {
		if d.LockingScript.IsData() {
			dataLen += len(*d.LockingScript)
		}
	}
	// get fees
	stdFee, err := fees.Fee(FeeTypeStandard)
	if err != nil {
		return nil, err
	}
	dataFee, err := fees.Fee(FeeTypeData)
	if err != nil {
		return nil, err
	}

	resp := &TxFees{
		StdFeeSats:  (uint64(totBytes) - uint64(dataLen)) * uint64(stdFee.MiningFee.Satoshis) / uint64(stdFee.MiningFee.Bytes),
		StdBytes:    uint64(totBytes) - uint64(dataLen),
		DataFeeSats: uint64(dataLen) * uint64(dataFee.MiningFee.Satoshis) / uint64(dataFee.MiningFee.Bytes),
		DataBytes:   uint64(dataLen),
	}
	resp.Total = resp.StdFeeSats + resp.DataFeeSats
	return resp, nil
}
