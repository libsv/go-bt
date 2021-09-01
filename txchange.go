package bt

import (
	"errors"

	"github.com/libsv/go-bt/v2/bscript"
)

// ChangeToAddress calculates the amount of fees needed to cover the transaction
// and adds the left over change in a new P2PKH output using the address provided.
func (tx *Tx) ChangeToAddress(addr string, f *FeeQuote) error {
	s, err := bscript.NewP2PKHFromAddress(addr)
	if err != nil {
		return err
	}

	return tx.Change(s, f)
}

// Change calculates the amount of fees needed to cover the transaction
//  and adds the left over change in a new output using the script provided.
func (tx *Tx) Change(s *bscript.Script, f *FeeQuote) error {
	_, _, err := tx.change(f, &changeOutput{
		lockingScript: s,
		newOutput:     true,
	})
	if err != nil {
		return err
	}
	return nil
}

// ChangeToExistingOutput will calculate fees and add them to an output at the index specified (0 based).
// If an invalid index is supplied and error is returned.
func (tx *Tx) ChangeToExistingOutput(index uint, f *FeeQuote) error {
	if int(index) > tx.OutputCount()-1 {
		return errors.New("index is greater than number of Inputs in transaction")
	}
	available, hasChange, err := tx.change(f, nil)
	if err != nil {
		return err
	}
	if hasChange {
		tx.Outputs[index].Satoshis += available
	}
	return nil
}

type changeOutput struct {
	lockingScript *bscript.Script
	newOutput     bool
}

// change will return the amount of satoshis to add to an input after fees are removed.
// True will be returned if change is required for this tx.
func (tx *Tx) change(f *FeeQuote, output *changeOutput) (uint64, bool, error) {
	inputAmount := tx.TotalInputSatoshis()
	outputAmount := tx.TotalOutputSatoshis()
	if inputAmount < outputAmount {
		return 0, false, errors.New("satoshis inputted to the tx are less than the outputted satoshis")
	}

	available := inputAmount - outputAmount

	standardFees, err := f.Fee(FeeTypeStandard)
	if err != nil {
		return 0, false, errors.New("standard fees not found")
	}

	var txFees *TxFees
	if txFees, err = tx.calculateTxFees(f); err != nil {
		return 0, false, err
	}
	changeFee, canAdd := tx.canAddChange(txFees, standardFees)
	if !canAdd {
		return 0, false, err
	}
	available -= txFees.Total + changeFee
	// if we want to add to a new output, set
	// newOutput to true, this will add the calculated change
	// into a new output.
	if output != nil && output.newOutput {
		tx.AddOutput(&Output{Satoshis: available, LockingScript: output.lockingScript})
	}

	return available, true, nil
}

// canAddChange will return true / false if the tx can have a change output
// added.
// Reasons this could be false are:
// - hitting max output limit
// - change would be below dust limit
// - not enough funds for change
// We also return the change output fee amount, if we can add change
func (tx *Tx) canAddChange(txFees *TxFees, standardFees *Fee) (uint64, bool) {
	varIntUpper := VarIntUpperLimitInc(uint64(tx.OutputCount()))
	if varIntUpper == -1 {
		return 0, false // upper limit of Outputs in one tx reached
	}
	changeOutputFee := uint64(varIntUpper)
	// 8 bytes for satoshi value +1 for varint length + 25 bytes for p2pkh script (e.g. 76a914cc...05388ac)
	changeP2pkhByteLen := 8 + 1 + 25
	changeOutputFee += uint64(changeP2pkhByteLen * standardFees.MiningFee.Satoshis / standardFees.MiningFee.Bytes)

	inputAmount := tx.TotalInputSatoshis()
	outputAmount := tx.TotalOutputSatoshis()
	available := inputAmount - outputAmount - txFees.Total

	// not enough change to add a whole change output so don't add anything and return
	if available >= changeOutputFee && available > 136 {
		return changeOutputFee, true
	}
	return 0, false
}
