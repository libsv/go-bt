package transaction

import (
	"encoding/binary"
	"encoding/hex"
	"errors"

	"bitbucket.org/simon_ordish/cryptolib"
	"github.com/btcsuite/btcd/btcec"
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

// Signature constants
const (
	SighashAll          = 0x00000001
	SighashNone         = 0x00000002
	SighashSingle       = 0x00000003
	SighashForkID       = 0x00000040
	SighashAnyoneCanPay = 0x00000080
	SighashAllForkID    = (0x00000001 | 0x00000040)
)

// A BitcoinTransaction wraps a bitcoin transaction
type BitcoinTransaction struct {
	Bytes    []byte
	Version  uint32
	Witness  bool
	Inputs   []*Input
	Outputs  []*Output
	Locktime uint32
}

// New comment
func New() *BitcoinTransaction {
	return &BitcoinTransaction{
		Version: 1,
	}
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

	var offset = 0

	bt.Version = binary.LittleEndian.Uint32(bytes[offset:4])
	offset += 4

	// There is an optional Flag of 2 bytes after the version. It is always "0001".
	if bytes[4] == 0x00 && bytes[5] == 0x01 {
		bt.Witness = true
		offset += 2
	}

	inputCount, size := cryptolib.DecodeVarInt(bt.Bytes[offset:])
	offset += size

	var i uint64
	for ; i < inputCount; i++ {
		input, size := NewInputFromBytes(bt.Bytes[offset:])
		offset += size

		bt.Inputs = append(bt.Inputs, input)
	}

	outputCount, size := cryptolib.DecodeVarInt(bt.Bytes[offset:])
	offset += size

	for i = 0; i < outputCount; i++ {
		output, size := NewOutputFromBytes(bt.Bytes[offset:])
		offset += size
		bt.Outputs = append(bt.Outputs, output)
	}

	bt.Locktime = binary.LittleEndian.Uint32(bytes[offset:])

	return &bt, nil
}

// HasWitnessData returns true if the optional Witness flag == 0001
func (bt *BitcoinTransaction) HasWitnessData() bool {
	return bt.Witness
}

// AddInput comment
func (bt *BitcoinTransaction) AddInput(input *Input) {
	bt.Inputs = append(bt.Inputs, input)
}

// InputCount returns the number of transaction inputs
func (bt *BitcoinTransaction) InputCount() int {
	return len(bt.Inputs)
}

// OutputCount returns the number of transaction inputs
func (bt *BitcoinTransaction) OutputCount() int {
	return len(bt.Outputs)
}

// AddOutput comment
func (bt *BitcoinTransaction) AddOutput(output *Output) {
	bt.Outputs = append(bt.Outputs, output)
}

// IsCoinbase determines if this transaction is a coinbase by
// seeing if any of the inputs have no inputs
func (bt *BitcoinTransaction) IsCoinbase() bool {
	if len(bt.Inputs) != 1 {
		return false
	}

	for _, v := range bt.Inputs[0].PreviousTxHash {
		if v != 0x00 {
			return false
		}
	}

	if bt.Inputs[0].PreviousTxOutIndex == 0xFFFFFFFF || bt.Inputs[0].SequenceNumber == 0xFFFFFFFF {
		return true
	}

	return false
}

// GetInputs comment
func (bt *BitcoinTransaction) GetInputs() []*Input {
	return bt.Inputs
}

// GetOutputs comment
func (bt *BitcoinTransaction) GetOutputs() []*Output {
	return bt.Outputs
}

// Hex comment
func (bt *BitcoinTransaction) Hex() []byte {
	return bt.hex(0, nil)
}

// HexWithClearedInputs comment
func (bt *BitcoinTransaction) HexWithClearedInputs(index int, scriptPubKey []byte) []byte {
	return bt.hex(index, scriptPubKey)
}

// GetSighashPayload assembles a payload of sighases for this TX, to be submitted to signing service.
func (bt *BitcoinTransaction) GetSighashPayload(sigType uint32) (*SigningPayload, error) {
	signingPayload, err := NewSigningPayloadFromTx(bt)
	if err != nil {
		return nil, err
	}
	return signingPayload, nil
}

func (bt *BitcoinTransaction) hex(index int, scriptPubKey []byte) []byte {
	hex := make([]byte, 0)

	hex = append(hex, cryptolib.GetLittleEndianBytes(bt.Version, 4)...)

	if bt.Witness {
		hex = append(hex, 0x00)
		hex = append(hex, 0x01)
	}

	hex = append(hex, cryptolib.VarInt(uint64(len(bt.GetInputs())))...)

	for i, in := range bt.GetInputs() {
		script := in.Hex(scriptPubKey != nil)
		if i == index && scriptPubKey != nil {
			hex = append(hex, cryptolib.VarInt(uint64(len(scriptPubKey)))...)
			hex = append(hex, scriptPubKey...)
		} else {
			hex = append(hex, script...)
		}
	}

	hex = append(hex, cryptolib.VarInt(uint64(len(bt.GetOutputs())))...)
	for _, out := range bt.GetOutputs() {
		hex = append(hex, out.Hex()...)
	}

	lt := make([]byte, 4)
	binary.LittleEndian.PutUint32(lt, bt.Locktime)
	hex = append(hex, lt...)

	return hex
}

// Sign comment
func (bt *BitcoinTransaction) Sign(privateKey *btcec.PrivateKey, sigType uint32) *BitcoinTransaction {
	privKeys := make([]*btcec.PrivateKey, 0)
	privKeys = append(privKeys, privateKey)

	if sigType == 0 {
		sigType = SighashAllForkID
	}

	sigs, _ := GetSignatures(bt, privKeys, sigType)
	for _, sig := range sigs {
		pubkey := privateKey.PubKey().SerializeCompressed()
		buf := make([]byte, 0)
		buf = append(buf, cryptolib.VarInt(uint64(len(sig.Signature)))...)
		buf = append(buf, sig.Signature...)
		buf = append(buf, cryptolib.VarInt(uint64(len(pubkey)))...)
		buf = append(buf, pubkey...)
		bt.GetInputs()[0].SigScript = NewScriptFromBytes(buf)
	}
	return bt
}

// ApplySignatures To sign our transaction, we go to the signing service and get a payload containing sigatures.
// We can then apply those signatures to our transaction inputs to sign the tx.
// The signing payload from the signing service should contain a signing item for each of the tx inputs.
// If the TX input does not belong to us, its signature will be blank unless its owner has already signed it.
// If the signing payload contains a signature for a given input, we apply that to the tx regardless of whether we own it or not.
func (bt *BitcoinTransaction) ApplySignatures(signingPayload *SigningPayload, sigType uint32) (*BitcoinTransaction, error) {
	if sigType == 0 {
		sigType = SighashAllForkID
	}

	if len(*signingPayload) != len(bt.GetInputs()) {
		return nil, errors.New("Error - signing payload number of signing items does not equal signing payload number of items")
	}

	sigsApplied := 0

	for index, signingItem := range *signingPayload {
		// Only use the items which have a pub key and signature in the payload
		if signingItem.Signature != "" && signingItem.PublicKey != "" {
			// If our tx input has a script, check it against our payload pubkeyhash for safety.
			// Note that this is not a complete check as we will probably have the same sighash multiple times in our payload but different sigs.
			// So the order is critical - payload items have a one to one mapping to inputs.
			if bt.Inputs[index].PreviousTxScript != nil {
				txPubKeyHash, err := bt.Inputs[index].PreviousTxScript.GetPublicKeyHash()
				if err != nil {
					return nil, err
				}
				if hex.EncodeToString(txPubKeyHash) != signingItem.PublicKeyHash {
					return nil, errors.New("Error public key hash from signing payload does not match tx")
				}
			}

			sigBytes, err := hex.DecodeString(signingItem.Signature)
			pubKeyBytes, err := hex.DecodeString(signingItem.PublicKey)
			if err != nil {
				return nil, err
			}

			const sigTypeLength = 1 // Include sighash all fork id hash type when we count length of signature.
			buf := make([]byte, 0)
			buf = append(buf, cryptolib.VarInt(uint64(len(sigBytes)+sigTypeLength))...)
			buf = append(buf, sigBytes...)
			buf = append(buf, (SighashAll | SighashForkID))
			buf = append(buf, cryptolib.VarInt(uint64(len(signingItem.PublicKey)/2))...)
			buf = append(buf, pubKeyBytes...)
			bt.Inputs[index].SigScript = NewScriptFromBytes(buf)
			sigsApplied++
		}
		if sigsApplied == 0 {
			return nil, errors.New("Error found no signatures in this payload to apply to this tx")
		}
	}
	return bt, nil
}
