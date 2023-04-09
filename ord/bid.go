package ord

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"

	"github.com/libsv/go-bt/v2"
	"github.com/libsv/go-bt/v2/bscript"
	"github.com/libsv/go-bt/v2/sighash"
)

// MakeBidArgs contains the arguments
// needed to make a bid to buy an
// ordinal.
type MakeBidArgs struct {
	BidAmount                 uint64
	OrdinalTxID               string
	OrdinalVOut               uint32
	BidderUTXOs               []*bt.UTXO
	BuyerReceiveOrdinalScript *bscript.Script
	DummyOutputScript         *bscript.Script
	ChangeScript              *bscript.Script
	FQ                        *bt.FeeQuote
}

// MakeBidToBuy1SatOrdinal makes a bid offer to buy a 1 sat ordinal
// at a specific price - this tx will be partially signed and will
// need to be completed by the seller if they accept the bid. Multiple
// people can make different bids and the seller will need to choose
// only one to go through and broadcast to the node network.
//
// Note: this function is meant for ordinals in 1 satoshi outputs instead
// of ordinal ranges in 1 output (>1 satoshi outputs).
func MakeBidToBuy1SatOrdinal(ctx context.Context, mba *MakeBidArgs) (*bt.Tx, error) {
	if len(mba.BidderUTXOs) < 2 {
		return nil, bt.ErrInsufficientUTXOs
	}

	// check at least 1 utxo is larger than the listed ordinal price
	validUTXOFound := false
	for i, u := range mba.BidderUTXOs {
		if u.Satoshis > mba.BidAmount {
			// Move the UTXO at index i to the beginning
			mba.BidderUTXOs = append([]*bt.UTXO{u}, append(mba.BidderUTXOs[:i], mba.BidderUTXOs[i+1:]...)...)
			validUTXOFound = true
			break
		}
	}
	if !validUTXOFound {
		return nil, bt.ErrInsufficientUTXOValue
	}

	tx := bt.NewTx()

	// add dummy inputs
	err := tx.FromUTXOs(mba.BidderUTXOs[0])
	if err != nil {
		return nil, fmt.Errorf(`failed to add inputs: %w`, err)
	}

	OrdinalTxIDBytes, err := hex.DecodeString(mba.OrdinalTxID)
	if err != nil {
		return nil, err
	}
	emptyOrdInput := &bt.Input{
		PreviousTxOutIndex: mba.OrdinalVOut,
		PreviousTxScript: func() *bscript.Script {
			//nolint:lll // add dummy ordinal PreviousTxScript
			// so that the change function can estimate
			// UnlockingScript sizes
			s, _ := bscript.NewFromHexString("76a914c25e9a2b70ec83d7b4fbd0f36f00a86723a48e6b88ac0063036f72645118746578742f706c61696e3b636861727365743d7574662d38000d48656c6c6f2c20776f726c642168") // hello world (text/plain) test inscription
			return s
		}(),
	}
	err = emptyOrdInput.PreviousTxIDAdd(OrdinalTxIDBytes)
	if err != nil {
		return nil, fmt.Errorf(`failed to add ordinal input: %w`, err)
	}
	tx.Inputs = append(tx.Inputs, emptyOrdInput)

	// add payment input(s)
	err = tx.FromUTXOs(mba.BidderUTXOs[1:]...)
	if err != nil {
		return nil, fmt.Errorf(`failed to add inputs: %w`, err)
	}

	// add dummy output
	tx.AddOutput(&bt.Output{
		LockingScript: mba.DummyOutputScript,
		Satoshis:      mba.BidderUTXOs[0].Satoshis - mba.BidAmount,
	})

	tx.AddOutput(&bt.Output{
		Satoshis: mba.BidAmount,
		LockingScript: func() *bscript.Script { // add dummy p2pkh script to calc fees accurately
			s, _ := bscript.NewP2PKHFromAddress("1FunnyJoke111111111111111112AVXh5")
			return s
		}(),
	})

	// add ordinal receive output
	tx.AddOutput(&bt.Output{
		LockingScript: mba.BuyerReceiveOrdinalScript,
		Satoshis:      1,
	})

	err = tx.Change(mba.ChangeScript, mba.FQ)
	if err != nil {
		return nil, err
	}

	//nolint: dupl // TODO: are 2 dummies useful or to be removed?
	for i, u := range mba.BidderUTXOs {
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
		err = tx.FillInput(ctx, *u.Unlocker, bt.UnlockerParams{
			InputIdx:     uint32(j),
			SigHashFlags: sighash.SingleForkID,
		})
		if err != nil {
			return nil, err
		}
	}

	return tx, nil
}

// ValidateBidArgs are the arguments needed to
// validate a specific bid to buy an ordinal.
// as they appear in the tx.
type ValidateBidArgs struct {
	OrdinalUTXO *bt.UTXO
	BidAmount   uint64
	ExpectedFQ  *bt.FeeQuote
}

// Validate a bid to buy an ordinal
// given specific validation parameters.
func (vba *ValidateBidArgs) Validate(pstx *bt.Tx) bool {
	if pstx.InputCount() < 3 {
		return false
	}
	if pstx.OutputCount() < 3 { // technically should have 4 including change
		return false
	}

	// check OrdinalUTXO matches supplied pstx input index 1
	pstxOrdinalInput := pstx.Inputs[1]
	if !bytes.Equal(pstxOrdinalInput.PreviousTxID(), vba.OrdinalUTXO.TxID) {
		return false
	}
	if uint64(pstxOrdinalInput.PreviousTxOutIndex) != uint64(vba.OrdinalUTXO.Vout) {
		return false
	}

	// set the value of the output for the bid amount
	pstx.Outputs[1].Satoshis = vba.BidAmount

	// check enough fees paid
	enough, err := pstx.IsFeePaidEnough(vba.ExpectedFQ)
	if err != nil || !enough {
		return false
	}

	// TODO: check signatures valid

	return true
}

// AcceptBidArgs contains the arguments
// needed to accept a bid to buy an
// ordinal.
type AcceptBidArgs struct {
	PSTx                *bt.Tx
	SellerReceiveScript *bscript.Script
	OrdinalUnlocker     bt.Unlocker
}

// AcceptBidToBuy1SatOrdinal accepts a partially signed Bitcoin
// transaction bid to buy an ordinal.
//
func AcceptBidToBuy1SatOrdinal(ctx context.Context, vba *ValidateBidArgs, aba *AcceptBidArgs) (*bt.Tx, error) {
	if valid := vba.Validate(aba.PSTx); !valid {
		return nil, bt.ErrInvalidSellOffer
	}

	tx := aba.PSTx.Clone()

	tx.Outputs[1].LockingScript = aba.SellerReceiveScript
	// check if fees paid are still enough with new
	// locking script
	enough, err := tx.IsFeePaidEnough(vba.ExpectedFQ)
	if err != nil || !enough {
		return nil, bt.ErrInsufficientFees
	}

	tx.Inputs[1].PreviousTxScript = vba.OrdinalUTXO.LockingScript
	tx.Inputs[1].PreviousTxSatoshis = vba.OrdinalUTXO.Satoshis
	err = tx.FillInput(ctx, aba.OrdinalUnlocker, bt.UnlockerParams{InputIdx: 1})
	if err != nil {
		return nil, err
	}

	return tx, nil
}
