package ord

import (
	"bytes"
	"context"
	"fmt"

	"github.com/libsv/go-bt/v2"
)

// AcceptOrdinalSaleListing2Dummies accepts a partially signed Bitcoin
// transaction offer to sell an ordinal. When accepting the offer,
// you will need to provide at least 3 UTXOs - with the first 2
// being dummy utxos that will just pass through, and the rest with
// the required payment and tx fees.
func AcceptOrdinalSaleListing2Dummies(ctx context.Context, vla *ValidateListingArgs,
	asoa *AcceptListingArgs) (*bt.Tx, error) {

	if valid := vla.Validate(asoa.PSTx); !valid {
		return nil, bt.ErrInvalidSellOffer
	}
	sellerOrdinalInput := asoa.PSTx.Inputs[0]
	sellerOutput := asoa.PSTx.Outputs[0]

	if len(asoa.UTXOs) < 3 {
		return nil, bt.ErrInsufficientUTXOs
	}

	tx := bt.NewTx()

	// add dummy inputs
	err := tx.FromUTXOs(asoa.UTXOs[0], asoa.UTXOs[1])
	if err != nil {
		return nil, fmt.Errorf(`failed to add inputs: %w`, err)
	}

	tx.Inputs = append(tx.Inputs, sellerOrdinalInput)

	// add payment input(s)
	err = tx.FromUTXOs(asoa.UTXOs[2:]...)
	if err != nil {
		return nil, fmt.Errorf(`failed to add inputs: %w`, err)
	}

	// add dummy output to passthrough dummy inputs
	tx.AddOutput(&bt.Output{
		LockingScript: asoa.DummyOutputScript,
		Satoshis:      asoa.UTXOs[0].Satoshis + asoa.UTXOs[1].Satoshis,
	})

	// add ordinal receive output
	tx.AddOutput(&bt.Output{
		LockingScript: asoa.BuyerReceiveOrdinalScript,
		Satoshis:      1,
	})

	tx.AddOutput(sellerOutput)

	err = tx.Change(asoa.ChangeScript, asoa.FQ)
	if err != nil {
		return nil, err
	}

	//nolint:dupl // false positive
	for i, u := range asoa.UTXOs {
		// skip 3rd input (ordinals input)
		j := i
		if i >= 2 {
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
