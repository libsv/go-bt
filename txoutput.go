package bt

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/libsv/go-bt/bscript"
	"github.com/libsv/go-bt/crypto"
)

// NewOutputFromBytes returns a transaction Output from the bytes provided
func NewOutputFromBytes(bytes []byte) (*Output, int, error) {
	if len(bytes) < 8 {
		return nil, 0, fmt.Errorf("output length too short < 8")
	}

	offset := 8
	l, size := DecodeVarInt(bytes[offset:])
	offset += size

	totalLength := offset + int(l)

	if len(bytes) < totalLength {
		return nil, 0, fmt.Errorf("output length too short < 8 + script")
	}

	s := bscript.Script(bytes[offset:totalLength])

	return &Output{
		Satoshis:      binary.LittleEndian.Uint64(bytes[0:8]),
		LockingScript: &s,
	}, totalLength, nil
}

// TotalOutputSatoshis returns the total Satoshis outputted from the transaction.
func (tx *Tx) TotalOutputSatoshis() (total uint64) {
	for _, o := range tx.Outputs {
		total += o.Satoshis
	}
	return
}

// AddP2PKHOutputFromPubKeyHashStr makes an output to a PKH with a value.
func (tx *Tx) AddP2PKHOutputFromPubKeyHashStr(publicKeyHash string, satoshis uint64) error {
	s, err := bscript.NewP2PKHFromPubKeyHashStr(publicKeyHash)
	if err != nil {
		return err
	}

	tx.AddOutput(&Output{
		Satoshis:      satoshis,
		LockingScript: s,
	})
	return nil
}

// AddP2PKHOutputFromPubKeyBytes makes an output to a PKH with a value.
func (tx *Tx) AddP2PKHOutputFromPubKeyBytes(publicKeyBytes []byte, satoshis uint64) error {
	s, err := bscript.NewP2PKHFromPubKeyBytes(publicKeyBytes)
	if err != nil {
		return err
	}

	tx.AddOutput(&Output{
		Satoshis:      satoshis,
		LockingScript: s,
	})
	return nil
}

// AddP2PKHOutputFromPubKeyStr makes an output to a PKH with a value.
func (tx *Tx) AddP2PKHOutputFromPubKeyStr(publicKey string, satoshis uint64) error {
	s, err := bscript.NewP2PKHFromPubKeyStr(publicKey)
	if err != nil {
		return err
	}

	tx.AddOutput(&Output{
		Satoshis:      satoshis,
		LockingScript: s,
	})
	return nil
}

// AddP2PKHOutputFromAddress makes an output to a PKH with a value.
func (tx *Tx) AddP2PKHOutputFromAddress(addr string, satoshis uint64) error {
	s, err := bscript.NewP2PKHFromAddress(addr)
	if err != nil {
		return err
	}

	tx.AddOutput(&Output{
		Satoshis:      satoshis,
		LockingScript: s,
	})
	return nil
}

// AddHashPuzzleOutput makes an output to a hash puzzle + PKH with a value.
func (tx *Tx) AddHashPuzzleOutput(secret, publicKeyHash string, satoshis uint64) error {

	publicKeyHashBytes, err := hex.DecodeString(publicKeyHash)
	if err != nil {
		return err
	}

	s := &bscript.Script{}

	s.AppendOpCode(bscript.OpHASH160)
	secretBytesHash := crypto.Hash160([]byte(secret))

	if err = s.AppendPushData(secretBytesHash); err != nil {
		return err
	}
	s.AppendOpCode(bscript.OpEQUALVERIFY)
	s.AppendOpCode(bscript.OpDUP)
	s.AppendOpCode(bscript.OpHASH160)

	if err = s.AppendPushData(publicKeyHashBytes); err != nil {
		return err
	}
	s.AppendOpCode(bscript.OpEQUALVERIFY)
	s.AppendOpCode(bscript.OpCHECKSIG)

	tx.AddOutput(&Output{
		Satoshis:      satoshis,
		LockingScript: s,
	})
	return nil
}

