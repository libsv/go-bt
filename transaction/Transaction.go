package transaction

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"

	"github.com/libsv/libsv/crypto"
	"github.com/libsv/libsv/script"
	"github.com/libsv/libsv/transaction/input"
	"github.com/libsv/libsv/transaction/output"
	"github.com/libsv/libsv/utils"
)

/*
General format of a Bitcoin transaction (inside a block)
--------------------------------------------------------
Field            Description                                                               Size

Version no	     currently 1	                                                             4 bytes

In-counter  	   positive integer VI = VarInt                                              1 - 9 bytes

list of inputs	 the first input of the first transaction is also called "coinbase"        <in-counter>-many inputs
                 (its content was ignored in earlier versions)

Out-counter    	 positive integer VI = VarInt                                              1 - 9 bytes

list of outputs  the outputs of the first transaction spend the mined                      <out-counter>-many outputs
								 bitcoins for the block

lock_time        if non-zero and sequence numbers are < 0xFFFFFFFF: block height or        4 bytes
                 timestamp when transaction is final
*/

// A Transaction wraps a bitcoin transaction
type Transaction struct {
	Version  uint32
	Inputs   []*input.Input
	Outputs  []*output.Output
	Locktime uint32
}

// New creates a new transaction object with default values.
func New() *Transaction {
	t := Transaction{}

	t.Version = 1
	t.Locktime = 0

	return &t
}

// NewFromString takes a toBytesHelper string representation of a bitcoin transaction
// and returns a Transaction object.
func NewFromString(str string) (*Transaction, error) {
	bytes, err := hex.DecodeString(str)
	if err != nil {
		return nil, err
	}

	return NewFromBytes(bytes), nil
}

// NewFromBytes takes an array of bytes, constructs a Transaction and returns it.
func NewFromBytes(b []byte) *Transaction {
	if len(b) < 10 {
		return nil // Even an empty transaction has 10 bytes.
	}

	t := Transaction{}

	var offset = 0

	t.Version = binary.LittleEndian.Uint32(b[offset:4])
	offset += 4

	inputCount, size := utils.DecodeVarInt(b[offset:])
	offset += size

	// create inputs
	var i uint64
	for ; i < inputCount; i++ {
		i, size := input.NewFromBytes(b[offset:])
		offset += size

		t.Inputs = append(t.Inputs, i)
	}

	// create outputs
	outputCount, size := utils.DecodeVarInt(b[offset:])
	offset += size
	for i = 0; i < outputCount; i++ {
		o, size := output.NewFromBytes(b[offset:])
		offset += size
		t.Outputs = append(t.Outputs, o)
	}

	t.Locktime = binary.LittleEndian.Uint32(b[offset:])
	offset += 4

	return &t
}

// AddInput adds a new input to the transaction.
func (bt *Transaction) AddInput(input *input.Input) {
	bt.Inputs = append(bt.Inputs, input)
}

// From adds a new input to the transaction from the specified UTXO fields.
func (bt *Transaction) From(txID string, vout uint32, scriptSig string, satoshis uint64) error {
	pts, err := script.NewFromHexString(scriptSig)
	if err != nil {
		return err
	}

	i := &input.Input{
		PreviousTxOutIndex: vout,
		PreviousTxScript:   pts,
		PreviousTxSatoshis: satoshis,
		SequenceNumber:     0xffffffff,
	}

	i.PreviousTxID = txID

	bt.AddInput(i)

	return nil
}

// InputCount returns the number of transaction inputs.
func (bt *Transaction) InputCount() int {
	return len(bt.Inputs)
}

// OutputCount returns the number of transaction inputs.
func (bt *Transaction) OutputCount() int {
	return len(bt.Outputs)
}

// AddOutput adds a new output to the transaction.
func (bt *Transaction) AddOutput(output *output.Output) {

	bt.Outputs = append(bt.Outputs, output)
}

// PayTo function
func (bt *Transaction) PayTo(addr string, satoshis uint64) error {
	o, err := output.NewP2PKHFromAddress(addr, satoshis)
	if err != nil {
		return err
	}

	bt.AddOutput(o)
	return nil
}

