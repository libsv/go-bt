package transaction

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"

	"bitbucket.org/simon_ordish/cryptolib"
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
	Value  uint64
	Script []byte
}

// NewOutput comment
func NewOutput() *Output {
	return &Output{
		Script: make([]byte, 0),
	}
}

// NewOutputForPublicKeyHash makes an output to a PKH with a value.
func NewOutputForPublicKeyHash(publicKeyHash string, satoshis uint64) (*Output, error) {
	o := Output{}
	o.Value = satoshis

	publicKeyHashBytes, err := hex.DecodeString(publicKeyHash)
	if err != nil {
		return nil, err
	}
	script := make([]byte, 0, len(publicKeyHash)+8)
	script = append(script, cryptolib.OpDUP)
	script = append(script, cryptolib.OpHASH160)
	script = append(script, cryptolib.VarInt(uint64(len(publicKeyHash)/2))...)
	script = append(script, publicKeyHashBytes...)
	script = append(script, cryptolib.OpEQUALVERIFY)
	script = append(script, cryptolib.OpCHECKSIG)
	o.Script = script
	return &o, nil
}

// NewOutputFromBytes returns a transaction Output from the bytes provided
func NewOutputFromBytes(bytes []byte) (*Output, int) {
	o := Output{}

	o.Value = binary.LittleEndian.Uint64(bytes[0:8])

	offset := 8
	i, size := cryptolib.DecodeVarInt(bytes[offset:])
	offset += size

	o.Script = bytes[offset : offset+int(i)]

	return &o, offset + int(i)
}

// NewOutputOpReturn comment
func NewOutputOpReturn(data []byte) (*Output, error) {

	b, err := cryptolib.EncodeParts([][]byte{data})
	if err != nil {
		return nil, err
	}
	script := make([]byte, 0)
	script = append(script, cryptolib.OpFALSE)
	script = append(script, cryptolib.OpRETURN)
	script = append(script, b...)

	o := Output{}
	o.Script = script
	return &o, nil
}

// NewOutputOpReturnPush comment
func NewOutputOpReturnPush(data [][]byte) (*Output, error) {

	b, err := cryptolib.EncodeParts(data)
	if err != nil {
		return nil, err
	}

	script := make([]byte, 0)
	script = append(script, cryptolib.OpFALSE)
	script = append(script, cryptolib.OpRETURN)
	script = append(script, b...)

	o := Output{}
	o.Script = script
	return &o, nil
}

// GetOutputScript comment
func (o *Output) GetOutputScript() []byte {
	return o.Script
}

func (o *Output) String() string {
	return fmt.Sprintf(`value:     %d
scriptLen: %d
script:    %x
`, o.Value, len(o.Script), o.Script)
}

// Hex comment
func (o *Output) Hex() []byte {

	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, o.Value)

	hex := make([]byte, 0)
	hex = append(hex, b...)
	hex = append(hex, cryptolib.VarInt(uint64(len(o.Script)))...)
	hex = append(hex, o.Script...)

	return hex
}

func (o *Output) getBytesForSigHash() []byte {
	buf := make([]byte, 0)

	satoshis := make([]byte, 8)
	binary.LittleEndian.PutUint64(satoshis, o.Value)
	buf = append(buf, satoshis...)

	buf = append(buf, cryptolib.VarInt(uint64(len(o.Script)))...)
	buf = append(buf, o.Script...)

	return buf
}
