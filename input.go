package bt

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"

	"github.com/libsv/go-bt/bscript"
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

// Input is a representation of a transaction input
//
// DO NOT CHANGE ORDER - Optimized for memory via maligned
//
type Input struct {
	PreviousTxIDBytes  []byte
	PreviousTxID       string
	PreviousTxSatoshis uint64
	PreviousTxScript   *bscript.Script
	UnlockingScript    *bscript.Script
	PreviousTxOutIndex uint32
	SequenceNumber     uint32
}

// DefaultSequenceNumber is the default starting sequence number
const DefaultSequenceNumber uint32 = 0xFFFFFFFF

// NewInput creates a new empty Input object with a finalized sequence number.
func NewInput() *Input {
	return &Input{
		UnlockingScript: bscript.NewFromBytes(make([]byte, 0)),
		SequenceNumber:  DefaultSequenceNumber,
	}
}

// NewInputFromBytes returns a transaction input from the bytes provided.
func NewInputFromBytes(b []byte) (*Input, int, error) {
	if len(b) < 36 {
		return nil, 0, fmt.Errorf("input length too short < 36")
	}

	r := bytes.NewReader(b)

	i, err := NewInputFromReader(r)
	if err != nil {
		return nil, 0, err
	}

	return i, len(i.ToBytes(false)), nil
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
		PreviousTxIDBytes:  ReverseBytes(previousTxID),
		PreviousTxID:       hex.EncodeToString(ReverseBytes(previousTxID)),
		PreviousTxOutIndex: binary.LittleEndian.Uint32(prevIndex),
		UnlockingScript:    bscript.NewFromBytes(script),
		SequenceNumber:     binary.LittleEndian.Uint32(sequence),
	}, nil
}

// NewInputFromUTXO returns a transaction input from the UTXO fields provided.
func NewInputFromUTXO(prevTxID string, prevTxIndex uint32, prevTxSats uint64,
	prevTxScript string, nSeq uint32) (*Input, error) {

	pts, err := bscript.NewFromHexString(prevTxScript)
	if err != nil {
		return nil, err
	}

	ptxid, err := hex.DecodeString(prevTxID)
	if err != nil {
		return nil, err
	}

	return &Input{
		PreviousTxIDBytes:  ptxid,
		PreviousTxID:       prevTxID,
		PreviousTxOutIndex: prevTxIndex,
		PreviousTxSatoshis: prevTxSats,
		PreviousTxScript:   pts,
		SequenceNumber:     nSeq,
	}, nil
}

func (i *Input) String() string {
	return fmt.Sprintf(
		`prevTxHash:   %x
prevOutIndex: %d
scriptLen:    %d
script:       %x
sequence:     %x
`,
		i.PreviousTxID,
		i.PreviousTxOutIndex,
		len(*i.UnlockingScript),
		i.UnlockingScript,
		i.SequenceNumber,
	)
}

// ToBytes encodes the Input into a hex byte array.
func (i *Input) ToBytes(clear bool) []byte {
	h := make([]byte, 0)

	// TODO: v2 make input (and other internal) elements private and not exposed
	// so that we only store previoustxid in bytes and then do the conversion
	// with getters and setters
	if i.PreviousTxIDBytes == nil {
		ptidb, err := hex.DecodeString(i.PreviousTxID)
		if err == nil {
			i.PreviousTxIDBytes = ptidb
		}
	}

	h = append(h, ReverseBytes(i.PreviousTxIDBytes)...)
	h = append(h, GetLittleEndianBytes(i.PreviousTxOutIndex, 4)...)
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

	return append(h, GetLittleEndianBytes(i.SequenceNumber, 4)...)
}
