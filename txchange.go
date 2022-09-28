package bt

import (
	"github.com/libsv/go-bt/v2/bscript"
)

const (
	// DustLimit is the current minimum txo output accepted by miners.
	DustLimit = 1
)

// ChangeToAddress calculates the amount of fees needed to cover the transaction
// and adds the leftover change in a new P2PKH output using the address provided.
func (tx *Tx) ChangeToAddress(addr string, f *FeeQuote) error {
	s, err := bscript.NewP2PKHFromAddress(addr)
	if err != nil {
		return err
	}

	return tx.Change(s, f)
}

// Change calculates the amount of fees needed to cover the transaction
//  and adds the leftover change in a new output using the script provided.
func (tx *Tx) Change(s *bscript.Script, f *FeeQuote) error {
	if _, _, err := tx.change(f, &changeOutput{
		lockingScript: s,
		newOutput:     true,
	}); err != nil {
		return err
	}
	return nil
}

// ChangeToExistingOutput will calculate fees and add them to an output at the index specified (0 based).
// If an invalid index is supplied and error is returned.
func (tx *Tx) ChangeToExistingOutput(index uint, f *FeeQuote) error {
	if int(index) > tx.OutputCount()-1 {
		return ErrOutputNoExist
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
		return 0, false, ErrInsufficientInputs
	}

	available := inputAmount - outputAmount
	size, err := tx.EstimateSizeWithTypes()
	if err != nil {
		return 0, false, err
	}
	stdFee, err := f.Fee(FeeTypeStandard)
	if err != nil {
		return 0, false, err
	}
	dataFee, err := f.Fee(FeeTypeData)
	if err != nil {
		return 0, false, err
	}
	varIntUpper := VarInt(tx.OutputCount()).UpperLimitInc()
	if varIntUpper == -1 {
		return 0, false, nil
	}
	changeOutputFee := varIntUpper
	changeP2pkhByteLen := uint64(0)
	if output != nil && output.newOutput {
		changeP2pkhByteLen = uint64(8 + 1 + 25)
	}

	sFees := (size.TotalStdBytes + changeP2pkhByteLen) * uint64(stdFee.MiningFee.Satoshis) / uint64(stdFee.MiningFee.Bytes)
	dFees := size.TotalDataBytes * uint64(dataFee.MiningFee.Satoshis) / uint64(dataFee.MiningFee.Bytes)
	txFees := sFees + dFees + uint64(changeOutputFee)

	// not enough to add change, no change to add
	if available <= txFees || available-txFees <= DustLimit {
		return 0, false, nil
	}

	// if we want to add to a new output, set
	// newOutput to true, this will add the calculated change
	// into a new output
	available -= txFees
	if output != nil && output.newOutput {
		tx.AddOutput(&Output{Satoshis: available, LockingScript: output.lockingScript})
	}
	return available, true, nil
}
