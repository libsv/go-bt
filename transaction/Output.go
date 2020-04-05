package transaction

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/jadwahab/libsv/crypto"
	"github.com/jadwahab/libsv/utils"
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

// NewOutput creates a new Output object.
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
	script = append(script, utils.OpDUP)
	script = append(script, utils.OpHASH160)
	script = append(script, utils.VarInt(uint64(len(publicKeyHash)/2))...)
	script = append(script, publicKeyHashBytes...)
	script = append(script, utils.OpEQUALVERIFY)
	script = append(script, utils.OpCHECKSIG)
	o.Script = script
	return &o, nil
}

// NewOutputForHashPuzzle makes an output to a hash puzzle + PKH with a value.
func NewOutputForHashPuzzle(secret string, publicKeyHash string, satoshis uint64) (*Output, error) {
	o := Output{}
	o.Value = satoshis

	publicKeyHashBytes, err := hex.DecodeString(publicKeyHash)
	if err != nil {
		return nil, err
	}
	s := NewScript()

	s.AppendOpCode(utils.OpHASH160)
	secretBytesHash := crypto.Hash160([]byte(secret))
	s.AppendPushDataToScript(secretBytesHash)
	s.AppendOpCode(utils.OpEQUALVERIFY)

	s.AppendOpCode(utils.OpDUP)
	s.AppendOpCode(utils.OpHASH160)
	s.AppendPushDataToScript(publicKeyHashBytes)
	s.AppendOpCode(utils.OpEQUALVERIFY)
	s.AppendOpCode(utils.OpCHECKSIG)

	o.Script = *s
	return &o, nil
}

// NewOutputFromBytes returns a transaction Output from the bytes provided
func NewOutputFromBytes(bytes []byte) (*Output, int) {
	o := Output{}

	o.Value = binary.LittleEndian.Uint64(bytes[0:8])

	offset := 8
	i, size := utils.DecodeVarInt(bytes[offset:])
	offset += size

	o.Script = bytes[offset : offset+int(i)]

	return &o, offset + int(i)
}

// NewOutputOpReturn creates a new Output with OP_FALSE OP_RETURN and then the data
// passed in encoded as hex.
func NewOutputOpReturn(data []byte) (*Output, error) {

	b, err := utils.EncodeParts([][]byte{data})
	if err != nil {
		return nil, err
	}
	script := make([]byte, 0)
	script = append(script, utils.OpFALSE)
	script = append(script, utils.OpRETURN)
	script = append(script, b...)

	o := Output{}
	o.Script = script
	return &o, nil
}

// NewOutputOpReturnPush creates a new Output with OP_FALSE OP_RETURN and then
// uses OP_PUSHDATA format to encode the multiple byte arrays passed in.
func NewOutputOpReturnPush(data [][]byte) (*Output, error) {

	b, err := utils.EncodeParts(data)
	if err != nil {
		return nil, err
	}

	script := make([]byte, 0)
	script = append(script, utils.OpFALSE)
	script = append(script, utils.OpRETURN)
	script = append(script, b...)

	o := Output{}
	o.Script = script
	return &o, nil
}

// GetOutputScript returns the script of the output
func (o *Output) GetOutputScript() []byte {
	return o.Script
}

func (o *Output) String() string {
	return fmt.Sprintf(`value:     %d
scriptLen: %d
script:    %x
`, o.Value, len(o.Script), o.Script)
}

// Hex encodes the Output into a hex byte array.
func (o *Output) Hex() []byte {

	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, o.Value)

	hex := make([]byte, 0)
	hex = append(hex, b...)
	hex = append(hex, utils.VarInt(uint64(len(o.Script)))...)
	hex = append(hex, o.Script...)

	return hex
}

func (o *Output) getBytesForSigHash() []byte {
	buf := make([]byte, 0)

	satoshis := make([]byte, 8)
	binary.LittleEndian.PutUint64(satoshis, o.Value)
	buf = append(buf, satoshis...)

	buf = append(buf, utils.VarInt(uint64(len(o.Script)))...)
	buf = append(buf, o.Script...)

	return buf
}
