package bt

import (
	"context"

	"github.com/libsv/go-bt/v2/sighash"
)

func MakeOfferToSellOrdinal(ctx context.Context, sellerOutput *Output,
	utxo *UTXO, unlocker Unlocker) (*Tx, error) {

	tx := NewTx()

	// add dummy input 0 to balance out tx
	// so that the next indices are equal
	// for the sighash single as well as
	// fo the ordinal theory rules to carry
	// over ordinals from input to output
	err := tx.From("0000000000000000000000000000000000000000000000000000000000000000", 0, "", 0)
	if err != nil {
		return nil, ErrDummyInput
	}

	err = tx.FromUTXOs(utxo)
	if err != nil {
		return nil, err
	}

	// add dummy output (to be replaced by buyer/taker)
	tx.AddDummyOutput()

	tx.AddOutput(sellerOutput)

	err = tx.FillInput(ctx, unlocker, UnlockerParams{
		InputIdx:     1,
		SigHashFlags: sighash.SingleForkID | sighash.AnyOneCanPay,
	})
	if err != nil {
		return nil, err
	}

	return tx, nil
}
