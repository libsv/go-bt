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
	PreviousTxID       [32]byte
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
func NewFromBytes(bytes []byte) (*Input, int) {
	i := Input{}

	copy(i.PreviousTxID[:], utils.ReverseBytes(bytes[0:32]))

	i.PreviousTxOutIndex = binary.LittleEndian.Uint32(bytes[32:36])

	offset := 36
	l, size := utils.DecodeVarInt(bytes[offset:])
	offset += size

	i.UnlockingScript = script.NewFromBytes(bytes[offset : offset+int(l)])

	i.SequenceNumber = binary.LittleEndian.Uint32(bytes[offset+int(l):])

	return &i, offset + int(l) + 4
}

// NewFromUTXO returns a transaction input from the UTXO fields provided.
func NewFromUTXO(prevTxID string, prevTxIndex uint32, prevTxSats uint64, prevTxScript string, nSeq uint32) (*Input, error) {
	var b32 [32]byte
	b, _ := hex.DecodeString(prevTxID)
	copy(b32[:], b[0:32])

	pts, err := script.NewFromHexString(prevTxScript)
	if err != nil {
		return nil, err
	}

	i := &Input{
		PreviousTxID:       b32,
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

// Hex encodes the Input into a hex byte array.
func (i *Input) Hex(clear bool) []byte {
	hex := make([]byte, 0)

	hex = append(hex, utils.ReverseBytes(i.PreviousTxID[:])...)
	hex = append(hex, utils.GetLittleEndianBytes(i.PreviousTxOutIndex, 4)...)
	if clear {
		hex = append(hex, 0x00)
	} else {
		if i.UnlockingScript == nil {
			hex = append(hex, utils.VarInt(0)...)
		} else {
			hex = append(hex, utils.VarInt(uint64(len(*i.UnlockingScript)))...)
			hex = append(hex, *i.UnlockingScript...)
		}
	}
	hex = append(hex, utils.GetLittleEndianBytes(i.SequenceNumber, 4)...)

	return hex
}
