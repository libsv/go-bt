package bt

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"

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
		PreviousTxOutIndex: prevTxIndex,
		PreviousTxSatoshis: prevTxSats,
		PreviousTxScript:   pts,
		SequenceNumber:     nSeq,
	}, nil
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

// ToBytes encodes the Input into a hex byte array.
func (i *Input) ToBytes(clear bool) []byte {
	h := make([]byte, 0)

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
