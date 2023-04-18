package ord

import (
	"bytes"
	"context"
	"fmt"

	"github.com/libsv/go-bt/v2"
	"github.com/libsv/go-bt/v2/bscript"
	"github.com/libsv/go-bt/v2/sighash"
)

// ListOrdinalArgs contains the arguments
// needed to make an offer to sell an
// ordinal.
type ListOrdinalArgs struct {
	SellerReceiveOutput *bt.Output
	OrdinalUTXO         *bt.UTXO
	OrdinalUnlocker     bt.Unlocker
}

// ListOrdinalForSale creates a PBST (Partially Signed Bitcoin
// Transaction) that offers a specific ordinal UTXO for sale at a
// specific price.
func ListOrdinalForSale(ctx context.Context, msoa *ListOrdinalArgs) (*bt.Tx, error) {
	tx := bt.NewTx()

	err := tx.FromUTXOs(msoa.OrdinalUTXO)
	if err != nil {
		return nil, err
	}

	tx.AddOutput(msoa.SellerReceiveOutput)

	err = tx.FillInput(ctx, msoa.OrdinalUnlocker, bt.UnlockerParams{
		InputIdx:     0,
		SigHashFlags: sighash.SingleForkID | sighash.AnyOneCanPay,
	})
	if err != nil {
		return nil, err
	}

	return tx, nil
}

// ValidateListingArgs are the arguments needed to
// validate a specific listing to sell an ordinal.
type ValidateListingArgs struct {
	ListedOrdinalUTXO *bt.UTXO
}

// Validate an ordinal sale offer listing
// given specific validation parameters.
func (vla *ValidateListingArgs) Validate(pstx *bt.Tx) bool {
	if pstx.InputCount() != 1 {
		return false
	}
	if pstx.OutputCount() != 1 {
		return false
	}

	// check lou (ListedOrdinalUTXO) matches supplied pstx input index 0
	pstxOrdinalInput := pstx.Inputs[0]
	if vla.ListedOrdinalUTXO == nil {
		return false
	}
	if !bytes.Equal(pstxOrdinalInput.PreviousTxID(), vla.ListedOrdinalUTXO.TxID) {
		return false
	}
	if uint64(pstxOrdinalInput.PreviousTxOutIndex) != uint64(vla.ListedOrdinalUTXO.Vout) {
		return false
	}

	// no need to check output value equals the listed value
	// since it's already signed by the input unlocking script

	// TODO: check signature valid

	return true
}

// AcceptListingArgs contains the arguments
// needed to make an offer to sell an
// ordinal.
type AcceptListingArgs struct {
	PSTx                      *bt.Tx
	UTXOs                     []*bt.UTXO
	BuyerReceiveOrdinalScript *bscript.Script
	DummyOutputScript         *bscript.Script
	ChangeScript              *bscript.Script
	FQ                        *bt.FeeQuote
}

// AcceptOrdinalSaleListing accepts a partially signed Bitcoin
// transaction offer to sell an ordinal. When accepting the offer,
// you will need to provide at least 2 UTXOs - with at least 1 being
// larger than the listed ordinal price.
func AcceptOrdinalSaleListing(ctx context.Context, vla *ValidateListingArgs, asoa *AcceptListingArgs) (*bt.Tx, error) {
	if valid := vla.Validate(asoa.PSTx); !valid {
		return nil, bt.ErrInvalidSellOffer
	}
	sellerOrdinalInput := asoa.PSTx.Inputs[0]
	sellerOutput := asoa.PSTx.Outputs[0]

	if len(asoa.UTXOs) < 2 {
		return nil, bt.ErrInsufficientUTXOs
	}

	if asoa.BuyerReceiveOrdinalScript == nil ||
		asoa.DummyOutputScript == nil ||
		asoa.ChangeScript == nil {
		return nil, bt.ErrEmptyScripts
	}

	// check at least 1 utxo is larger than the listed ordinal price
	validUTXOFound := false
	for i, u := range asoa.UTXOs {
		if u.Satoshis > sellerOutput.Satoshis {
			// Move the UTXO at index i to the beginning
			asoa.UTXOs = append([]*bt.UTXO{u}, append(asoa.UTXOs[:i], asoa.UTXOs[i+1:]...)...)
			validUTXOFound = true
			break
		}
	}
	if !validUTXOFound {
		return nil, bt.ErrInsufficientUTXOValue
	}

	tx := bt.NewTx()

	// add first input to pay for ordinal
	err := tx.FromUTXOs(asoa.UTXOs[0])
	if err != nil {
		return nil, fmt.Errorf(`failed to add input: %w`, err)
	}

	tx.Inputs = append(tx.Inputs, sellerOrdinalInput)

	// add input(s) to pay for tx fees
	err = tx.FromUTXOs(asoa.UTXOs[1:]...)
	if err != nil {
		return nil, fmt.Errorf(`failed to add inputs: %w`, err)
	}

	// add dummy output
	tx.AddOutput(&bt.Output{
		LockingScript: asoa.DummyOutputScript,
		Satoshis:      asoa.UTXOs[0].Satoshis - sellerOutput.Satoshis,
	})

	tx.AddOutput(sellerOutput)

	// add ordinal receive output
	tx.AddOutput(&bt.Output{
		LockingScript: asoa.BuyerReceiveOrdinalScript,
		Satoshis:      1,
	})

	err = tx.Change(asoa.ChangeScript, asoa.FQ)
	if err != nil {
		return nil, err
	}

	//nolint:dupl // TODO: are 2 dummies useful or to be removed?
	for i, u := range asoa.UTXOs {
		// skip 2nd input (ordinals input)
		j := i
		if i >= 1 {
			j++
		}

		if tx.Inputs[j] == nil {
			return nil, fmt.Errorf("input expected at index %d doesn't exist", j)
		}
		if !(bytes.Equal(u.TxID, tx.Inputs[j].PreviousTxID())) {
			return nil, bt.ErrUTXOInputMismatch
		}
		if *u.Unlocker == nil {
			return nil, fmt.Errorf("UTXO unlocker at index %d not found", i)
		}
		err = tx.FillInput(ctx, *u.Unlocker, bt.UnlockerParams{InputIdx: uint32(j)})
		if err != nil {
			return nil, err
		}
	}

	return tx, nil
}
