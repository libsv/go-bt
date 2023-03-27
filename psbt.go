package bt

import (
	"bytes"
	"context"
	"fmt"

	"github.com/libsv/go-bt/v2/bscript"
	"github.com/libsv/go-bt/v2/sighash"
	"github.com/pkg/errors"
)

// ListOrdinalArgs contains the arguments
// needed to make an offer to sell an
// ordinal (PSBT).
type ListOrdinalArgs struct {
	SellerReceiveOutput *Output
	OrdinalUTXO         *UTXO
	OrdinalUnlocker     Unlocker
}

// ListOrdinalForSale creates a PBST (Partially Signed Bitcoin
// Transaction) that offers a specific ordinal UTXO for sale at a
// specific price.
func ListOrdinalForSale(ctx context.Context, msoa *ListOrdinalArgs) (*Tx, error) {
	tx := NewTx()

	err := tx.FromUTXOs(msoa.OrdinalUTXO)
	if err != nil {
		return nil, err
	}

	tx.AddOutput(msoa.SellerReceiveOutput)

	err = tx.FillInput(ctx, msoa.OrdinalUnlocker, UnlockerParams{
		InputIdx:     0,
		SigHashFlags: sighash.SingleForkID | sighash.AnyOneCanPay,
	})
	if err != nil {
		return nil, err
	}

	return tx, nil
}

// AcceptListingArgs contains the arguments
// needed to make an offer to sell an
// ordinal (PSBT).
type AcceptListingArgs struct {
	PSTx                      *Tx
	Utxos                     []*UTXO
	BuyerReceiveOrdinalScript *bscript.Script
	DummyOutputScript         *bscript.Script
	ChangeScript              *bscript.Script
	FQ                        *FeeQuote
}

// AcceptOrdinalSaleListing accepts a partially signed Bitcoin
// transaction offer to sell an ordinal. When accepting the offer,
// you will need to provide at least 3 UTXOs - with the first 2
// being dummy utxos that will just pass through, and the rest with
// the required payment and tx fees.
func AcceptOrdinalSaleListing(ctx context.Context, asoa *AcceptListingArgs) (*Tx, error) {

	// TODO: ValidateSellOffer()
	// check if input 1 sat

	if len(asoa.Utxos) < 3 {
		return nil, ErrInsufficientUTXOs
	}

	tx := NewTx()

	// add dummy inputs
	err := tx.FromUTXOs(asoa.Utxos[0], asoa.Utxos[1])
	if err != nil {
		return nil, fmt.Errorf(`failed to add inputs: %w`, err)
	}

	sellerOrdinalInput := asoa.PSTx.Inputs[0] // TODO: get from psbt
	tx.addInput(sellerOrdinalInput)

	// add payment input(s)
	err = tx.FromUTXOs(asoa.Utxos[2:]...)
	if err != nil {
		return nil, fmt.Errorf(`failed to add inputs: %w`, err)
	}

	// add dummy output to passthrough dummy inputs
	tx.AddOutput(&Output{
		LockingScript: asoa.DummyOutputScript,
		Satoshis:      asoa.Utxos[0].Satoshis + asoa.Utxos[1].Satoshis,
	})

	// add ordinal receive output
	tx.AddOutput(&Output{
		LockingScript: asoa.BuyerReceiveOrdinalScript,
		Satoshis:      1,
	})

	sellerOutput := asoa.PSTx.Outputs[0] // TODO: get from psbt
	tx.AddOutput(sellerOutput)

	err = tx.Change(asoa.ChangeScript, asoa.FQ)
	if err != nil {
		return nil, err
	}

	for i, u := range asoa.Utxos {
		// skip 3rd input (ordinals input)
		j := i
		if i >= 2 {
			j++
		}

		if !(bytes.Equal(u.TxID, tx.Inputs[j].previousTxID)) {
			return nil, errors.New("input and utxo mismatch") // TODO: move to errors.go
		}
		err = tx.FillInput(context.Background(), *u.Unlocker, UnlockerParams{InputIdx: uint32(j)})
		if err != nil {
			return nil, err
		}
	}

	return tx, nil
}
