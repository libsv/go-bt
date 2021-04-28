package bt

import (
	"errors"

	"github.com/libsv/go-bt/bscript"
)

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
// and adds the left over change in a new output using the script provided.
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
