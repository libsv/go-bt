package transaction

import (
	"encoding/binary"
	"fmt"

	"cryptolib"
)

/*
General format (inside a block) of each output of a transaction - Txout
Field	                        Description	                                Size
-----------------------------------------------------------------------------------------------------
value                         non negative integer giving the number of   8 bytes
                              Satoshis(BTC/10^8) to be transfered
Txout-script length           non negative integer                        1 - 9 bytes VI = VarInt
Txout-script / scriptPubKey   Script                                      <out-script length>-many bytes

*/

// Output is a representation of a transaction output
type Output struct {
	value     uint64
	scriptLen uint64
	script    []byte
}

// NewOutput returns a transaction Output from the bytes provided
func NewOutput(bytes []byte) (*Output, int) {
	o := Output{}

	o.value = binary.LittleEndian.Uint64(bytes[0:8])

	offset := 8
	i, size := cryptolib.DecodeVarInt(bytes[offset:])
	o.scriptLen = i
	offset += size

	o.script = bytes[offset : offset+int(i)]

	return &o, offset + int(i)
}

// GetOutputScript comment
func (o *Output) GetOutputScript() []byte {
	return o.script
}

func (o *Output) String() string {
	return fmt.Sprintf(`value:     %d
scriptLen: %d
script:    %x
`, o.value, o.scriptLen, o.script)
}

// Hex comment
func (o *Output) Hex() []byte {

	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, o.value)

	hex := make([]byte, 0)
	hex = append(hex, b...)
	hex = append(hex, cryptolib.VarInt(o.scriptLen)...)
	hex = append(hex, o.script...)

	return hex
}

func (o *Output) getBytesForSigHash() []byte {
	buf := make([]byte, 0)

	satoshis := make([]byte, 8)
	binary.LittleEndian.PutUint64(satoshis, o.value)
	buf = append(buf, satoshis...)

	buf = append(buf, cryptolib.VarInt(o.scriptLen)...)
	buf = append(buf, o.script...)

	return buf
}
