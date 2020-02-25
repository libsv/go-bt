package transaction

import (
	"encoding/binary"
	"fmt"

	"bitbucket.org/simon_ordish/cryptolib"
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
	PreviousTxHash     [32]byte
	PreviousTxOutIndex uint32
	PreviousTxSatoshis uint64
	PreviousTxScript   *Script
	SigScript          *Script
	SequenceNumber     uint32
}

// NewInput comment
func NewInput() *Input {
	b := make([]byte, 0)
	s := NewScriptFromBytes(b)

	return &Input{
		SigScript:      s,
		SequenceNumber: 0xFFFFFFFF,
	}
}

// NewInputFromBytes returns a transaction input from the bytes provided
func NewInputFromBytes(bytes []byte) (*Input, int) {
	i := Input{}

	copy(i.PreviousTxHash[:], cryptolib.ReverseBytes(bytes[0:32]))

	i.PreviousTxOutIndex = binary.LittleEndian.Uint32(bytes[32:36])

	offset := 36
	l, size := cryptolib.DecodeVarInt(bytes[offset:])
	offset += size

	i.SigScript = NewScriptFromBytes(bytes[offset : offset+int(l)])

	i.SequenceNumber = binary.LittleEndian.Uint32(bytes[offset+int(l):])

	return &i, offset + int(l) + 4
}

func (i *Input) String() string {
	return fmt.Sprintf(`prevTxHash:   %x
prevOutIndex: %d
scriptLen:    %d
script:       %x
sequence:     %x
`, i.PreviousTxHash, i.PreviousTxOutIndex, len(*i.SigScript), i.SigScript, i.SequenceNumber)
}

// Hex comment
func (i *Input) Hex(clear bool) []byte {
	hex := make([]byte, 0)

	hex = append(hex, cryptolib.ReverseBytes(i.PreviousTxHash[:])...)
	hex = append(hex, cryptolib.GetLittleEndianBytes(i.PreviousTxOutIndex, 4)...)
	if clear {
		hex = append(hex, 0x00)
	} else {
		if i.SigScript == nil {
			hex = append(hex, cryptolib.VarInt(0)...)
		} else {
			hex = append(hex, cryptolib.VarInt(uint64(len(*i.SigScript)))...)
			hex = append(hex, *i.SigScript...)
		}
	}
	hex = append(hex, cryptolib.GetLittleEndianBytes(i.SequenceNumber, 4)...)

	return hex
}
