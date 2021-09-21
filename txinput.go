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

// ErrNoUTXO signals the UTXOGetterFunc has reached the end of its input.
var ErrNoUTXO = errors.New("no remaining utxos")

// UTXOGetterFunc is used for tx.FromUTXOs. It expects []*bt.UTXO to be returned containing
// utxos of which an input can be built.
//
// It is expected that bt.ErrNoUTXO will be returned once the utxo source is depleted.
type UTXOGetterFunc func(ctx context.Context, deficit uint64) ([]*UTXO, error)

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
	// Given that the prevTxID never changes, calculate it once up front.
	prevTxID := pvsTx.TxID()
	for i, utxo := range pvsTx.Outputs {
		utxoPkHASH160, err := utxo.LockingScript.PublicKeyHash()
		if err != nil {
			return err
		}

		if bytes.Equal(utxoPkHASH160, crypto.Hash160(matchPK)) {
			if err := tx.From(&UTXO{
				TxID:          prevTxID,
				Vout:          uint32(i),
				Satoshis:      utxo.Satoshis,
				LockingScript: utxo.LockingScriptHexString(),
			}); err != nil {
				return err
			}
		}
	}

	return nil
}

// From adds a new input to the transaction from the specified UTXO fields, using the default
// finalised sequence number (0xFFFFFFFF). If you want a different nSeq, change it manually
// afterwards.
func (tx *Tx) From(utxo *UTXO) error {
	pts, err := bscript.NewFromHexString(utxo.LockingScript)
	if err != nil {
		return err
	}

	i := &Input{
		PreviousTxOutIndex: utxo.Vout,
		PreviousTxSatoshis: utxo.Satoshis,
		PreviousTxScript:   pts,
		SequenceNumber:     DefaultSequenceNumber, // use default finalised sequence number
	}
	if err := i.PreviousTxIDAddStr(utxo.TxID); err != nil {
		return err
	}

	tx.addInput(i)
	return nil
}

// FromUTXOs continuously calls the provided bt.UTXOGetterFunc, adding each returned input
// as an input via tx.From(...), until it is estimated that inputs cover the outputs + fees.
//
// After completion, the receiver is ready for `Change(...)` to be called, and then be signed.
// Note, this function works under the assumption that receiver *bt.Tx alread has all the outputs
// which need covered.
//
// Example usage, for when working with a list:
//    tx.FromUTXOs(ctx, bt.NewFeeQuote(), func(ctx context.Context, deficit satoshis) ([]*bt.UTXO, error) {
//        utxos := make([]*bt.UTXO, 0)
//        for _, f := range funds {
//            deficit -= satoshis
//            utxos := append(utxos, &bt.UTXO{
//                TxID: f.TxID,
//                Vout: f.Vout,
//                LockingScript: f.Script,
//                Satoshis: f.Satoshis,
//            })
//            if deficit == 0 {
//                return utxos, nil
//            }
//        }
//        return nil, bt.ErrNoUTXO
//    })
func (tx *Tx) FromUTXOs(ctx context.Context, fq *FeeQuote, next UTXOGetterFunc) error {
	deficit, err := tx.estimateDeficit(fq)
	if err != nil {
		return err
	}
	for deficit != 0 {
		utxos, err := next(ctx, deficit)
		if err != nil {
			if errors.Is(err, ErrNoUTXO) {
				break
			}

			return err
		}

		for _, utxo := range utxos {
			if err = tx.From(utxo); err != nil {
				return err
			}
		}

		deficit, err = tx.estimateDeficit(fq)
		if err != nil {
			return err
		}
	}
	if deficit != 0 {
		return errors.New("insufficient utxos provided")
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
