package bt

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"

	"github.com/libsv/go-bt/bscript"
	"github.com/libsv/go-bt/crypto"
)

/*
General format (inside a block) of each output of a transaction - Txout
Field	                        Description	                                Size
-----------------------------------------------------------------------------------------------------
value                         non negative integer giving the number of   8 bytes
                              Satoshis(BTC/10^8) to be transferred
Txout-script length           non negative integer                        1 - 9 bytes VI = VarInt
Txout-script / scriptPubKey   Script                                      <out-script length>-many bytes
(lockingScript)

*/

// Output is a representation of a transaction output
type Output struct {
	Satoshis      uint64
	LockingScript *bscript.Script
}

// NewOutputFromBytes returns a transaction Output from the bytes provided
func NewOutputFromBytes(bytes []byte) (*Output, int, error) {
	if len(bytes) < 8 {
		return nil, 0, fmt.Errorf("output length too short < 8")
	}

	o := Output{}

	o.Satoshis = binary.LittleEndian.Uint64(bytes[0:8])

	offset := 8
	l, size := DecodeVarInt(bytes[offset:])
	offset += size

	totalLength := offset + int(l)

	if len(bytes) < totalLength {
		return nil, 0, fmt.Errorf("output length too short < 8 + script")
	}

	s := bscript.Script(bytes[offset:totalLength])
	o.LockingScript = &s

	return &o, totalLength, nil
}

// NewP2PKHOutputFromPubKeyHash makes an output to a PKH with a value.
func NewP2PKHOutputFromPubKeyHash(publicKeyHash string, satoshis uint64) (*Output, error) {
	s, err := bscript.NewP2PKHFromPubKeyHashStr(publicKeyHash)
	if err != nil {
		return nil, err
	}

	return &Output{
		Satoshis:      satoshis,
		LockingScript: s,
	}, nil
}

// NewP2PKHOutputFromAddress makes an output to a PKH with a value.
func NewP2PKHOutputFromAddress(addr string, satoshis uint64) (*Output, error) {
	s, err := bscript.NewP2PKHFromAddress(addr)
	if err != nil {
		return nil, err
	}

	return &Output{
		Satoshis:      satoshis,
		LockingScript: s,
	}, nil
}

// NewHashPuzzleOutput makes an output to a hash puzzle + PKH with a value.
func NewHashPuzzleOutput(secret string, publicKeyHash string, satoshis uint64) (*Output, error) {
	o := Output{}
	o.Satoshis = satoshis

	publicKeyHashBytes, err := hex.DecodeString(publicKeyHash)
	if err != nil {
		return nil, err
	}

	s := &bscript.Script{}

	s.AppendOpCode(bscript.OpHASH160)
	secretBytesHash := crypto.Hash160([]byte(secret))
	err = s.AppendPushData(secretBytesHash)
	if err != nil {
		return nil, err
	}
	s.AppendOpCode(bscript.OpEQUALVERIFY)
	s.AppendOpCode(bscript.OpDUP)
	s.AppendOpCode(bscript.OpHASH160)
	err = s.AppendPushData(publicKeyHashBytes)
	if err != nil {
		return nil, err
	}
	s.AppendOpCode(bscript.OpEQUALVERIFY)
	s.AppendOpCode(bscript.OpCHECKSIG)

	o.LockingScript = s
	return &o, nil
}

// NewOpReturnOuput creates a new Output with OP_FALSE OP_RETURN and then the data
// passed in encoded as hex.
func NewOpReturnOuput(data []byte) (*Output, error) {
	o, err := createOpReturnOutput([][]byte{data})
	if err != nil {
		return nil, err
	}

	return o, nil
}

// NewOpReturnPartsOutput creates a new Output with OP_FALSE OP_RETURN and then
// uses OP_PUSHDATA format to encode the multiple byte arrays passed in.
func NewOpReturnPartsOutput(data [][]byte) (*Output, error) {
	o, err := createOpReturnOutput(data)
	if err != nil {
		return nil, err
	}

	return o, nil
}

func createOpReturnOutput(data [][]byte) (*Output, error) {
	s := &bscript.Script{}

	s.AppendOpCode(bscript.OpFALSE)
	s.AppendOpCode(bscript.OpRETURN)
	err := s.AppendPushDataArray(data)
	if err != nil {
		return nil, err
	}
	o := Output{}
	o.LockingScript = s
	return &o, nil
}

// GetLockingScriptHexString returns the locking script
// of an output encoded as a hex string.
func (o *Output) GetLockingScriptHexString() string {
	return hex.EncodeToString(*o.LockingScript)
}

func (o *Output) String() string {
	return fmt.Sprintf(`value:     %d
scriptLen: %d
script:    %x
`, o.Satoshis, len(*o.LockingScript), o.LockingScript)
}

// ToBytes encodes the Output into a byte array.
func (o *Output) ToBytes() []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, o.Satoshis)

	h := make([]byte, 0)
	h = append(h, b...)
	h = append(h, VarInt(uint64(len(*o.LockingScript)))...)
	h = append(h, *o.LockingScript...)

	return h
}

// GetBytesForSigHash returns the proper serialization
// of an output to be hashed and signed (sighash).
func (o *Output) GetBytesForSigHash() []byte {
	buf := make([]byte, 0)

	satoshis := make([]byte, 8)
	binary.LittleEndian.PutUint64(satoshis, o.Satoshis)
	buf = append(buf, satoshis...)

	buf = append(buf, VarInt(uint64(len(*o.LockingScript)))...)
	buf = append(buf, *o.LockingScript...)

	return buf
}
