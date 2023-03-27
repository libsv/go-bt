package bt

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"

	"github.com/libsv/go-bt/v2/bscript"
	"github.com/libsv/go-bt/v2/sighash"
	"github.com/pkg/errors"
)

// MakeBidArgs contains the arguments
// needed to make a bid to buy an
// ordinal.
type MakeBidArgs struct {
	BidAmount                 uint64
	OrdinalTxID               string
	OrdinalVOut               uint32
	BidderUTXOs               []*UTXO
	BuyerReceiveOrdinalScript *bscript.Script
	DummyOutputScript         *bscript.Script
	ChangeScript              *bscript.Script
	FQ                        *FeeQuote
}

// MakeBidToBuy1SatOrdinal makes a bid offer to buy a 1 sat ordinal
// at a specific price - this tx will be partially signed and will
// need to be completed by the seller if they accept the bid. Multiple
// people can make different bids and the seller will need to choose
// only one to go through and broadcast to the node network.
//
// Note: this function is meant for ordinals in 1 satoshi outputs instead
// of ordinal ranges in 1 output (>1 satoshi outputs).
func MakeBidToBuy1SatOrdinal(ctx context.Context, mba *MakeBidArgs) (*Tx, error) {
	if len(mba.BidderUTXOs) < 3 {
		return nil, ErrInsufficientUTXOs
	}

	tx := NewTx()

	// add dummy inputs
	err := tx.FromUTXOs(mba.BidderUTXOs[0], mba.BidderUTXOs[1])
	if err != nil {
		return nil, fmt.Errorf(`failed to add inputs: %w`, err)
	}

	OrdinalTxIDBytes, err := hex.DecodeString(mba.OrdinalTxID)
	if err != nil {
		return nil, err
	}
	tx.addInput(&Input{
		previousTxID:       OrdinalTxIDBytes,
		PreviousTxOutIndex: mba.OrdinalVOut,
		PreviousTxScript: func() *bscript.Script {
			//nolint:lll // add dummy ordinal PreviousTxScript
			// so that the change function can estimate
			// UnlockingScript sizes
			s, _ := bscript.NewFromHexString("76a914c25e9a2b70ec83d7b4fbd0f36f00a86723a48e6b88ac0063036f72645118746578742f706c61696e3b636861727365743d7574662d38000d48656c6c6f2c20776f726c642168") // hello world (text/plain) test inscription
			return s
		}(),
	})

	// add payment input(s)
	err = tx.FromUTXOs(mba.BidderUTXOs[2:]...)
	if err != nil {
		return nil, fmt.Errorf(`failed to add inputs: %w`, err)
	}

	// add dummy output to passthrough dummy inputs
	tx.AddOutput(&Output{
		LockingScript: mba.DummyOutputScript,
		Satoshis:      mba.BidderUTXOs[0].Satoshis + mba.BidderUTXOs[1].Satoshis,
	})

	// add ordinal receive output
	tx.AddOutput(&Output{
		LockingScript: mba.BuyerReceiveOrdinalScript,
		Satoshis:      1,
	})

	tx.AddOutput(&Output{
		Satoshis: mba.BidAmount,
		LockingScript: func() *bscript.Script { // add dummy p2pkh script to calc fees accurately
			s, _ := bscript.NewP2PKHFromAddress("1FunnyJoke111111111111111112AVXh5")
			return s
		}(),
	})

	err = tx.Change(mba.ChangeScript, mba.FQ)
	if err != nil {
		return nil, err
	}

	for i, u := range mba.BidderUTXOs {
		// skip 3rd input (ordinals input)
		j := i
		if i >= 2 {
			j++
		}

		if tx.Inputs[j] == nil {
			return nil, fmt.Errorf("input expected at index %d doesn't exist", j)
		}
		if !(bytes.Equal(u.TxID, tx.Inputs[j].previousTxID)) {
			return nil, ErrUTXOInputMismatch
		}
		if *u.Unlocker == nil {
			return nil, fmt.Errorf("UTXO unlocker at index %d not found", i)
		}
		err = tx.FillInput(ctx, *u.Unlocker, UnlockerParams{
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
//
// Note: index 2 should be the listed ordinal input.
type ValidateBidArgs struct {
	PreviousUTXOs []*UTXO // index 2 should be the listed ordinal input
	BidAmount     uint64
	ExpectedFQ    *FeeQuote
}

// Validate a bid to buy an ordinal
// given specific validation parameters.
func (vba *ValidateBidArgs) Validate(pstx *Tx) bool {
	if pstx.InputCount() < 4 {
		return false
	}
	if pstx.OutputCount() < 4 {
		return false
	}

	// check previous utxos match inputs
	if len(vba.PreviousUTXOs) != pstx.InputCount() {
		return false
	}
	for i := range vba.PreviousUTXOs {
		if !bytes.Equal(pstx.Inputs[i].PreviousTxID(), vba.PreviousUTXOs[i].TxID) {
			return false
		}
		if uint64(pstx.Inputs[i].PreviousTxOutIndex) != uint64(vba.PreviousUTXOs[i].Vout) {
			return false
		}
	}

	// check passthrough dummy inputs and output to avoid
	// mismatching and losing the ordinal to another output
	if (vba.PreviousUTXOs[0].Satoshis + vba.PreviousUTXOs[1].Satoshis) != pstx.Outputs[0].Satoshis {
		return false
	}

	// check lou (ListedOrdinalUTXO) matches supplied pstx input index 2
	pstxOrdinalInput := pstx.Inputs[2]
	if !bytes.Equal(pstxOrdinalInput.PreviousTxID(), vba.PreviousUTXOs[2].TxID) {
		return false
	}
	if uint64(pstxOrdinalInput.PreviousTxOutIndex) != uint64(vba.PreviousUTXOs[2].Vout) {
		return false
	}

	// check enough funds paid
	pstx.Outputs[2].Satoshis = vba.BidAmount
	enough, err := pstx.IsFeePaidEnough(vba.ExpectedFQ)
	if err != nil || !enough {
		return false
	}

	// TODO: check signatures valid

	return true
}

// AcceptBidArgs contains the arguments
// needed to make an offer to sell an
// ordinal.
type AcceptBidArgs struct {
	PSTx                       *Tx
	SellerReceiveOrdinalScript *bscript.Script
	OrdinalUnlocker            Unlocker
	ExtraUTXOs                 []*UTXO
}

// AcceptBidToBuy1SatOrdinal creates a PBST (Partially Signed Bitcoin
// Transaction) that offers a specific ordinal UTXO for sale at a
// specific price.
func AcceptBidToBuy1SatOrdinal(ctx context.Context, vba *ValidateBidArgs, aba *AcceptBidArgs) (*Tx, error) {
	if valid := vba.Validate(aba.PSTx); !valid {
		return nil, ErrInvalidSellOffer
	}

	if !aba.SellerReceiveOrdinalScript.IsP2PKH() {
		// TODO: if a script different to/bigger than p2pkh is used to
		// receive the ordinal, then the seller may need to add extra
		// utxos `aba.ExtraUTXOs` to cover the extra bytes since the
		// bidder only accounted for p2pkh script when calculating their
		// change.
		return nil, errors.New("only receive to p2pkh supported for now")
	}

	tx, err := NewTxFromBytes(aba.PSTx.Bytes())
	if err != nil {
		return nil, err
	}

	if tx.Outputs[2] == nil {
		return nil, ErrOrdinalOutputNoExist
	}
	tx.Outputs[2].LockingScript = aba.SellerReceiveOrdinalScript

	if tx.Inputs[2] == nil {
		return nil, ErrOrdinalInputNoExist
	}
	tx.Inputs[2].PreviousTxScript = vba.PreviousUTXOs[2].LockingScript
	tx.Inputs[2].PreviousTxSatoshis = vba.PreviousUTXOs[2].Satoshis
	err = tx.FillInput(ctx, aba.OrdinalUnlocker, UnlockerParams{InputIdx: 2})
	if err != nil {
		return nil, err
	}

	return tx, nil
}
