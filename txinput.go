package bt

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"

	"github.com/libsv/go-bk/crypto"

	"github.com/libsv/go-bt/v2/bscript"
)

// ErrNoUTXO signals the UTXOGetterFunc has reached the end of its input.
var ErrNoUTXO = errors.New("no remaining utxos")

// UTXOGetterFunc is used for tx.Fund(...). It provides the amount of satoshis required
// for funding as `deficit`, and expects []*bt.UTXO to be returned containing
// utxos of which *bt.Input's can be built.
// If the returned []*bt.UTXO does not cover the deficit after fee recalculation, then
// this UTXOGetterFunc is called again, with the newly calculated deficit passed in.
//
// It is expected that bt.ErrNoUTXO will be returned once the utxo source is depleted.
type UTXOGetterFunc func(ctx context.Context, deficit uint64) ([]*UTXO, error)

// NewInputFromBytes returns a transaction input from the bytes provided.
func NewInputFromBytes(b []byte) (*Input, int, error) {
	if len(b) < 36 {
		return nil, 0, fmt.Errorf("input length too short < 36")
	}

	offset := 36
	l, size := DecodeVarInt(b[offset:])
	offset += size

	totalLength := offset + int(l) + 4 // 4 bytes for nSeq

	if len(b) < totalLength {
		return nil, 0, fmt.Errorf("input length too short < 36 + script + 4")
	}

	r := bytes.NewReader(b)

	i, err := NewInputFromReader(r)
	if err != nil {
		return nil, 0, err
	}

	return i, len(i.Bytes(false)), nil
}

// NewInputFromReader returns a transaction input from the io.Reader provided.
func NewInputFromReader(r io.Reader) (*Input, error) {
	previousTxID := make([]byte, 32)
	if n, err := io.ReadFull(r, previousTxID); n != 32 || err != nil {
		return nil, fmt.Errorf("Could not read previousTxID(32), got %d bytes and err: %w", n, err)
	}

	prevIndex := make([]byte, 4)
	if n, err := io.ReadFull(r, prevIndex); n != 4 || err != nil {
		return nil, fmt.Errorf("Could not read prevIndex(4), got %d bytes and err: %w", n, err)
	}

	l, _, err := DecodeVarIntFromReader(r)
	if err != nil {
		return nil, fmt.Errorf("Could not read varint: %w", err)
	}

	script := make([]byte, l)
	if n, err := io.ReadFull(r, script); uint64(n) != l || err != nil {
		return nil, fmt.Errorf("Could not read script(%d), got %d bytes and err: %w", l, n, err)
	}

	sequence := make([]byte, 4)
	if n, err := io.ReadFull(r, sequence); n != 4 || err != nil {
		return nil, fmt.Errorf("Could not read sequence(4), got %d bytes and err: %w", n, err)
	}

	return &Input{
		previousTxID:       ReverseBytes(previousTxID),
		PreviousTxOutIndex: binary.LittleEndian.Uint32(prevIndex),
		UnlockingScript:    bscript.NewFromBytes(script),
		SequenceNumber:     binary.LittleEndian.Uint32(sequence),
	}, nil
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
	prevTxIDBytes := pvsTx.TxIDBytes()
	for i, utxo := range pvsTx.Outputs {
		utxoPkHASH160, err := utxo.LockingScript.PublicKeyHash()
		if err != nil {
			return err
		}

		if bytes.Equal(utxoPkHASH160, crypto.Hash160(matchPK)) {
			if err := tx.FromUTXOs(&UTXO{
				TxID:          prevTxIDBytes,
				Vout:          uint32(i),
				Satoshis:      utxo.Satoshis,
				LockingScript: utxo.LockingScript,
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
func (tx *Tx) From(prevTxID string, vout uint32, prevTxLockingScript string, satoshis uint64) error {
	pts, err := bscript.NewFromHexString(prevTxLockingScript)
	if err != nil {
		return err
	}
	pti, err := hex.DecodeString(prevTxID)
	if err != nil {
		return err
	}

	return tx.FromUTXOs(&UTXO{
		TxID:          pti,
		Vout:          vout,
		LockingScript: pts,
		Satoshis:      satoshis,
	})
}

// FromUTXOs adds a new input to the transaction from the specified *bt.UTXO fields, using the default
// finalised sequence number (0xFFFFFFFF). If you want a different nSeq, change it manually
// afterwards.
func (tx *Tx) FromUTXOs(utxos ...*UTXO) error {
	for _, utxo := range utxos {
		i := &Input{
			PreviousTxOutIndex: utxo.Vout,
			PreviousTxSatoshis: utxo.Satoshis,
			PreviousTxScript:   utxo.LockingScript,
			SequenceNumber:     DefaultSequenceNumber, // use default finalised sequence number
		}
		if err := i.PreviousTxIDAdd(utxo.TxID); err != nil {
			return err
		}

		tx.addInput(i)
	}

	return nil
}

// Fund continuously calls the provided bt.UTXOGetterFunc, adding each returned input
// as an input via tx.From(...), until it is estimated that inputs cover the outputs + fees.
//
// After completion, the receiver is ready for `Change(...)` to be called, and then be signed.
// Note, this function works under the assumption that receiver *bt.Tx already has all the outputs
// which need covered.
//
// Example usage, for when working with a list:
//    tx.Fund(ctx, bt.NewFeeQuote(), func(ctx context.Context, deficit satoshis) ([]*bt.UTXO, error) {
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
func (tx *Tx) Fund(ctx context.Context, fq *FeeQuote, next UTXOGetterFunc) error {
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
			if err = tx.FromUTXOs(utxo); err != nil {
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
