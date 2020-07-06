package input

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"

	"github.com/libsv/libsv/script"
	"github.com/libsv/libsv/utils"
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
type Input struct {
	PreviousTxID       string
	PreviousTxOutIndex uint32
	PreviousTxSatoshis uint64
	PreviousTxScript   *script.Script
	UnlockingScript    *script.Script
	SequenceNumber     uint32
}

// New creates a new empty Input object with a finalised sequence number.
func New() *Input {
	b := make([]byte, 0)
	s := script.NewFromBytes(b)

	return &Input{
		UnlockingScript: s,
		SequenceNumber:  0xFFFFFFFF,
	}
}

// NewFromBytes returns a transaction input from the bytes provided.
func NewFromBytes(bytes []byte) (*Input, int, error) {
	if len(bytes) < 36 {
		return nil, 0, fmt.Errorf("input length too short < 36")
	}

	i := Input{}

	i.PreviousTxID = hex.EncodeToString(utils.ReverseBytes(bytes[0:32]))

	i.PreviousTxOutIndex = binary.LittleEndian.Uint32(bytes[32:36])

	offset := 36
	l, size := utils.DecodeVarInt(bytes[offset:])
	offset += size

	totalLength := offset + int(l) + 4 // 4 bytes for nSeq

	if len(bytes) < totalLength {
		return nil, 0, fmt.Errorf("input length too short < 36 + script + 4")
	}

	i.UnlockingScript = script.NewFromBytes(bytes[offset : offset+int(l)])

	i.SequenceNumber = binary.LittleEndian.Uint32(bytes[offset+int(l):])

	return &i, totalLength, nil
}

// NewFromUTXO returns a transaction input from the UTXO fields provided.
func NewFromUTXO(prevTxID string, prevTxIndex uint32, prevTxSats uint64, prevTxScript string, nSeq uint32) (*Input, error) {
	pts, err := script.NewFromHexString(prevTxScript)
	if err != nil {
		return nil, err
	}

	i := &Input{
		PreviousTxID:       prevTxID,
		PreviousTxOutIndex: prevTxIndex,
		PreviousTxSatoshis: prevTxSats,
		PreviousTxScript:   pts,
		SequenceNumber:     nSeq,
	}

	return i, nil
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

	pid, _ := hex.DecodeString(i.PreviousTxID)

	h = append(h, utils.ReverseBytes(pid)...)
	h = append(h, utils.GetLittleEndianBytes(i.PreviousTxOutIndex, 4)...)
	if clear {
		h = append(h, 0x00)
	} else {
		if i.UnlockingScript == nil {
			h = append(h, utils.VarInt(0)...)
		} else {
			h = append(h, utils.VarInt(uint64(len(*i.UnlockingScript)))...)
			h = append(h, *i.UnlockingScript...)
		}
	}
	h = append(h, utils.GetLittleEndianBytes(i.SequenceNumber, 4)...)

	return h
}
