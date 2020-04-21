package transaction

import (
	"encoding/binary"
	"encoding/hex"
	"errors"

	"github.com/libsv/libsv/crypto"
	"github.com/libsv/libsv/script"
	"github.com/libsv/libsv/transaction/input"
	"github.com/libsv/libsv/transaction/output"
	"github.com/libsv/libsv/utils"

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
	SighashAllForkID    = 0x00000001 | 0x00000040
)

// A BitcoinTransaction wraps a bitcoin transaction
type BitcoinTransaction struct {
	Bytes    []byte
	Version  uint32
	Witness  bool
	Inputs   []*input.Input
	Outputs  []*output.Output
	Locktime uint32
}

// NewFromString takes a h string representation of a bitcoin transaction
// and returns a BitcoinTransaction object.
func NewFromString(str string) (*BitcoinTransaction, error) {
	bytes, err := hex.DecodeString(str)
	if err != nil {
		return nil, err
	}

	return NewFromBytes(bytes), nil
}

// NewFromBytes takes an array of bytes and constructs a BitcoinTransaction.
func NewFromBytes(bytes []byte) *BitcoinTransaction {
	bt, _ := NewFromBytesWithUsed(bytes)
	return bt
}

// NewFromBytesWithUsed takes an array of bytes and constructs a BitcoinTransaction
// and returns the offset (length of tx).
func NewFromBytesWithUsed(bytes []byte) (*BitcoinTransaction, int) {
	if len(bytes) < 10 {
		// Even an empty transaction has 10 bytes.
		return nil, 0
	}

	bt := BitcoinTransaction{}

	var offset = 0

	bt.Version = binary.LittleEndian.Uint32(bytes[offset:4])
	offset += 4

	// There is an optional Flag of 2 bytes after the version. It is always "0001".
	if bytes[4] == 0x00 && bytes[5] == 0x01 {
		bt.Witness = true
		offset += 2
	}

	inputCount, size := utils.DecodeVarInt(bytes[offset:])
	offset += size

	var i uint64
	for ; i < inputCount; i++ {
		i, size := input.NewInputFromBytes(bytes[offset:])
		offset += size

		bt.Inputs = append(bt.Inputs, i)
	}

	outputCount, size := utils.DecodeVarInt(bytes[offset:])
	offset += size

	for i = 0; i < outputCount; i++ {
		o, size := output.NewFromBytes(bytes[offset:])
		offset += size
		bt.Outputs = append(bt.Outputs, o)
	}

	bt.Locktime = binary.LittleEndian.Uint32(bytes[offset:])
	offset += 4

	bt.Bytes = bytes[0:offset]

	return &bt, offset
}

// HasWitnessData returns true if the optional Witness flag == 0001
func (bt *BitcoinTransaction) HasWitnessData() bool {
	return bt.Witness
}

// AddInput adds a new input to the transaction.
func (bt *BitcoinTransaction) AddInput(input *input.Input) {
	bt.Inputs = append(bt.Inputs, input)
}

// AddUTXO function
func (bt *BitcoinTransaction) AddUTXO(txID string, vout uint32, scriptSig string, satoshis uint64) error {
	i := &input.Input{
		PreviousTxOutIndex: vout,
		PreviousTxScript:   script.NewFromString(scriptSig),
		PreviousTxSatoshis: satoshis,
	}

	h, err := hex.DecodeString(txID)
	if err != nil {
		return err
	}
	copy(i.PreviousTxHash[:], h)

	bt.AddInput(i)

	return nil
}

// InputCount returns the number of transaction inputs.
func (bt *BitcoinTransaction) InputCount() int {
	return len(bt.Inputs)
}

// OutputCount returns the number of transaction inputs.
func (bt *BitcoinTransaction) OutputCount() int {
	return len(bt.Outputs)
}

// AddOutput adds a new output to the transaction.
func (bt *BitcoinTransaction) AddOutput(output *output.Output) {
	bt.Outputs = append(bt.Outputs, output)
}

