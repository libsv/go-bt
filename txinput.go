package bt

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"

	"github.com/libsv/go-bt/bscript"
	"github.com/libsv/go-bt/crypto"
)

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
		PreviousTxIDBytes:  ReverseBytes(bytes[0:32]),
		PreviousTxOutIndex: binary.LittleEndian.Uint32(bytes[32:36]),
		SequenceNumber:     binary.LittleEndian.Uint32(bytes[offset+int(l):]),
		UnlockingScript:    bscript.NewFromBytes(bytes[offset : offset+int(l)]),
	}, totalLength, nil
}

func (tx *Tx) addInput(input *Input) {
	tx.Inputs = append(tx.Inputs, input)
}

// AddInputFromTx take all outputs from previous transaction
// that match a specific public key, add it as input to this new transaction.
func (tx *Tx) AddInputFromTx(pvsTx *Tx, matchPK []byte) error {
	matchPKHASH160 := crypto.Hash160(matchPK)
	for i, utxo := range pvsTx.Outputs {
		utxoPkHASH160, errPK := utxo.LockingScript.PublicKeyHash()
		if errPK != nil {
			return errPK
		}
		if !bytes.Equal(utxoPkHASH160, matchPKHASH160) {
			continue
		}
		tx.addInput(&Input{
			PreviousTxIDBytes:  pvsTx.TxIDAsBytes(),
			PreviousTxOutIndex: uint32(i),
			PreviousTxSatoshis: utxo.Satoshis,
			PreviousTxScript:   utxo.LockingScript,
			SequenceNumber:     0xffffffff,
		})
	}
	return nil
}

// From adds a new input to the transaction from the specified UTXO fields, using the default
// finalized sequence number (0xFFFFFFFF). If you want a different nSeq, change it manually
// afterwards.
func (tx *Tx) From(prevTxID string, vout uint32, prevTxLockingScript string, satoshis uint64) error {
	pts, err := bscript.NewFromHexString(prevTxLockingScript)
	if err != nil {
		return err
	}

	ptxid, err := hex.DecodeString(prevTxID)
	if err != nil {
		return err
	}

	tx.addInput(&Input{
		PreviousTxIDBytes:  ptxid,
		PreviousTxOutIndex: vout,
		PreviousTxSatoshis: satoshis,
		PreviousTxScript:   pts,
		SequenceNumber:     DefaultSequenceNumber, // use default finalized sequence number
	})

	return nil
}

// InputCount returns the number of transaction inputs.
func (tx *Tx) InputCount() int {
	return len(tx.Inputs)
}
