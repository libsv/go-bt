package bt

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"

	"github.com/libsv/go-bt/v2/bscript"
	"github.com/pkg/errors"
)

/*
General format (inside a block) of each output of a transaction - Txout
Field	                        Description	                                Size
-----------------------------------------------------------------------------------------------------
value                         non-negative integer giving the number of   8 bytes
                              Satoshis(BTC/10^8) to be transferred
Txout-script length           non-negative integer                        1 - 9 bytes VI = VarInt
Txout-script / scriptPubKey   Script                                      <out-script length>-many bytes
(lockingScript)

*/

// Output is a representation of a transaction output
type Output struct {
	Satoshis      uint64          `json:"satoshis"`
	LockingScript *bscript.Script `json:"locking_script"`
}

// ReadFrom reads from the `io.Reader` into the `bt.Output`.
func (o *Output) ReadFrom(r io.Reader) (int64, error) {
	*o = Output{}
	var bytesRead int64

	satoshis := make([]byte, 8)
	n, err := io.ReadFull(r, satoshis)
	bytesRead += int64(n)
	if err != nil {
		return bytesRead, errors.Wrapf(err, "satoshis(8): got %d bytes", n)
	}

	var l VarInt
	n64, err := l.ReadFrom(r)
	bytesRead += n64
	if err != nil {
		return bytesRead, err
	}

	script := make([]byte, l)
	n, err = io.ReadFull(r, script)
	bytesRead += int64(n)
	if err != nil {
		return bytesRead, errors.Wrapf(err, "lockingScript(%d): got %d bytes", l, n)
	}

	o.Satoshis = binary.LittleEndian.Uint64(satoshis)
	o.LockingScript = bscript.NewFromBytes(script)

	return bytesRead, nil
}

// LockingScriptHexString returns the locking script
// of an output encoded as a hex string.
func (o *Output) LockingScriptHexString() string {
	return hex.EncodeToString(*o.LockingScript)
}

func (o *Output) String() string {
	return fmt.Sprintf(`value:     %d
scriptLen: %d
script:    %s
`, o.Satoshis, len(*o.LockingScript), o.LockingScript)
}

// Bytes encodes the Output into a byte array.
func (o *Output) Bytes() []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, o.Satoshis)

	h := make([]byte, 0)
	h = append(h, b...)
	h = append(h, VarInt(uint64(len(*o.LockingScript))).Bytes()...)
	h = append(h, *o.LockingScript...)

	return h
}

// BytesForSigHash returns the proper serialisation
// of an output to be hashed and signed (sighash).
func (o *Output) BytesForSigHash() []byte {
	buf := make([]byte, 0)

	satoshis := make([]byte, 8)
	binary.LittleEndian.PutUint64(satoshis, o.Satoshis)
	buf = append(buf, satoshis...)

	buf = append(buf, VarInt(uint64(len(*o.LockingScript))).Bytes()...)
	buf = append(buf, *o.LockingScript...)

	return buf
}

// NodeJSON returns a wrapped *bt.Output for marshalling/unmarshalling into a node output format.
//
// Marshalling usage example:
//  bb, err := json.Marshal(output.NodeJSON())
//
// Unmarshalling usage example:
//  output := &bt.Output{}
//  if err := json.Unmarshal(bb, output.NodeJSON()); err != nil {}
func (o *Output) NodeJSON() interface{} {
	return &nodeOutputWrapper{Output: o}
}