// IsCoinbase determines if this transaction is a coinbase by
// checking if the tx input is a standard coinbase input.
func (bt *Transaction) IsCoinbase() bool {
	if len(bt.Inputs) != 1 {
		return false
	}

	if bt.Inputs[0].PreviousTxID != "0000000000000000000000000000000000000000000000000000000000000000" {
		return false
	}

	if bt.Inputs[0].PreviousTxOutIndex == 0xFFFFFFFF || bt.Inputs[0].SequenceNumber == 0xFFFFFFFF {
		return true
	}

	return false
}

// GetInputs returns an array of all inputs in the transaction.
func (bt *Transaction) GetInputs() []*input.Input {
	return bt.Inputs
}

// GetOutputs returns an array of all outputs in the transaction.
func (bt *Transaction) GetOutputs() []*output.Output {
	return bt.Outputs
}

// GetTotalOutputSatoshis returns an array of all outputs in the transaction.
func (bt *Transaction) GetTotalOutputSatoshis() uint64 {
	var total uint64
	for _, o := range bt.GetOutputs() {
		total += o.Satoshis
	}

	return total
}

// GetTxID returns the transaction ID of the transaction
// (which is also the transaction hash).
func (bt *Transaction) GetTxID() string {
	return hex.EncodeToString(utils.ReverseBytes(crypto.Sha256d(bt.ToBytes())))
}

// ToString encodes the transaction into a hex string.
func (bt *Transaction) ToString() string {
	return hex.EncodeToString(bt.ToBytes())
}

// ToBytes encodes the transaction into a byte array.
// See https://chainquery.com/bitcoin-cli/decoderawtransaction
func (bt *Transaction) ToBytes() []byte {
	return bt.toBytesHelper(0, nil)
}

// ToBytesWithClearedInputs encodes the transaction into a byte array but clears its inputs first.
// This is used when signing transactions.
func (bt *Transaction) ToBytesWithClearedInputs(index int, scriptPubKey []byte) []byte {
	return bt.toBytesHelper(index, scriptPubKey)
}

func (bt *Transaction) toBytesHelper(index int, scriptPubKey []byte) []byte {
	h := make([]byte, 0)

	h = append(h, utils.GetLittleEndianBytes(bt.Version, 4)...)

	h = append(h, utils.VarInt(uint64(len(bt.GetInputs())))...)

	for i, in := range bt.GetInputs() {
		s := in.ToBytes(scriptPubKey != nil)
		if i == index && scriptPubKey != nil {
			h = append(h, utils.VarInt(uint64(len(scriptPubKey)))...)
			h = append(h, scriptPubKey...)
		} else {
			h = append(h, s...)
		}
	}

	h = append(h, utils.VarInt(uint64(len(bt.GetOutputs())))...)
	for _, out := range bt.GetOutputs() {
		h = append(h, out.ToBytes()...)
	}

	lt := make([]byte, 4)
	binary.LittleEndian.PutUint32(lt, bt.Locktime)
	h = append(h, lt...)

	return h
}

// Sign is used to sign the transaction at a specific input index.
// It takes a Signed interface as a parameter so that different
// signing implementations can be used to sign the transaction -
// for example internal/local or external signing.
func (bt *Transaction) Sign(index uint32, s Signer) error {
	signedTx, err := s.Sign(index, bt)
	if err != nil {
		return err
	}
	*bt = *signedTx
	return nil
}

// SignAuto is used to automatically check which P2PKH inputs are
// able to be signed (match the public key) and then sign them.
// It takes a Signed interface as a parameter so that different
// signing implementations can be used to sign the transaction -
// for example internal/local or external signing.
func (bt *Transaction) SignAuto(s Signer) error {
	signedTx, err := s.SignAuto(bt)
	if err != nil {
		return err
	}
	*bt = *signedTx
	return nil
}

func (bt *Transaction) ApplyUnlockingScript(index uint32, s *script.Script) error {
	if bt.Inputs[index] != nil {
		bt.Inputs[index].UnlockingScript = s
		return nil
	}

	return fmt.Errorf("no input at index %d", index)
}
