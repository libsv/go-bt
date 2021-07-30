package bt

import (
	"errors"

	"github.com/libsv/go-bt/v2/bscript"
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
//  and adds the left over change in a new output using the script provided.
func (tx *Tx) Change(s *bscript.Script, f []*Fee) error {
	available, hasChange, err := tx.change(s, f, true)
	if err != nil {
		return err
	}
	if hasChange {
		// add rest of available sats to the change output
		tx.Outputs[tx.OutputCount()-1].Satoshis = available
	}
	return nil
}

// ChangeToExistingOutput will calculate fees and add them to an output at the index specified (0 based).
// If an invalid index is supplied and error is returned.
func (tx *Tx) ChangeToExistingOutput(index uint, f []*Fee) error {
	if int(index) > tx.OutputCount()-1 {
		return errors.New("index is greater than number of Inputs in transaction")
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
	total := tx.TotalInputSatoshis() - tx.TotalOutputSatoshis()
	sats, _, err := tx.change(nil, f, false)
	if err != nil {
		return 0, err
	}
	return total - sats, nil
}

// change will return the amount of satoshis to add to an input after fees are removed.
// True will be returned if change has been added.
func (tx *Tx) change(s *bscript.Script, f []*Fee, newOutput bool) (uint64, bool, error) {
	inputAmount := tx.TotalInputSatoshis()
	outputAmount := tx.TotalOutputSatoshis()

	if inputAmount < outputAmount {
		return 0, false, errors.New("satoshis inputted to the tx are less than the outputted satoshis")
	}

	available := inputAmount - outputAmount

	standardFees, err := ExtractStandardFee(f)
	if err != nil {
		return 0, false, err
	}

	if !tx.canAddChange(available, standardFees) {
		return 0, false, err
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

	available -= preSignedFeeRequired + expectedUnlockingScriptFees

	return available, true, nil
}

func (tx *Tx) canAddChange(available uint64, standardFees *Fee) bool {

	varIntUpper := VarIntUpperLimitInc(uint64(tx.OutputCount()))
	if varIntUpper == -1 {
		return false // upper limit of Outputs in one tx reached
	}

	changeOutputFee := uint64(varIntUpper)

	// 8 bytes for satoshi value +1 for varint length + 25 bytes for p2pkh script (e.g. 76a914cc...05388ac)
	changeP2pkhByteLen := 8 + 1 + 25
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
	// Subtract the value of each output as well as keeping track of data Outputs
	for _, out := range tx.Outputs {
		if out.LockingScript.IsData() && len(*out.LockingScript) > 0 {
			dataBytes += len(*out.LockingScript)
		}
	}

	standardBytes = len(tx.Bytes()) - dataBytes
	return
}