// PayTo function
func (bt *BitcoinTransaction) PayTo(addr string, satoshis uint64) error {
	o, err := output.NewP2PKHFromAddress(addr, satoshis)
	if err != nil {
		return err
	}

	bt.AddOutput(o)
	return nil
}

// IsCoinbase determines if this transaction is a coinbase by
// seeing if any of the inputs have no inputs.
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

// GetInputs returns an array of all inputs in the transaction.
func (bt *BitcoinTransaction) GetInputs() []*input.Input {
	return bt.Inputs
}

// GetOutputs returns an array of all outputs in the transaction.
func (bt *BitcoinTransaction) GetOutputs() []*output.Output {
	return bt.Outputs
}

// GetTxID returns the transaction ID of the transaction
// (which is also the transaction hash).
func (bt *BitcoinTransaction) GetTxID() string {
	return hex.EncodeToString(utils.ReverseBytes(crypto.Sha256d(bt.Hex())))
}

// Hex encodes the transaction into a h byte array.
// See https://chainquery.com/bitcoin-cli/decoderawtransaction
func (bt *BitcoinTransaction) Hex() []byte {
	return bt.h(0, nil)
}

// HexWithClearedInputs encodes the transaction into a h byte array but clears its inputs first.
// This is used when signing transactions.
func (bt *BitcoinTransaction) HexWithClearedInputs(index int, scriptPubKey []byte) []byte {
	return bt.h(index, scriptPubKey)
}

// GetSighashPayload assembles a payload of sighases for this TX, to be submitted to signing service.
func (bt *BitcoinTransaction) GetSighashPayload(sigType uint32) (*SigningPayload, error) {
	signingPayload, err := NewSigningPayloadFromTx(bt, sigType)
	if err != nil {
		return nil, err
	}
	return signingPayload, nil
}

func (bt *BitcoinTransaction) h(index int, scriptPubKey []byte) []byte {
	h := make([]byte, 0)

	h = append(h, utils.GetLittleEndianBytes(bt.Version, 4)...)

	if bt.Witness {
		h = append(h, 0x00)
		h = append(h, 0x01)
	}

	h = append(h, utils.VarInt(uint64(len(bt.GetInputs())))...)

	for i, in := range bt.GetInputs() {
		s := in.Hex(scriptPubKey != nil)
		if i == index && scriptPubKey != nil {
			h = append(h, utils.VarInt(uint64(len(scriptPubKey)))...)
			h = append(h, scriptPubKey...)
		} else {
			h = append(h, s...)
		}
	}

	h = append(h, utils.VarInt(uint64(len(bt.GetOutputs())))...)
	for _, out := range bt.GetOutputs() {
		h = append(h, out.Hex()...)
	}

	lt := make([]byte, 4)
	binary.LittleEndian.PutUint32(lt, bt.Locktime)
	h = append(h, lt...)

	return h
}

// ApplySignatures applies the signatures passed in through SigningPayload parameter to the transaction inputs
// The signing payload from the signing service should contain a signing item for each of the tx inputs.
// If the TX input does not belong to us, its signature will be blank unless its owner has already signed it.
// If the signing payload contains a signature for a given input, we apply that to the tx regardless of whether we own it or not.
func (bt *BitcoinTransaction) ApplySignatures(signingPayload *SigningPayload, sigType uint32) error {
	if sigType == 0 {
		sigType = SighashAllForkID
	}

	if len(*signingPayload) != len(bt.GetInputs()) {
		return errors.New("error - signing payload number of items does not equal number of inputs")
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
					return err
				}
				if hex.EncodeToString(txPubKeyHash) != signingItem.PublicKeyHash {
					return errors.New("error public key hash from signing payload does not match tx")
				}
			}

			sigBytes, err := hex.DecodeString(signingItem.Signature)
			pubKeyBytes, err := hex.DecodeString(signingItem.PublicKey)
			if err != nil {
				return err
			}

			const sigTypeLength = 1 // Include sighash all fork id hash type when we count length of signature.
			buf := make([]byte, 0)
			buf = append(buf, utils.VarInt(uint64(len(sigBytes)+sigTypeLength))...)
			buf = append(buf, sigBytes...)
			buf = append(buf, SighashAll|SighashForkID)
			buf = append(buf, utils.VarInt(uint64(len(signingItem.PublicKey)/2))...)
			buf = append(buf, pubKeyBytes...)
			bt.Inputs[index].UnlockingScript = script.NewFromBytes(buf)
			sigsApplied++
		}
	}
	if sigsApplied == 0 {
		return errors.New("error - libsv found no signatures in signingPayload to apply to this tx")
	}
	return nil
}

