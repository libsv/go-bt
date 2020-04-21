package transaction

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"

	"github.com/libsv/libsv/crypto"
	"github.com/libsv/libsv/script"
	"github.com/libsv/libsv/utils"
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
	Value         uint64
	LockingScript []byte
}

// NewOutputFromBytes returns a transaction Output from the bytes provided
func NewOutputFromBytes(bytes []byte) (*Output, int) {
	o := Output{}

	o.Value = binary.LittleEndian.Uint64(bytes[0:8])

	offset := 8
	i, size := utils.DecodeVarInt(bytes[offset:])
	offset += size

	o.LockingScript = bytes[offset : offset+int(i)]

	return &o, offset + int(i)
}

// NewOutputForPublicKeyHash makes an output to a PKH with a value.
func NewOutputForPublicKeyHash(publicKeyHash string, satoshis uint64) (*Output, error) {
	o := Output{}
	o.Value = satoshis

	publicKeyHashBytes, err := hex.DecodeString(publicKeyHash)
	if err != nil {
		return nil, err
	}
	s := make([]byte, 0, len(publicKeyHash)+8)
	s = append(s, script.OpDUP)
	s = append(s, script.OpHASH160)
	s = append(s, utils.VarInt(uint64(len(publicKeyHash)/2))...)
	s = append(s, publicKeyHashBytes...)
	s = append(s, script.OpEQUALVERIFY)
	s = append(s, script.OpCHECKSIG)
	o.LockingScript = s
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
	s := &script.Script{}

	s.AppendOpCode(script.OpHASH160)
	secretBytesHash := crypto.Hash160([]byte(secret))
	err = s.AppendPushDataToScript(secretBytesHash)
	if err != nil {
		return nil, err
	}
	s.AppendOpCode(script.OpEQUALVERIFY)
	s.AppendOpCode(script.OpDUP)
	s.AppendOpCode(script.OpHASH160)
	err = s.AppendPushDataToScript(publicKeyHashBytes)
	if err != nil {
		return nil, err
	}
	s.AppendOpCode(script.OpEQUALVERIFY)
	s.AppendOpCode(script.OpCHECKSIG)

	o.LockingScript = *s
	return &o, nil
}

// NewOutputOpReturn creates a new Output with OP_FALSE OP_RETURN and then the data
// passed in encoded as hex.
func NewOutputOpReturn(data []byte) (*Output, error) {

	b, err := script.EncodeParts([][]byte{data})
	if err != nil {
		return nil, err
	}
	s := make([]byte, 0)
	s = append(s, script.OpFALSE)
	s = append(s, script.OpRETURN)
	s = append(s, b...)

	o := Output{}
	o.LockingScript = s
	return &o, nil
}

// NewOutputOpReturnPush creates a new Output with OP_FALSE OP_RETURN and then
// uses OP_PUSHDATA format to encode the multiple byte arrays passed in.
func NewOutputOpReturnPush(data [][]byte) (*Output, error) {

	b, err := script.EncodeParts(data)
	if err != nil {
		return nil, err
	}

	s := make([]byte, 0)
	s = append(s, script.OpFALSE)
	s = append(s, script.OpRETURN)
	s = append(s, b...)

	o := Output{}
	o.LockingScript = s
	return &o, nil
}

func (o *Output) String() string {
	return fmt.Sprintf(`value:     %d
scriptLen: %d
script:    %x
`, o.Value, len(o.LockingScript), o.LockingScript)
}

// Hex encodes the Output into a byte array.
func (o *Output) Hex() []byte {

	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, o.Value)

	h := make([]byte, 0)
	h = append(h, b...)
	h = append(h, utils.VarInt(uint64(len(o.LockingScript)))...)
	h = append(h, o.LockingScript...)

	return h
}

func (o *Output) getBytesForSigHash() []byte {
	buf := make([]byte, 0)

	satoshis := make([]byte, 8)
	binary.LittleEndian.PutUint64(satoshis, o.Value)
	buf = append(buf, satoshis...)

	buf = append(buf, utils.VarInt(uint64(len(o.LockingScript)))...)
	buf = append(buf, o.LockingScript...)

	return buf
}
