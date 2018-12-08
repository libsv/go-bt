package transaction

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"

	"bitbucket.org/simon_ordish/cryptolib"
)

/*
General format of a Bitcoin transaction (inside a block)
--------------------------------------------------------
Field            Description                                                               Size

Version no	     currently 1	                                                             4 bytes

Flag	           If present, always 0001, and indicates the presence of witness data       optional 2 byte array

In-counter  	   positive integer VI = VarInt                                              1 - 9 bytes

list of inputs	 the first input of the first transaction is also called "coinbase"        <in-counter>-many inputs
                 (its content was ignored in earlier versions)

Out-counter    	 positive integer VI = VarInt                                              1 - 9 bytes

list of outputs  the outputs of the first transaction spend the mined                      <out-counter>-many outputs
								 bitcoins for the block

Witnesses        A list of witnesses, 1 for each input, omitted if flag above is missing	 variable, see Segregated_Witness

lock_time        if non-zero and sequence numbers are < 0xFFFFFFFF: block height or        4 bytes
                 timestamp when transaction is final
*/

// A BitcoinTransaction wraps a bitcoin transaction
type BitcoinTransaction struct {
	Bytes   []byte
	Witness bool
	Inputs  []*Input
	Outputs []*Output
}

// New comment
func New() *BitcoinTransaction {
	return &BitcoinTransaction{}
}

// NewFromString takes a hex string representation of a bitcoin transaction
// and returns a BitcoinTransaction object
func NewFromString(str string) (*BitcoinTransaction, error) {
	bytes, err := hex.DecodeString(str)
	if err != nil {
		return nil, err
	}

	return NewFromBytes(bytes)
}

// NewFromBytes takes an array of bytes and constructs a BitcoinTransaction
func NewFromBytes(bytes []byte) (*BitcoinTransaction, error) {
	bt := BitcoinTransaction{
		Bytes: bytes,
	}

	var offset = 4 // Skip 4 version bytes

	// There is an optional Flag of 2 bytes after the version. It is always "0001".
	if bytes[4] == 0x00 && bytes[5] == 0x01 {
		bt.Witness = true
		offset += 2
	}

	inputCount, size := cryptolib.DecodeVarInt(bt.Bytes[offset:])
	offset += size

	var i uint64
	for ; i < inputCount; i++ {
		input, size := NewInput(bt.Bytes[offset:])
		offset += size
		bt.Inputs = append(bt.Inputs, input)
	}

	outputCount, size := cryptolib.DecodeVarInt(bt.Bytes[offset:])
	offset += size

	for i = 0; i < outputCount; i++ {
		output, size := NewOutput(bt.Bytes[offset:])
		offset += size
		bt.Outputs = append(bt.Outputs, output)
	}

	return &bt, nil
}

// Version returns the 4 byte version as a uint32 (litte endian)
func (bt *BitcoinTransaction) Version() uint32 {
	bytes := bt.Bytes[0:4]
	return binary.LittleEndian.Uint32(bytes)
}

// VersionHex returns the version of the transaction
func (bt *BitcoinTransaction) VersionHex() string {
	return hex.EncodeToString(bt.Bytes[0:4])
}

// HasWitnessData returns true if the optional Witness flag == 0001
func (bt *BitcoinTransaction) HasWitnessData() bool {
	return bt.Witness
}

// InputCount returns the number of transaction inputs
func (bt *BitcoinTransaction) InputCount() int {
	return len(bt.Inputs)
}

// IsCoinbase determines if this transaction is a coinbase by
// seeing if any of the inputs have no inputs
func (bt *BitcoinTransaction) IsCoinbase() bool {
	if len(bt.Inputs) != 1 {
		return false
	}

	fmt.Println(bt.Inputs[0].previousTxOutIndex)
	fmt.Println(bt.Inputs[0].sequenceNumber)
	for _, v := range bt.Inputs[0].previousTxHash {
		if v != 0x00 {
			return false
		}
	}

	if bt.Inputs[0].previousTxOutIndex == 0xFFFFFFFF || bt.Inputs[0].sequenceNumber == 0xFFFFFFFF {
		return true
	}

	return false
}
