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
	PreviousTxID       string
	PreviousTxSatoshis uint64
	PreviousTxScript   *bscript.Script
	UnlockingScript    *bscript.Script
	PreviousTxOutIndex uint32
	SequenceNumber     uint32
}

// NewInput creates a new empty Input object with a finalized sequence number.
func NewInput() *Input {
	b := make([]byte, 0)
	s := bscript.NewFromBytes(b)

	return &Input{
		UnlockingScript: s,
		SequenceNumber:  0xFFFFFFFF,
	}
}

// NewInputFromBytes returns a transaction input from the bytes provided.
func NewInputFromBytes(bytes []byte) (*Input, int, error) {
	if len(bytes) < 36 {
		return nil, 0, fmt.Errorf("input length too short < 36")
	}

	i := Input{}

	i.PreviousTxID = hex.EncodeToString(ReverseBytes(bytes[0:32]))

	i.PreviousTxOutIndex = binary.LittleEndian.Uint32(bytes[32:36])

	offset := 36
	l, size := DecodeVarInt(bytes[offset:])
	offset += size

	totalLength := offset + int(l) + 4 // 4 bytes for nSeq

	if len(bytes) < totalLength {
		return nil, 0, fmt.Errorf("input length too short < 36 + script + 4")
	}

	i.UnlockingScript = bscript.NewFromBytes(bytes[offset : offset+int(l)])

	i.SequenceNumber = binary.LittleEndian.Uint32(bytes[offset+int(l):])

	return &i, totalLength, nil
}

// NewInputFromUTXO returns a transaction input from the UTXO fields provided.
func NewInputFromUTXO(prevTxID string, prevTxIndex uint32, prevTxSats uint64,
	prevTxScript string, nSeq uint32) (*Input, error) {
	pts, err := bscript.NewFromHexString(prevTxScript)
	if err != nil {
		return nil, err
	}

	return &Input{
		PreviousTxID:       prevTxID,
		PreviousTxOutIndex: prevTxIndex,
		PreviousTxSatoshis: prevTxSats,
		PreviousTxScript:   pts,
		SequenceNumber:     nSeq,
	}, nil
}

func (i *Input) String() string {
	return fmt.Sprintf(`prevTxHash:   %x
prevOutIndex: %d
scriptLen:    %d
script:       %x
sequence:     %x
`, i.PreviousTxID, i.PreviousTxOutIndex, len(*i.UnlockingScript), i.UnlockingScript, i.SequenceNumber)
}

// ToBytes encodes the Input into a hex byte array.
func (i *Input) ToBytes(clear bool) []byte {
	h := make([]byte, 0)

	// todo: not checking error
	pid, _ := hex.DecodeString(i.PreviousTxID)

	h = append(h, ReverseBytes(pid)...)
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
	h = append(h, GetLittleEndianBytes(i.SequenceNumber, 4)...)

	return h
}
