package bt

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/libsv/go-bk/crypto"

	"github.com/libsv/go-bt/v2/bscript"
)

// ErrNoFund signals the FundGetterFunc has reached the end of its funds.
var ErrNoFund = errors.New("no remainings funds")

// FundGetterFunc is used for FromFunds. It expects *bt.Fund to be returned containing
// relevant input information, and an err informing any retrieval errors.
//
// It is expected that bt.ErrNoFund will be returned once the fund source is depleted.
type FundGetterFunc func(ctx context.Context) (*Fund, error)

// Fund contains information relating to the current fund. Its fields are intended
// for use with tx.From(...).
type Fund struct {
	TxID          string
	OutIndex      uint32
	LockingScript string
	Satoshis      uint64
}

// NewInputFromBytes returns a transaction input from the bytes provided.
func NewInputFromBytes(bytes []byte) (*Input, int, error) {
	if len(bytes) < 36 {
		return nil, 0, fmt.Errorf("input length too short < 36")
	}

	offset := 36
	l, size := DecodeVarInt(bytes[offset:])
	offset += size

	totalLength := offset + int(l) + 4 // 4 bytes for nSeq

	if len(bytes) < totalLength {
		return nil, 0, fmt.Errorf("input length too short < 36 + script + 4")
	}

	return &Input{
		previousTxID:       ReverseBytes(bytes[0:32]),
		PreviousTxOutIndex: binary.LittleEndian.Uint32(bytes[32:36]),
		SequenceNumber:     binary.LittleEndian.Uint32(bytes[offset+int(l):]),
		UnlockingScript:    bscript.NewFromBytes(bytes[offset : offset+int(l)]),
	}, totalLength, nil
}

// TotalInputSatoshis returns the total Satoshis inputted to the transaction.
func (tx *Tx) TotalInputSatoshis() (total uint64) {
	for _, in := range tx.Inputs {
		total += in.PreviousTxSatoshis
	}
	return
}

func (tx *Tx) addInput(input *Input) {
	tx.Inputs = append(tx.Inputs, input)
}

// AddP2PKHInputsFromTx will add all Outputs of given previous transaction
// that match a specific public key to your transaction.
func (tx *Tx) AddP2PKHInputsFromTx(pvsTx *Tx, matchPK []byte) error {
	for i, utxo := range pvsTx.Outputs {
		utxoPkHASH160, err := utxo.LockingScript.PublicKeyHash()
		if err != nil {
			return err
		}

		if bytes.Equal(utxoPkHASH160, crypto.Hash160(matchPK)) {
			if err := tx.From(pvsTx.TxID(), uint32(i), utxo.LockingScriptHexString(), utxo.Satoshis); err != nil {
				return err
			}
		}
	}

	return nil
}

// From adds a new input to the transaction from the specified UTXO fields, using the default
// finalised sequence number (0xFFFFFFFF). If you want a different nSeq, change it manually
// afterwards.
func (tx *Tx) From(prevTxID string, vout uint32, prevTxLockingScript string, satoshis uint64) error {
	pts, err := bscript.NewFromHexString(prevTxLockingScript)
	if err != nil {
		return err
	}

	i := &Input{
		PreviousTxOutIndex: vout,
		PreviousTxSatoshis: satoshis,
		PreviousTxScript:   pts,
		SequenceNumber:     DefaultSequenceNumber, // use default finalised sequence number
	}
	if err := i.PreviousTxIDAddStr(prevTxID); err != nil {
		return err
	}
	tx.addInput(i)

	return nil
}

// FromFunds continuously calls the provided bt.FundGetterFunc, adding each returned iteration
// as an input via tx.From(...), until it is estimated that inputs cover the outputs + fees.
//
// After completion, the receiver is ready for `Change(...)` to be called, and then be signed.
// Note, this function works under the assumption that receiver *bt.Tx alread has all the outputs
// which need covered.
//
// Example usage, for when working with a list:
//    tx.FromFunds(ctx, bt.NewFeeQuote(), func() bt.FundGetterFunc {
//        idx := 0
//        return func(ctx context.Context) (*bt.Fund, error) {
//            if idx >= len(funds) {
//                return nil, bt.ErrNoFund
//            }
//            defer func() { idx++ }()
//            return &bt.Fund{
//                TxID: funds[idx].TxID,
//                LockingScript: funds[idx].Script,
//                OutIndex: funds[idx].OutIndex,
//                Satoshis: funds[idx].Satoshis,
//            }, true
//        }
//    }())
func (tx *Tx) FromFunds(ctx context.Context, fq *FeeQuote, next FundGetterFunc) (err error) {
	var feesPaid bool
	for !feesPaid {
		fund, err := next(ctx)
		if err != nil {
			if err == ErrNoFund {
				break
			}

			return err
		}

		if err = tx.From(fund.TxID, fund.OutIndex,
			fund.LockingScript, fund.Satoshis); err != nil {
			return err
		}

		feesPaid, err = tx.EstimateIsFeePaidEnough(fq)
		if err != nil {
			return err
		}
	}
	if !feesPaid {
		return errors.New("insufficient funds from iterator")
	}

	return nil
}

// InputCount returns the number of transaction Inputs.
func (tx *Tx) InputCount() int {
	return len(tx.Inputs)
}

// PreviousOutHash returns a byte slice of inputs outpoints, for creating a signature hash
func (tx *Tx) PreviousOutHash() []byte {
	buf := make([]byte, 0)

	for _, in := range tx.Inputs {
		buf = append(buf, ReverseBytes(in.PreviousTxID())...)
		oi := make([]byte, 4)
		binary.LittleEndian.PutUint32(oi, in.PreviousTxOutIndex)
		buf = append(buf, oi...)
	}

	return crypto.Sha256d(buf)
}

// SequenceHash returns a byte slice of inputs SequenceNumber, for creating a signature hash
func (tx *Tx) SequenceHash() []byte {
	buf := make([]byte, 0)

	for _, in := range tx.Inputs {
		oi := make([]byte, 4)
		binary.LittleEndian.PutUint32(oi, in.SequenceNumber)
		buf = append(buf, oi...)
	}

	return crypto.Sha256d(buf)
}
