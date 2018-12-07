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
	previousTxHash     [32]byte
	previousTxOutIndex uint32
	txInScriptLength   uint64
	txInScript         []byte
	sequenceNumber     uint32
}

// NewInput returns a transaction input from the bytes provided
func NewInput(bytes []byte) (*Input, int) {
	ti := Input{}

	copy(ti.previousTxHash[:], cryptolib.ReverseBytes(bytes[0:32]))

	ti.previousTxOutIndex = binary.LittleEndian.Uint32(bytes[32:36])

	i, size := cryptolib.DecodeVarInt(bytes[36:])
	ti.txInScriptLength = i

	ti.txInScript = bytes[36+size : 36+size+int(i)]

	ti.sequenceNumber = binary.LittleEndian.Uint32(bytes[36+size+int(i):])

	return &ti, 36 + size + int(i) + 4
}

func (i *Input) String() string {
	return fmt.Sprintf(`prevTxHash:   %x
prevOutIndex: %d
scriptLen:    %d
script:       %x
sequence:     %x
`, i.previousTxHash, i.previousTxOutIndex, i.txInScriptLength, i.txInScript, i.sequenceNumber)
}
