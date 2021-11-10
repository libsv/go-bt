package bt

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"

	"github.com/libsv/go-bk/crypto"

	"github.com/libsv/go-bt/v2/bscript"
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

// Txs a collection of *bt.Tx.
type Txs []*Tx

// NewTx creates a new transaction object with default values.
func NewTx() *Tx {
	return &Tx{Version: 1, LockTime: 0, Inputs: make([]*Input, 0)}
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
		return nil, ErrNLockTimeLength
	}

	return tx, nil
}

// NewTxFromStream takes an array of bytes and constructs a Tx from it, returning the Tx and the bytes used.
// Despite the name, this is not actually reading a stream in the true sense: it is a byte slice that contains
// many transactions one after another.
func NewTxFromStream(b []byte) (*Tx, int, error) {
	if len(b) < 10 {
		return nil, 0, ErrTxTooShort
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
		input, size, err = newInputFromBytes(b[offset:])
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
		output, size, err = newOutputFromBytes(b[offset:])
		if err != nil {
			return nil, 0, err
		}
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

// NodeJSON returns a wrapped *bt.Tx for marshalling/unmarshalling into a node tx format.
//
// Marshalling usage example:
//  bb, err := json.Marshal(tx.NodeJSON())
//
// Unmarshalling usage example:
//  tx := bt.NewTx()
//  if err := json.Unmarshal(bb, tx.NodeJSON()); err != nil {}
func (tx *Tx) NodeJSON() interface{} {
	return &nodeTxWrapper{Tx: tx}
}

// NodeJSON returns a wrapped bt.Txs for marshalling/unmarshalling into a node tx format.
//
// Marshalling usage example:
//  bb, err := json.Marshal(txs.NodeJSON())
//
// Unmarshalling usage example:
//  var txs bt.Txs
//  if err := json.Unmarshal(bb, txs.NodeJSON()); err != nil {}
func (tt *Txs) NodeJSON() interface{} {
	return (*nodeTxsWrapper)(tt)
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

// TxSize contains the size breakdown of a transaction
// including the breakdown of data bytes vs standard bytes.
// This information can be used when calculating fees.
type TxSize struct {
	// TotalBytes are the amount of bytes for the entire tx.
	TotalBytes uint64
	// TotalStdBytes are the amount of bytes for the tx minus the data bytes.
	TotalStdBytes uint64
	// TotalDataBytes is the size in bytes of the op_return / data outputs.
	TotalDataBytes uint64
}

// Size will return the size of tx in bytes.
func (tx *Tx) Size() int {
	return len(tx.Bytes())
}

// SizeWithTypes will return the size of tx in bytes
// and include the different data types (std/data/etc.).
func (tx *Tx) SizeWithTypes() *TxSize {
	totBytes := tx.Size()

	// calculate data outputs
	dataLen := 0
	for _, d := range tx.Outputs {
		if d.LockingScript.IsData() {
			dataLen += len(*d.LockingScript)
		}
	}
	return &TxSize{
		TotalBytes:     uint64(totBytes),
		TotalStdBytes:  uint64(totBytes - dataLen),
		TotalDataBytes: uint64(dataLen),
	}
}

// EstimateSize will return the size of tx in bytes and will add 107 bytes
// to the unlocking script of any unsigned inputs (only P2PKH for now) found
// to give a final size estimate of the tx size.
func (tx *Tx) EstimateSize() (int, error) {
	tempTx, err := tx.estimatedFinalTx()
	if err != nil {
		return 0, err
	}

	return tempTx.Size(), nil
}

// EstimateSizeWithTypes will return the size of tx in bytes, including the
// different data types (std/data/etc.), and will add 107 bytes to the unlocking
// script of any unsigned inputs (only P2PKH for now) found to give a final size
// estimate of the tx size.
func (tx *Tx) EstimateSizeWithTypes() (*TxSize, error) {
	tempTx, err := tx.estimatedFinalTx()
	if err != nil {
		return nil, err
	}

	return tempTx.SizeWithTypes(), nil
}

func (tx *Tx) estimatedFinalTx() (*Tx, error) {
	tempTx := tx.Clone()

	for _, in := range tempTx.Inputs {
		if !in.PreviousTxScript.IsP2PKH() {
			return nil, ErrUnsupportedScript
		}
		if in.UnlockingScript == nil || len(*in.UnlockingScript) == 0 {
			// nolint:lll // insert dummy p2pkh unlocking script (sig + pubkey)
			dummyUnlockingScript, _ := hex.DecodeString("4830450221009c13cbcbb16f2cfedc7abf3a4af1c3fe77df1180c0e7eee30d9bcc53ebda39da02207b258005f1bc3cf9dffa06edb358d6db2bcfc87f50516fac8e3f4686fc2a03df412103107feff22788a1fc8357240bf450fd7bca4bd45d5f8bac63818c5a7b67b03876")
			in.UnlockingScript = bscript.NewFromBytes(dummyUnlockingScript)
		}
	}
	return tempTx, nil
}

// TxFees is returned when CalculateFee is called and contains
// a breakdown of the fees including the total and the size breakdown of
// the tx in bytes.
type TxFees struct {
	// TotalFeePaid is the total amount of fees this tx will pay.
	TotalFeePaid uint64
	// StdFeePaid is the amount of fee to cover the standard inputs and outputs etc.
	StdFeePaid uint64
	// DataFeePaid is the amount of fee to cover the op_return data outputs.
	DataFeePaid uint64
}

// IsFeePaidEnough will calculate the fees that this transaction is paying
// including the individual fee types (std/data/etc.).
func (tx *Tx) IsFeePaidEnough(fees *FeeQuote) (bool, error) {
	expFeesPaid, err := tx.feesPaid(tx.SizeWithTypes(), fees)
	if err != nil {
		return false, err
	}
	totalInputSatoshis := tx.TotalInputSatoshis()
	totalOutputSatoshis := tx.TotalOutputSatoshis()

	if totalInputSatoshis < totalOutputSatoshis {
		return false, nil
	}

	actualFeePaid := totalInputSatoshis - totalOutputSatoshis
	return actualFeePaid >= expFeesPaid.TotalFeePaid, nil
}

// EstimateIsFeePaidEnough will calculate the fees that this transaction is paying
// including the individual fee types (std/data/etc.), and will add 107 bytes to the unlocking
// script of any unsigned inputs (only P2PKH for now) found to give a final size
// estimate of the tx size for fee calculation.
func (tx *Tx) EstimateIsFeePaidEnough(fees *FeeQuote) (bool, error) {
	tempTx, err := tx.estimatedFinalTx()
	if err != nil {
		return false, err
	}
	expFeesPaid, err := tempTx.feesPaid(tempTx.SizeWithTypes(), fees)
	if err != nil {
		return false, err
	}
	totalInputSatoshis := tempTx.TotalInputSatoshis()
	totalOutputSatoshis := tempTx.TotalOutputSatoshis()

	if totalInputSatoshis < totalOutputSatoshis {
		return false, nil
	}

	actualFeePaid := totalInputSatoshis - totalOutputSatoshis
	return actualFeePaid >= expFeesPaid.TotalFeePaid, nil
}

// EstimateFeesPaid will estimate how big the tx will be when finalised
// by estimating input unlocking scripts that have not yet been filled
// including the individual fee types (std/data/etc.).
func (tx *Tx) EstimateFeesPaid(fees *FeeQuote) (*TxFees, error) {
	size, err := tx.EstimateSizeWithTypes()
	if err != nil {
		return nil, err
	}
	return tx.feesPaid(size, fees)
}

func (tx *Tx) feesPaid(size *TxSize, fees *FeeQuote) (*TxFees, error) {
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
		StdFeePaid:  size.TotalStdBytes * uint64(stdFee.MiningFee.Satoshis) / uint64(stdFee.MiningFee.Bytes),
		DataFeePaid: size.TotalDataBytes * uint64(dataFee.MiningFee.Satoshis) / uint64(dataFee.MiningFee.Bytes),
	}
	resp.TotalFeePaid = resp.StdFeePaid + resp.DataFeePaid
	return resp, nil

}

func (tx *Tx) estimateDeficit(fees *FeeQuote) (uint64, error) {
	totalInputSatoshis := tx.TotalInputSatoshis()
	totalOutputSatoshis := tx.TotalOutputSatoshis()

	expFeesPaid, err := tx.EstimateFeesPaid(fees)
	if err != nil {
		return 0, err
	}

	if totalInputSatoshis > totalOutputSatoshis+expFeesPaid.TotalFeePaid {
		return 0, nil
	}

	return totalOutputSatoshis + expFeesPaid.TotalFeePaid - totalInputSatoshis, nil
}
