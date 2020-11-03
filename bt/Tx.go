package bt

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"

	mapi "github.com/bitcoin-sv/merchantapi-reference/utils"
	"github.com/libsv/libsv/bt/fees"
	"github.com/libsv/libsv/bt/input"
	"github.com/libsv/libsv/bt/output"
	"github.com/libsv/libsv/crypto"
	"github.com/libsv/libsv/script"
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

// A Tx wraps a bitcoin transaction
type Tx struct {
	// TODO: make variables private?
	Version  uint32
	Inputs   []*input.Input
	Outputs  []*output.Output
	Locktime uint32
}

// New creates a new transaction object with default values.
func New() *Tx {
	t := Tx{}

	t.Version = 1
	t.Locktime = 0

	return &t
}

// NewFromString takes a toBytesHelper string representation of a bitcoin transaction
// and returns a Tx object.
func NewFromString(str string) (*Tx, error) {
	bytes, err := hex.DecodeString(str)
	if err != nil {
		return nil, err
	}

	return NewFromBytes(bytes)
}

// NewFromBytes takes an array of bytes, constructs a Tx and returns it.
func NewFromBytes(b []byte) (*Tx, error) {
	if len(b) < 10 {
		return nil, fmt.Errorf("too short to be a tx - even an empty tx has 10 bytes")
	}

	t := Tx{}

	var offset = 0

	t.Version = binary.LittleEndian.Uint32(b[offset:4])
	offset += 4

	inputCount, size := utils.DecodeVarInt(b[offset:])
	offset += size

	// create inputs
	var i uint64
	for ; i < inputCount; i++ {
		i, size, err := input.NewFromBytes(b[offset:])
		if err != nil {
			return nil, err
		}
		offset += size

		t.Inputs = append(t.Inputs, i)
	}

	// create outputs
	outputCount, size := utils.DecodeVarInt(b[offset:])
	offset += size
	for i = 0; i < outputCount; i++ {
		o, size, err := output.NewFromBytes(b[offset:])
		if err != nil {
			return nil, err
		}
		offset += size
		t.Outputs = append(t.Outputs, o)
	}

	nLT := b[offset:]

	if len(nLT) != 4 {
		return nil, fmt.Errorf("nLockTime length must be 4 bytes long")
	}

	t.Locktime = binary.LittleEndian.Uint32(b[offset:])
	offset += 4

	return &t, nil
}

// AddInput adds a new input to the transaction.
func (bt *Tx) AddInput(input *input.Input) {
	bt.Inputs = append(bt.Inputs, input)
}

