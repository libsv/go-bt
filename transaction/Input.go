package transaction

import (
	"encoding/binary"
	"fmt"

	"cryptolib"

	"github.com/btcsuite/btcd/btcec"
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
	previousTxHash     [32]byte
	previousTxOutIndex uint32
	previousTxAmount   uint64
	script             *Script
	sequenceNumber     uint32
}

// NewInput returns a transaction input from the bytes provided
func NewInput(bytes []byte) (*Input, int) {
	i := Input{}

	copy(i.previousTxHash[:], cryptolib.ReverseBytes(bytes[0:32]))

	i.previousTxOutIndex = binary.LittleEndian.Uint32(bytes[32:36])

	offset := 36
	l, size := cryptolib.DecodeVarInt(bytes[offset:])
	offset += size

	i.script = NewScript(bytes[offset : offset+int(l)])

	i.sequenceNumber = binary.LittleEndian.Uint32(bytes[offset+int(l):])

	return &i, offset + int(l) + 4
}

func (i *Input) String() string {
	return fmt.Sprintf(`prevTxHash:   %x
prevOutIndex: %d
scriptLen:    %d
script:       %x
sequence:     %x
`, i.previousTxHash, i.previousTxOutIndex, len(*i.script), i.script, i.sequenceNumber)
}

// Hex comment
func (i *Input) Hex(clear bool) []byte {
	hex := make([]byte, 0)

	hex = append(hex, cryptolib.ReverseBytes(i.previousTxHash[:])...)
	hex = append(hex, cryptolib.GetLittleEndianBytes(i.previousTxOutIndex, 4)...)
	if clear {
		hex = append(hex, 0x00)
	} else {
		hex = append(hex, cryptolib.VarInt(uint64(len(*i.script)))...)
		hex = append(hex, *i.script...)
	}
	hex = append(hex, cryptolib.GetLittleEndianBytes(i.sequenceNumber, 4)...)

	return hex
}

func (i *Input) getSignatures(transaction *BitcoinTransaction, privKey *btcec.PrivateKey, index int, sigtype int, hashData []byte) {

}
