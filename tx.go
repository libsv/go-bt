package bt

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/libsv/go-bt/bscript"
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

const (
	// DustLimit is the current minimum output satoshis accepted by the network.
	DustLimit = 136
)

// Tx wraps a bitcoin transaction
//
// DO NOT CHANGE ORDER - Optimized memory via malign
//
type Tx struct {
	// TODO: make variables private?
	Inputs   []*Input
	Outputs  []*Output
	Version  uint32
	LockTime uint32
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

// AddInput adds a new input to the transaction.
func (tx *Tx) AddInput(input *Input) {

	// TODO: v2 make input (and other internal) elements private and not exposed
	// so that we only store previoustxid in bytes and then do the conversion
	// with getters and setters
	if input.PreviousTxIDBytes == nil {
		ptidb, err := hex.DecodeString(input.PreviousTxID)
		if err == nil {
			input.PreviousTxIDBytes = ptidb
		}
	}

	tx.Inputs = append(tx.Inputs, input)
}

// AddInputFromTx take all outputs from previous transaction
// that match a specific public key, add it as input to this new transaction.
func (tx *Tx) AddInputFromTx(pvsTx *Tx, matchPK []byte) error {
	matchPKHASH160 := crypto.Hash160(matchPK)
	for i, utxo := range pvsTx.Outputs {
		utxoPkHASH160, errPK := utxo.LockingScript.GetPublicKeyHash()
		if errPK != nil {
			return errPK
		}
		if !bytes.Equal(utxoPkHASH160, matchPKHASH160) {
			continue
		}
		tx.AddInput(&Input{
			PreviousTxIDBytes:  pvsTx.GetTxIDAsBytes(),
			PreviousTxID:       pvsTx.GetTxID(),
			PreviousTxOutIndex: uint32(i),
			PreviousTxSatoshis: utxo.Satoshis,
			PreviousTxScript:   utxo.LockingScript,
			SequenceNumber:     0xffffffff,
		})
	}
	return nil
}

// From adds a new input to the transaction from the specified UTXO fields.
func (tx *Tx) From(txID string, vout uint32, prevTxLockingScript string, satoshis uint64) error {
	pts, err := bscript.NewFromHexString(prevTxLockingScript)
	if err != nil {
		return err
	}

	ptid, err := hex.DecodeString(txID)
	if err != nil {
		return err
	}

	tx.AddInput(&Input{
		PreviousTxIDBytes:  ptid,
		PreviousTxID:       txID,
		PreviousTxOutIndex: vout,
		PreviousTxSatoshis: satoshis,
		PreviousTxScript:   pts,
		SequenceNumber:     DefaultSequenceNumber,
	})

	return nil
}

// InputCount returns the number of transaction inputs.
func (tx *Tx) InputCount() int {
	return len(tx.Inputs)
}

// OutputCount returns the number of transaction inputs.
func (tx *Tx) OutputCount() int {
	return len(tx.Outputs)
}

// AddOutput adds a new output to the transaction.
func (tx *Tx) AddOutput(output *Output) {
	tx.Outputs = append(tx.Outputs, output)
}

// PayTo creates a new P2PKH output from a BitCoin address (base58)
// and the satoshis amount and adds that to the transaction.
func (tx *Tx) PayTo(addr string, satoshis uint64) error {
	o, err := NewP2PKHOutputFromAddress(addr, satoshis)
	if err != nil {
		return err
	}

	tx.AddOutput(o)
	return nil
}

// ChangeToAddress calculates the amount of fees needed to cover the transaction
// and adds the left over change in a new P2PKH output using the address provided.
func (tx *Tx) ChangeToAddress(addr string, f []*Fee) error {
	s, err := bscript.NewP2PKHFromAddress(addr)
	if err != nil {
		return err
	}
	return tx.Change(s, f)
}

// HasOutputsWithAddress will return the index of any outputs found matching
// the address 'addr'.
//
// bool will be false if none have been found.
// err will not be nil if the addr is not a valid P2PKH address.
func (tx *Tx) HasOutputsWithAddress(addr string) ([]int, bool, error) {
	cs, err := bscript.NewP2PKHFromAddress(addr)
	if err != nil {
		return nil, false, err
	}
	ii, ok := tx.HasOutputsWithScript(cs)
	return ii, ok, nil
}

// HasOutputsWithScript will return the index of any outputs found matching
// the locking script 's'.
//
// bool will be false if none have been found.
func (tx *Tx) HasOutputsWithScript(s *bscript.Script) ([]int, bool) {
	idx := make([]int, 0)
	for i, o := range tx.Outputs {
		if bytes.Equal(*o.LockingScript, *s) {
			idx = append(idx, i)
		}
	}
	return idx, len(idx) > 0
}

// Change calculates the amount of fees needed to cover the transaction
// and adds the left over change in a new output using the script provided.
func (tx *Tx) Change(s *bscript.Script, f []*Fee) error {
	available, hasChange, err := tx.change(s, f, true)
	if err != nil {
		return err
	}
	if hasChange {
		// add rest of available sats to the change output
		tx.Outputs[len(tx.GetOutputs())-1].Satoshis = available
	}
	return nil
}

// ChangeToOutput will calculate fees and add them to an output at the index specified (0 based).
// If an invalid index is supplied and error is returned.
func (tx *Tx) ChangeToOutput(index uint, f []*Fee) error {
	if int(index) > len(tx.Outputs)-1 {
		return errors.New("index is greater than number of inputs in transaction")
	}
	available, hasChange, err := tx.change(tx.Outputs[index].LockingScript, f, false)
	if err != nil {
		return err
	}
	if hasChange {
		tx.Outputs[index].Satoshis += available
	}
	return nil
}

// CalculateFee will return the amount of fees the current transaction will
// require.
func (tx *Tx) CalculateFee(f []*Fee) (uint64, error) {
	total := tx.GetTotalInputSatoshis() - tx.GetTotalOutputSatoshis()
	sats, _, err := tx.change(nil, f, false)
	if err != nil {
		return 0, err
	}
	return total - sats, nil
}

// change will return the amount of satoshis to add to an output after fees are removed.
// True will be returned if a change output has been added.
func (tx *Tx) change(s *bscript.Script, f []*Fee, newOutput bool) (uint64, bool, error) {
	inputAmount := tx.GetTotalInputSatoshis()
	outputAmount := tx.GetTotalOutputSatoshis()

	if inputAmount < outputAmount {
		return 0, false, errors.New("satoshis inputted to the tx are less than the outputted satoshis")
	}

	available := inputAmount - outputAmount

	standardFees, err := GetStandardFee(f)
	if err != nil {
		return 0, false, err
	}

	if !tx.canAddChange(available, standardFees) {
		return 0, false, nil
	}
	if newOutput {
		tx.AddOutput(&Output{Satoshis: 0, LockingScript: s})
	}

	var preSignedFeeRequired uint64
	if preSignedFeeRequired, err = tx.getPreSignedFeeRequired(f); err != nil {
		return 0, false, err
	}

	var expectedUnlockingScriptFees uint64
	if expectedUnlockingScriptFees, err = tx.getExpectedUnlockingScriptFees(f); err != nil {
		return 0, false, err
	}

	if available < (preSignedFeeRequired + expectedUnlockingScriptFees) {
		if newOutput {
			tx.Outputs = tx.Outputs[:tx.OutputCount()-1]
		}
		return 0, false, nil
	}
	available -= preSignedFeeRequired + expectedUnlockingScriptFees
	if available <= DustLimit {
		if newOutput {
			tx.Outputs = tx.Outputs[:tx.OutputCount()-1]
		}
		return 0, false, nil
	}

	return available, true, nil
}

func (tx *Tx) canAddChange(available uint64, standardFees *Fee) bool {
	varIntUpper := VarIntUpperLimitInc(uint64(tx.OutputCount()))
	if varIntUpper == -1 {
		return false // upper limit of outputs in one tx reached
	}

	changeOutputFee := uint64(varIntUpper)

	changeP2pkhByteLen := 8 + 25 // 8 bytes for satoshi value + 25 bytes for p2pkh script (e.g. 76a914cc...05388ac)
	changeOutputFee += uint64(changeP2pkhByteLen * standardFees.MiningFee.Satoshis / standardFees.MiningFee.Bytes)

	// not enough change to add a whole change output so don't add anything and return
	return available >= changeOutputFee
}

func (tx *Tx) getPreSignedFeeRequired(f []*Fee) (uint64, error) {

	standardBytes, dataBytes := tx.getStandardAndDataBytes()

	standardFee, err := GetStandardFee(f)
	if err != nil {
		return 0, err
	}

	fr := standardBytes * standardFee.MiningFee.Satoshis / standardFee.MiningFee.Bytes

	var dataFee *Fee
	if dataFee, err = GetDataFee(f); err != nil {
		return 0, err
	}

	fr += dataBytes * dataFee.MiningFee.Satoshis / dataFee.MiningFee.Bytes
	return uint64(fr), nil
}

func (tx *Tx) getExpectedUnlockingScriptFees(f []*Fee) (uint64, error) {

	standardFee, err := GetStandardFee(f)
	if err != nil {
		return 0, err
	}

	var expectedBytes int

	for _, in := range tx.GetInputs() {
		if !in.PreviousTxScript.IsP2PKH() {
			return 0, errors.New("non-P2PKH input used in the tx - unsupported")
		}
		expectedBytes += 109 // = 1 oppushdata + 70-73 sig + 1 sighash + 1 oppushdata + 33 public key
	}

	return uint64(expectedBytes * standardFee.MiningFee.Satoshis / standardFee.MiningFee.Bytes), nil
}

func (tx *Tx) getStandardAndDataBytes() (standardBytes, dataBytes int) {
	// Subtract the value of each output as well as keeping track of data outputs
	for _, out := range tx.GetOutputs() {
		if out.LockingScript.IsData() && len(*out.LockingScript) > 0 {
			dataBytes += len(*out.LockingScript)
		}
	}

	standardBytes = len(tx.ToBytes()) - dataBytes
	return
}

// HasDataOutputs returns true if the transaction has
// at least one data (OP_RETURN) output in it.
func (tx *Tx) HasDataOutputs() bool {
	for _, out := range tx.GetOutputs() {
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

	// todo: make constant(s)?
	if tx.Inputs[0].PreviousTxID != "0000000000000000000000000000000000000000000000000000000000000000" {
		return false
	}

	if tx.Inputs[0].PreviousTxOutIndex == DefaultSequenceNumber || tx.Inputs[0].SequenceNumber == DefaultSequenceNumber {
		return true
	}

	return false
}

// GetInputs returns an array of all inputs in the transaction.
func (tx *Tx) GetInputs() []*Input {
	return tx.Inputs
}

// GetTotalInputSatoshis returns the total Satoshis inputted to the transaction.
func (tx *Tx) GetTotalInputSatoshis() (total uint64) {
	for _, in := range tx.GetInputs() {
		total += in.PreviousTxSatoshis
	}
	return
}

// GetOutputs returns an array of all outputs in the transaction.
func (tx *Tx) GetOutputs() []*Output {
	return tx.Outputs
}

// GetTotalOutputSatoshis returns the total Satoshis outputted from the transaction.
func (tx *Tx) GetTotalOutputSatoshis() (total uint64) {
	for _, o := range tx.GetOutputs() {
		total += o.Satoshis
	}
	return
}

// GetTxIDAsBytes returns the transaction ID of the transaction as bytes
// (which is also the transaction hash).
func (tx *Tx) GetTxIDAsBytes() []byte {
	return ReverseBytes(crypto.Sha256d(tx.ToBytes()))
}

// GetTxID returns the transaction ID of the transaction
// (which is also the transaction hash).
func (tx *Tx) GetTxID() string {
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

	h = append(h, GetLittleEndianBytes(tx.Version, 4)...)

	h = append(h, VarInt(uint64(len(tx.GetInputs())))...)

	for i, in := range tx.GetInputs() {
		s := in.ToBytes(lockingScript != nil)
		if i == index && lockingScript != nil {
			h = append(h, VarInt(uint64(len(lockingScript)))...)
			h = append(h, lockingScript...)
		} else {
			h = append(h, s...)
		}
	}

	h = append(h, VarInt(uint64(len(tx.GetOutputs())))...)
	for _, out := range tx.GetOutputs() {
		h = append(h, out.ToBytes()...)
	}

	lt := make([]byte, 4)
	binary.LittleEndian.PutUint32(lt, tx.LockTime)

	return append(h, lt...)
}

// Sign is used to sign the transaction at a specific input index.
// It takes a Signed interface as a parameter so that different
// signing implementations can be used to sign the transaction -
// for example internal/local or external signing.
func (tx *Tx) Sign(index uint32, s Signer) error {
	// TODO: v2 put tx serialization here so that the Signer.Sign
	// func only does signing and not also serialization which
	// should be done here.

	signedTx, err := s.Sign(index, tx)
	if err != nil {
		return err
	}
	*tx = *signedTx
	return nil
}

// SignAuto is used to automatically check which P2PKH inputs are
// able to be signed (match the public key) and then sign them.
// It takes a Signed interface as a parameter so that different
// signing implementations can be used to sign the transaction -
// for example internal/local or external signing.
func (tx *Tx) SignAuto(s Signer) (inputsSigned []int, err error) {
	var signedTx *Tx
	if signedTx, inputsSigned, err = s.SignAuto(tx); err != nil {
		return
	}
	*tx = *signedTx
	return
}

// ApplyUnlockingScript applies a script to the transaction at a specific index in
// unlocking script field.
func (tx *Tx) ApplyUnlockingScript(index uint32, s *bscript.Script) error {
	if tx.Inputs[index] != nil {
		tx.Inputs[index].UnlockingScript = s
		return nil
	}

	return fmt.Errorf("no input at index %d", index)
}