// AddOpReturnOutput creates a new Output with OP_FALSE OP_RETURN and then the data
// passed in encoded as hex.
func (tx *Tx) AddOpReturnOutput(data []byte) error {
	o, err := createOpReturnOutput([][]byte{data})
	if err != nil {
		return err
	}

	tx.AddOutput(o)
	return nil
}

// AddOpReturnPartsOutput creates a new Output with OP_FALSE OP_RETURN and then
// uses OP_PUSHDATA format to encode the multiple byte arrays passed in.
func (tx *Tx) AddOpReturnPartsOutput(data [][]byte) error {
	o, err := createOpReturnOutput(data)
	if err != nil {
		return err
	}

	tx.AddOutput(o)
	return nil
}

func createOpReturnOutput(data [][]byte) (*Output, error) {
	s := &bscript.Script{}

	s.AppendOpCode(bscript.OpFALSE)
	s.AppendOpCode(bscript.OpRETURN)
	err := s.AppendPushDataArray(data)
	if err != nil {
		return nil, err
	}

	return &Output{LockingScript: s}, nil
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
	return tx.AddP2PKHOutputFromAddress(addr, satoshis)
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

// Change calculates the amount of fees needed to cover the transaction
//  and adds the left over change in a new output using the script provided.
func (tx *Tx) Change(s *bscript.Script, f []*Fee) error {

	inputAmount := tx.TotalInputSatoshis()
	outputAmount := tx.TotalOutputSatoshis()

	if inputAmount < outputAmount {
		return errors.New("satoshis inputted to the tx are less than the outputted satoshis")
	}

	available := inputAmount - outputAmount

	standardFees, err := ExtractStandardFee(f)
	if err != nil {
		return err
	}

	if !tx.canAddChange(available, standardFees) {
		return nil
	}

	tx.AddOutput(&Output{Satoshis: 0, LockingScript: s})

	var preSignedFeeRequired uint64
	if preSignedFeeRequired, err = tx.getPreSignedFeeRequired(f); err != nil {
		return err
	}

	var expectedUnlockingScriptFees uint64
	if expectedUnlockingScriptFees, err = tx.getExpectedUnlockingScriptFees(f); err != nil {
		return err
	}

	available -= preSignedFeeRequired + expectedUnlockingScriptFees

	// add rest of available sats to the change output
	tx.Outputs[len(tx.Outputs)-1].Satoshis = available

	return nil
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

	standardFee, err := ExtractStandardFee(f)
	if err != nil {
		return 0, err
	}

	fr := standardBytes * standardFee.MiningFee.Satoshis / standardFee.MiningFee.Bytes

	var dataFee *Fee
	if dataFee, err = ExtractDataFee(f); err != nil {
		return 0, err
	}

	fr += dataBytes * dataFee.MiningFee.Satoshis / dataFee.MiningFee.Bytes

	return uint64(fr), nil
}

func (tx *Tx) getExpectedUnlockingScriptFees(f []*Fee) (uint64, error) {

	standardFee, err := ExtractStandardFee(f)
	if err != nil {
		return 0, err
	}

	var expectedBytes int

	for _, in := range tx.Inputs {
		if !in.PreviousTxScript.IsP2PKH() {
			return 0, errors.New("non-P2PKH input used in the tx - unsupported")
		}
		expectedBytes += 109 // = 1 oppushdata + 70-73 sig + 1 sighash + 1 oppushdata + 33 public key
	}

	return uint64(expectedBytes * standardFee.MiningFee.Satoshis / standardFee.MiningFee.Bytes), nil
}

func (tx *Tx) getStandardAndDataBytes() (standardBytes, dataBytes int) {
	// Subtract the value of each output as well as keeping track of data outputs
	for _, out := range tx.Outputs {
		if out.LockingScript.IsData() && len(*out.LockingScript) > 0 {
			dataBytes += len(*out.LockingScript)
		}
	}

	standardBytes = len(tx.ToBytes()) - dataBytes
	return
}
