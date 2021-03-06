package bt

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"

	"github.com/libsv/go-bt/bscript"
	"github.com/libsv/go-bt/crypto"
)

/*
Field	                     Description                                                   Size
--------------------------------------------------------------------------------------------------------
Previous Transaction hash  doubled SHA256-hashed of a (previous) to-be-used transaction	 32 bytes
Previous Txout-index       non negative integer indexing an output of the to-be-used      4 bytes
                           transaction
Txin-script length         non negative integer VI = VarInt                               1-9 bytes
Txin-script / scriptSig	   Script	                                                        <in-script length>-many bytes
sequence_no	               normally 0xFFFFFFFF; irrelevant unless transaction's           4 bytes
                           lock_time is > 0
*/

// DefaultSequenceNumber is the default starting sequence number
const DefaultSequenceNumber uint32 = 0xFFFFFFFF

// Input is a representation of a transaction input
//
// DO NOT CHANGE ORDER - Optimized for memory via maligned
//
type Input struct {
	PreviousTxIDBytes  []byte
	PreviousTxSatoshis uint64
	PreviousTxScript   *bscript.Script
	UnlockingScript    *bscript.Script
	PreviousTxOutIndex uint32
	SequenceNumber     uint32
}

// PreviousTxIDStr returns the Previous TxID as a hex string.
func (i *Input) PreviousTxIDStr() string {
	return hex.EncodeToString(i.PreviousTxIDBytes)
}

func (i *Input) String() string {
	return fmt.Sprintf(
		`prevTxHash:   %s
prevOutIndex: %d
scriptLen:    %d
script:       %x
sequence:     %x
`,
		hex.EncodeToString(i.PreviousTxIDBytes),
		i.PreviousTxOutIndex,
		len(*i.UnlockingScript),
		i.UnlockingScript,
		i.SequenceNumber,
	)
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
		PreviousTxIDBytes:  ReverseBytes(bytes[0:32]),
		PreviousTxOutIndex: binary.LittleEndian.Uint32(bytes[32:36]),
		SequenceNumber:     binary.LittleEndian.Uint32(bytes[offset+int(l):]),
		UnlockingScript:    bscript.NewFromBytes(bytes[offset : offset+int(l)]),
	}, totalLength, nil
}

// ToBytes encodes the Input into a hex byte array.
func (i *Input) ToBytes(clear bool) []byte {
	h := make([]byte, 0)

	h = append(h, ReverseBytes(i.PreviousTxIDBytes)...)
	h = append(h, LittleEndianBytes(i.PreviousTxOutIndex, 4)...)
	if clear {
		h = append(h, 0x00)
	} else {
		if i.UnlockingScript == nil {
			h = append(h, VarInt(0)...)
		} else {
			h = append(h, VarInt(uint64(len(*i.UnlockingScript)))...)
			h = append(h, *i.UnlockingScript...)
		}
	}

	return append(h, LittleEndianBytes(i.SequenceNumber, 4)...)
}

// AddInput adds a new input to the transaction.
func (tx *Tx) AddInput(input *Input) {
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
		tx.AddInput(&Input{
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

	tx.AddInput(&Input{
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