// From adds a new input to the transaction from the specified UTXO fields.
func (bt *Tx) From(txID string, vout uint32, prevTxLockingScript string, satoshis uint64) error {
	pts, err := script.NewFromHexString(prevTxLockingScript)
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
func (bt *Tx) InputCount() int {
	return len(bt.Inputs)
}

// OutputCount returns the number of transaction inputs.
func (bt *Tx) OutputCount() int {
	return len(bt.Outputs)
}

// AddOutput adds a new output to the transaction.
func (bt *Tx) AddOutput(output *output.Output) {

	bt.Outputs = append(bt.Outputs, output)
}

// PayTo creates a new P2PKH output from a BitCoin address (base58)
// and the satoshis amount and adds thats to the transaction.
func (bt *Tx) PayTo(addr string, satoshis uint64) error {
	o, err := output.NewP2PKHFromAddress(addr, satoshis)
	if err != nil {
		return err
	}

	bt.AddOutput(o)
	return nil
}

// ChangeToAddress calculates the amount of fees needed to cover the transaction
// and adds the left over change in a new P2PKH output using the address provided.
func (bt *Tx) ChangeToAddress(addr string, f []*mapi.Fee) error {
	s, err := script.NewP2PKHFromAddress(addr)
	if err != nil {
		return err
	}

	return bt.Change(s, f)
}

// Change calculates the amount of fees needed to cover the transaction
//  and adds the left over change in a new output using the script provided.
func (bt *Tx) Change(s *script.Script, f []*mapi.Fee) error {

	inputAmount := bt.GetTotalInputSatoshis()
	outputAmount := bt.GetTotalOutputSatoshis()

	if inputAmount < outputAmount {
		return errors.New("satoshis inputted to the tx are less than the outputted satoshis")
	}

	available := inputAmount - outputAmount

	stdFees, err := fees.GetStandardFee(f)
	if err != nil {
		return err
	}

	if !bt.canAddChange(available, stdFees) {
		return nil
	}

	o := output.Output{
		Satoshis:      0,
		LockingScript: s,
	}
	bt.AddOutput(&o)

	presignedFeeRequired, err := bt.getPresignedFeeRequired(f)
	if err != nil {
		return err
	}

	expectedUnlockingScriptFees, err := bt.getExpectedUnlockingScriptFees(f)
	if err != nil {
		return err
	}

	available -= (presignedFeeRequired + expectedUnlockingScriptFees)

	// add rest of available sats to the change output
	bt.Outputs[len(bt.GetOutputs())-1].Satoshis = available

	return nil
}

func (bt *Tx) canAddChange(available uint64, stdFees *mapi.Fee) bool {

	outputLen := bt.OutputCount()
	viuli := utils.VarIntUpperLimitInc(uint64(outputLen))

	if viuli == -1 {
		return false // upper limit of outputs in one tx reached
	}

	changeOutputFee := uint64(viuli)

	changeP2pkhByteLen := 8 + 25 // 8 bytes for satoshi value + 25 bytes for p2pkh script (e.g. 76a914cc...05388ac)
	changeOutputFee += uint64(changeP2pkhByteLen * stdFees.MiningFee.Satoshis / stdFees.MiningFee.Bytes)

	if available < changeOutputFee {
		return false // not enough change to add a whole change output so don't add anything and return
	}

	return true
}

func (bt *Tx) getPresignedFeeRequired(f []*mapi.Fee) (feeRequired uint64, err error) {

	stdBytes, dataBytes := bt.getStandardAndDataBytes()

	stdFee, err := fees.GetStandardFee(f)
	if err != nil {
		return 0, err
	}

	fr := stdBytes * stdFee.MiningFee.Satoshis / stdFee.MiningFee.Bytes

	dataFee, err := fees.GetDataFee(f)
	if err != nil {
		return 0, err
	}

	fr += dataBytes * dataFee.MiningFee.Satoshis / dataFee.MiningFee.Bytes

	return uint64(fr), nil

}

func (bt *Tx) getExpectedUnlockingScriptFees(f []*mapi.Fee) (feeRequired uint64, err error) {

	stdFee, err := fees.GetStandardFee(f)
	if err != nil {
		return 0, err
	}

	var expectedBytes int

	for _, in := range bt.GetInputs() {
		if !in.PreviousTxScript.IsP2PKH() {
			return 0, errors.New("non-P2PKH input used in the tx - unsupported")
		}
		expectedBytes += 109 // = 1 oppushdata + 70-73 sig + 1 sighash + 1 oppushdata + 33 public key
	}

	fr := expectedBytes * stdFee.MiningFee.Satoshis / stdFee.MiningFee.Bytes

	return uint64(fr), nil
}

func (bt *Tx) getStandardAndDataBytes() (stdBytes int, dataBytes int) {
	// Subtract the value of each output as well as keeping track of data outputs
	for _, out := range bt.GetOutputs() {
		if out.LockingScript.IsData() && len(*out.LockingScript) > 0 {
			dataBytes += len(*out.LockingScript)
		}
	}

	stdBytes = len(bt.ToBytes()) - dataBytes

	return
}

// HasDataOutputs returns true if the transaction has
// at least one data (OP_RETURN) output in it.
func (bt *Tx) HasDataOutputs() (hasDataOutputs bool) {
	for _, out := range bt.GetOutputs() {
		if out.LockingScript.IsData() {
			return true
		}
	}

	return false
}

// IsCoinbase determines if this transaction is a coinbase by
// checking if the tx input is a standard coinbase input.
func (bt *Tx) IsCoinbase() bool {
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
func (bt *Tx) GetInputs() []*input.Input {
	return bt.Inputs
}

// GetTotalInputSatoshis returns the total Satoshis inputted to the transaction.
func (bt *Tx) GetTotalInputSatoshis() uint64 {
	var total uint64
	for _, in := range bt.GetInputs() {
		total += in.PreviousTxSatoshis
	}

	return total
}

// GetOutputs returns an array of all outputs in the transaction.
func (bt *Tx) GetOutputs() []*output.Output {
	return bt.Outputs
}

// GetTotalOutputSatoshis returns the total Satoshis outputted from the transaction.
func (bt *Tx) GetTotalOutputSatoshis() uint64 {
	var total uint64
	for _, o := range bt.GetOutputs() {
		total += o.Satoshis
	}

	return total
}

// GetTxID returns the transaction ID of the transaction
// (which is also the transaction hash).
func (bt *Tx) GetTxID() string {
	return hex.EncodeToString(utils.ReverseBytes(crypto.Sha256d(bt.ToBytes())))
}

// ToString encodes the transaction into a hex string.
func (bt *Tx) ToString() string {
	return hex.EncodeToString(bt.ToBytes())
}

// ToBytes encodes the transaction into a byte array.
// See https://chainquery.com/bitcoin-cli/decoderawtransaction
func (bt *Tx) ToBytes() []byte {
	return bt.toBytesHelper(0, nil)
}

// ToBytesWithClearedInputs encodes the transaction into a byte array but clears its inputs first.
// This is used when signing transactions.
func (bt *Tx) ToBytesWithClearedInputs(index int, lockingScript []byte) []byte {
	return bt.toBytesHelper(index, lockingScript)
}

func (bt *Tx) toBytesHelper(index int, lockingScript []byte) []byte {
	h := make([]byte, 0)

	h = append(h, utils.GetLittleEndianBytes(bt.Version, 4)...)

	h = append(h, utils.VarInt(uint64(len(bt.GetInputs())))...)

	for i, in := range bt.GetInputs() {
		s := in.ToBytes(lockingScript != nil)
		if i == index && lockingScript != nil {
			h = append(h, utils.VarInt(uint64(len(lockingScript)))...)
			h = append(h, lockingScript...)
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
func (bt *Tx) Sign(index uint32, s Signer) error {
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
func (bt *Tx) SignAuto(s Signer) error {
	signedTx, err := s.SignAuto(bt)
	if err != nil {
		return err
	}
	*bt = *signedTx
	return nil
}

// ApplyUnlockingScript applies a script to the transaction at a specific index in
// unlocking script field.
func (bt *Tx) ApplyUnlockingScript(index uint32, s *script.Script) error {
	if bt.Inputs[index] != nil {
		bt.Inputs[index].UnlockingScript = s
		return nil
	}

	return fmt.Errorf("no input at index %d", index)
}
