package bt

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/libsv/go-bt/v2/bscript"
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

// DefaultSequenceNumber is the default starting sequence number
const DefaultSequenceNumber uint32 = 0xFFFFFFFF

// Input is a representation of a transaction input
//
// DO NOT CHANGE ORDER - Optimised for memory via maligned
//
type Input struct {
	previousTxID       []byte
	PreviousTxSatoshis uint64
	PreviousTxScript   *bscript.Script
	UnlockingScript    *bscript.Script
	PreviousTxOutIndex uint32
	SequenceNumber     uint32
}

// inputJSON is used to covnert an input to and from json.
// Script is duplicated as we have our own name for unlockingScript
// but want to be compatible with node json also.
type inputJSON struct {
	UnlockingScript *struct {
		Asm string `json:"asm"`
		Hex string `json:"hex"`
	} `json:"unlockingScript,omitempty"`
	ScriptSig *struct {
		Asm string `json:"asm"`
		Hex string `json:"hex"`
	} `json:"scriptSig,omitempty"`
	TxID     string `json:"txid"`
	Vout     uint32 `json:"vout"`
	Sequence uint32 `json:"sequence"`
}

// MarshalJSON will convert an input to json, expanding upon the
// input struct to add additional fields.
func (i *Input) MarshalJSON() ([]byte, error) {
	asm, err := i.UnlockingScript.ToASM()
	if err != nil {
		return nil, err
	}
	input := &inputJSON{
		TxID: hex.EncodeToString(i.previousTxID),
		Vout: i.PreviousTxOutIndex,
		UnlockingScript: &struct {
			Asm string `json:"asm"`
			Hex string `json:"hex"`
		}{
			Asm: asm,
			Hex: i.UnlockingScript.String(),
		},
		Sequence: i.SequenceNumber,
	}
	return json.Marshal(input)
}

// UnmarshalJSON will convert a JSON input to an input.
func (i *Input) UnmarshalJSON(b []byte) error {
	var ij inputJSON
	if err := json.Unmarshal(b, &ij); err != nil {
		return err
	}
	ptxID, err := hex.DecodeString(ij.TxID)
	if err != nil {
		return err
	}
	sig := ij.UnlockingScript
	if sig == nil {
		sig = ij.ScriptSig
	}
	s, err := bscript.NewFromHexString(sig.Hex)
	if err != nil {
		return err
	}
	i.UnlockingScript = s
	i.previousTxID = ptxID
	i.PreviousTxOutIndex = ij.Vout
	i.SequenceNumber = ij.Sequence
	return nil
}

// PreviousTxIDAdd will add the supplied txID bytes to the Input,
// if it isn't a valid transaction id an ErrInvalidTxID error will be returned.
func (i *Input) PreviousTxIDAdd(txID []byte) error {
	if !IsValidTxID(txID) {
		return ErrInvalidTxID
	}
	i.previousTxID = txID
	return nil
}

// PreviousTxIDAddStr will validate and add the supplied txID string to the Input,
// if it isn't a valid transaction id an ErrInvalidTxID error will be returned.
func (i *Input) PreviousTxIDAddStr(txID string) error {
	bb, err := hex.DecodeString(txID)
	if err != nil {
		return err
	}
	return i.PreviousTxIDAdd(bb)
}

// PreviousTxID will return the PreviousTxID if set.
func (i *Input) PreviousTxID() []byte {
	return i.previousTxID
}

// PreviousTxIDStr returns the Previous TxID as a hex string.
func (i *Input) PreviousTxIDStr() string {
	return hex.EncodeToString(i.previousTxID)
}

// String implements the Stringer interface and returns a string
// representation of a transaction input.
func (i *Input) String() string {
	return fmt.Sprintf(
		`prevTxHash:   %s
prevOutIndex: %d
scriptLen:    %d
script:       %s
sequence:     %x
`,
		hex.EncodeToString(i.previousTxID),
		i.PreviousTxOutIndex,
		len(*i.UnlockingScript),
		i.UnlockingScript,
		i.SequenceNumber,
	)
}

// Bytes encodes the Input into a hex byte array.
func (i *Input) Bytes(clear bool) []byte {
	h := make([]byte, 0)

	h = append(h, ReverseBytes(i.previousTxID)...)
	h = append(h, LittleEndianBytes(i.PreviousTxOutIndex, 4)...)
	if clear {
		h = append(h, 0x00)
	} else {
		if i.UnlockingScript == nil {
			h = append(h, VarInt(0)...)
		} else {
			h = append(h, VarInt(uint64(len(*i.UnlockingScript)))...)
			h = append(h, *i.UnlockingScript...)
		}
	}

	return append(h, LittleEndianBytes(i.SequenceNumber, 4)...)
}
