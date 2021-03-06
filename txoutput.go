package bt

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"

	"github.com/libsv/go-bt/bscript"
	"github.com/libsv/go-bt/crypto"
)

// NewOutputFromBytes returns a transaction Output from the bytes provided
func NewOutputFromBytes(bytes []byte) (*Output, int, error) {
	if len(bytes) < 8 {
		return nil, 0, fmt.Errorf("output length too short < 8")
	}

	offset := 8
	l, size := DecodeVarInt(bytes[offset:])
	offset += size

	totalLength := offset + int(l)

	if len(bytes) < totalLength {
		return nil, 0, fmt.Errorf("output length too short < 8 + script")
	}

	s := bscript.Script(bytes[offset:totalLength])

	return &Output{
		Satoshis:      binary.LittleEndian.Uint64(bytes[0:8]),
		LockingScript: &s,
	}, totalLength, nil
}

// NewP2PKHOutputFromPubKeyHashStr makes an output to a PKH with a value.
func NewP2PKHOutputFromPubKeyHashStr(publicKeyHash string, satoshis uint64) (*Output, error) {
	s, err := bscript.NewP2PKHFromPubKeyHashStr(publicKeyHash)
	if err != nil {
		return nil, err
	}

	return &Output{
		Satoshis:      satoshis,
		LockingScript: s,
	}, nil
}

// NewP2PKHOutputFromPubKeyBytes makes an output to a PKH with a value.
func NewP2PKHOutputFromPubKeyBytes(publicKeyBytes []byte, satoshis uint64) (*Output, error) {
	s, err := bscript.NewP2PKHFromPubKeyBytes(publicKeyBytes)
	if err != nil {
		return nil, err
	}

	return &Output{
		Satoshis:      satoshis,
		LockingScript: s,
	}, nil
}

// NewP2PKHOutputFromPubKeyStr makes an output to a PKH with a value.
func NewP2PKHOutputFromPubKeyStr(publicKey string, satoshis uint64) (*Output, error) {
	s, err := bscript.NewP2PKHFromPubKeyStr(publicKey)
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
func NewHashPuzzleOutput(secret, publicKeyHash string, satoshis uint64) (*Output, error) {

	publicKeyHashBytes, err := hex.DecodeString(publicKeyHash)
	if err != nil {
		return nil, err
	}

	s := &bscript.Script{}

	s.AppendOpCode(bscript.OpHASH160)
	secretBytesHash := crypto.Hash160([]byte(secret))

	if err = s.AppendPushData(secretBytesHash); err != nil {
		return nil, err
	}
	s.AppendOpCode(bscript.OpEQUALVERIFY)
	s.AppendOpCode(bscript.OpDUP)
	s.AppendOpCode(bscript.OpHASH160)

	if err = s.AppendPushData(publicKeyHashBytes); err != nil {
		return nil, err
	}
	s.AppendOpCode(bscript.OpEQUALVERIFY)
	s.AppendOpCode(bscript.OpCHECKSIG)

	return &Output{
		Satoshis:      satoshis,
		LockingScript: s,
	}, nil
}

// NewOpReturnOutput creates a new Output with OP_FALSE OP_RETURN and then the data
// passed in encoded as hex.
func NewOpReturnOutput(data []byte) (*Output, error) {
	return createOpReturnOutput([][]byte{data})
}

// NewOpReturnPartsOutput creates a new Output with OP_FALSE OP_RETURN and then
// uses OP_PUSHDATA format to encode the multiple byte arrays passed in.
func NewOpReturnPartsOutput(data [][]byte) (*Output, error) {
	return createOpReturnOutput(data)
}

func createOpReturnOutput(data [][]byte) (*Output, error) {
	s := &bscript.Script{}

	s.AppendOpCode(bscript.OpFALSE)
	s.AppendOpCode(bscript.OpRETURN)
	err := s.AppendPushDataArray(data)
	if err != nil {
		return nil, err
	}

	return &Output{LockingScript: s}, nil
}