// Sign the transaction
// Normally we'd expect the signing service to do this, but we include this for testing purposes
func (bt *BitcoinTransaction) Sign(privateKey *btcec.PrivateKey, sigType uint32) error {
	if sigType == 0 {
		sigType = SighashAllForkID
	}

	payload, err := bt.GetSighashPayload(sigType)
	if err != nil {
		return err
	}
	signedPayload, err := submitToDummySigningService(payload, privateKey)
	if err != nil {
		return err
	}
	err = bt.ApplySignatures(signedPayload, sigType)
	if err != nil {
		return err
	}
	return nil
}

// submitToDummySigningService local service for testing, which can sign payloads like the signing service.
func submitToDummySigningService(payload *SigningPayload, privateKey *btcec.PrivateKey) (*SigningPayload, error) {
	for _, signingItem := range *payload {
		h, err := hex.DecodeString(signingItem.SigHash)
		if err != nil {
			return nil, err
		}
		sig, err := privateKey.Sign(utils.ReverseBytes(h))
		if err != nil {
			return nil, err
		}
		pubkey := privateKey.PubKey().SerializeCompressed()
		signingItem.PublicKey = hex.EncodeToString(pubkey)
		signingItem.Signature = hex.EncodeToString(sig.Serialize())
	}
	return payload, nil
}

// ApplySignaturesWithoutP2PKHCheck applies signatures without checking if the input previous script equals
// to a P2PKH script matching the private key (see func SignWithoutP2PKHCheck below)
func (bt *BitcoinTransaction) ApplySignaturesWithoutP2PKHCheck(signingPayload *SigningPayload, sigType uint32) error {
	if sigType == 0 {
		sigType = SighashAllForkID
	}

	if len(*signingPayload) != len(bt.GetInputs()) {
		return errors.New("error - signing payload number of items does not equal number of inputs")
	}

	sigsApplied := 0

	for index, signingItem := range *signingPayload {
		// Only use the items which have a pub key and signature in the payload
		if signingItem.Signature != "" && signingItem.PublicKey != "" {
			sigBytes, err := hex.DecodeString(signingItem.Signature)
			pubKeyBytes, err := hex.DecodeString(signingItem.PublicKey)
			if err != nil {
				return err
			}

			const sigTypeLength = 1 // Include sighash all fork id hash type when we count length of signature.
			buf := make([]byte, 0)
			buf = append(buf, utils.VarInt(uint64(len(sigBytes)+sigTypeLength))...)
			buf = append(buf, sigBytes...)
			buf = append(buf, SighashAll|SighashForkID)
			buf = append(buf, utils.VarInt(uint64(len(signingItem.PublicKey)/2))...)
			buf = append(buf, pubKeyBytes...)
			bt.Inputs[index].UnlockingScript = script.NewFromBytes(buf)
			sigsApplied++
		}
	}
	if sigsApplied == 0 {
		return errors.New("error - libsv found no signatures in signingPayload to apply to this tx")
	}
	return nil
}

// SignWithoutP2PKHCheck signs the transaction without checking if the input previous script equals
// to a P2PKH script matching the private key
func (bt *BitcoinTransaction) SignWithoutP2PKHCheck(privateKey *btcec.PrivateKey, sigType uint32) error {
	if sigType == 0 {
		sigType = SighashAllForkID
	}

	payload, err := bt.GetSighashPayload(sigType)
	if err != nil {
		return err
	}
	signedPayload, err := submitToDummySigningService(payload, privateKey)
	if err != nil {
		return err
	}
	err = bt.ApplySignaturesWithoutP2PKHCheck(signedPayload, sigType)
	if err != nil {
		return err
	}
	return nil
}
