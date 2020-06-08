package transaction

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"github.com/bitcoinsv/bsvd/bsvec"
	"github.com/libsv/libsv/crypto"
	"github.com/libsv/libsv/script"
	"github.com/libsv/libsv/transaction/input"
	"github.com/libsv/libsv/transaction/output"
	"github.com/libsv/libsv/utils"
)

/*
General format of a Bitcoin transaction (inside a block)
--------------------------------------------------------
Field            Description                                                               Size

Version no	     currently 1	                                                             4 bytes

In-counter  	   positive integer VI = VarInt                                              1 - 9 bytes

list of inputs	 the first input of the first transaction is also called "coinbase"        <in-counter>-many inputs
                 (its content was ignored in earlier versions)

Out-counter    	 positive integer VI = VarInt                                              1 - 9 bytes

list of outputs  the outputs of the first transaction spend the mined                      <out-counter>-many outputs
								 bitcoins for the block

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

// A Transaction wraps a bitcoin transaction
type Transaction struct {
	Version  uint32
	Inputs   []*input.Input
	Outputs  []*output.Output
	Locktime uint32
}

// New creates a new transaction object with default values.
func New() *Transaction {
	t := Transaction{}

	t.Version = 2
	t.Locktime = 0

	return &t
}

// NewFromString takes a toBytesHelper string representation of a bitcoin transaction
// and returns a Transaction object.
func NewFromString(str string) (*Transaction, error) {
	bytes, err := hex.DecodeString(str)
	if err != nil {
		return nil, err
	}

	return NewFromBytes(bytes), nil
}

// NewFromBytes takes an array of bytes, constructs a Transaction and returns it.
func NewFromBytes(b []byte) *Transaction {
	if len(b) < 10 {
		return nil // Even an empty transaction has 10 bytes.
	}

	t := Transaction{}

	var offset = 0

	t.Version = binary.LittleEndian.Uint32(b[offset:4])
	offset += 4

	inputCount, size := utils.DecodeVarInt(b[offset:])
	offset += size

	// create inputs
	var i uint64
	for ; i < inputCount; i++ {
		i, size := input.NewFromBytes(b[offset:])
		offset += size

		t.Inputs = append(t.Inputs, i)
	}

	// create outputs
	outputCount, size := utils.DecodeVarInt(b[offset:])
	offset += size
	for i = 0; i < outputCount; i++ {
		o, size := output.NewFromBytes(b[offset:])
		offset += size
		t.Outputs = append(t.Outputs, o)
	}

	t.Locktime = binary.LittleEndian.Uint32(b[offset:])
	offset += 4

	return &t
}

// AddInput adds a new input to the transaction.
func (bt *Transaction) AddInput(input *input.Input) {
	bt.Inputs = append(bt.Inputs, input)
}

// From adds a new input to the transaction from the specified UTXO fields.
func (bt *Transaction) From(txID string, vout uint32, scriptSig string, satoshis uint64) error {
	pts, err := script.NewFromHexString(scriptSig)
	if err != nil {
		return err
	}

	i := &input.Input{
		PreviousTxOutIndex: vout,
		PreviousTxScript:   pts,
		PreviousTxSatoshis: satoshis,
	}

	i.PreviousTxID = txID

	bt.AddInput(i)

	return nil
}

// InputCount returns the number of transaction inputs.
func (bt *Transaction) InputCount() int {
	return len(bt.Inputs)
}

// OutputCount returns the number of transaction inputs.
func (bt *Transaction) OutputCount() int {
	return len(bt.Outputs)
}

// AddOutput adds a new output to the transaction.
func (bt *Transaction) AddOutput(output *output.Output) {

	bt.Outputs = append(bt.Outputs, output)
}

// PayTo function
func (bt *Transaction) PayTo(addr string, satoshis uint64) error {
	o, err := output.NewP2PKHFromAddress(addr, satoshis)
	if err != nil {
		return err
	}

	bt.AddOutput(o)
	return nil
}

// IsCoinbase determines if this transaction is a coinbase by
// checking if the tx input is a standard coinbase input.
func (bt *Transaction) IsCoinbase() bool {
	if len(bt.Inputs) != 1 {
		return false
	}

	if bt.Inputs[0].PreviousTxID != "0000000000000000000000000000000000000000000000000000000000000000" {
		return false
	}

	if bt.Inputs[0].PreviousTxOutIndex == 0xFFFFFFFF || bt.Inputs[0].SequenceNumber == 0xFFFFFFFF {
		return true
	}

	return false
}

// GetInputs returns an array of all inputs in the transaction.
func (bt *Transaction) GetInputs() []*input.Input {
	return bt.Inputs
}

// GetOutputs returns an array of all outputs in the transaction.
func (bt *Transaction) GetOutputs() []*output.Output {
	return bt.Outputs
}

// GetTxID returns the transaction ID of the transaction
// (which is also the transaction hash).
func (bt *Transaction) GetTxID() string {
	return hex.EncodeToString(utils.ReverseBytes(crypto.Sha256d(bt.ToBytes())))
}

// ToHex encodes the transaction into a hex string.
func (bt *Transaction) ToHex() string {
	return hex.EncodeToString(bt.ToBytes())
}

// ToBytes encodes the transaction into a byte array.
// See https://chainquery.com/bitcoin-cli/decoderawtransaction
func (bt *Transaction) ToBytes() []byte {
	return bt.toBytesHelper(0, nil)
}

// ToBytesWithClearedInputs encodes the transaction into a byte array but clears its inputs first.
// This is used when signing transactions.
func (bt *Transaction) ToBytesWithClearedInputs(index int, scriptPubKey []byte) []byte {
	return bt.toBytesHelper(index, scriptPubKey)
}

func (bt *Transaction) toBytesHelper(index int, scriptPubKey []byte) []byte {
	h := make([]byte, 0)

	h = append(h, utils.GetLittleEndianBytes(bt.Version, 4)...)

	h = append(h, utils.VarInt(uint64(len(bt.GetInputs())))...)

	for i, in := range bt.GetInputs() {
		s := in.ToBytes(scriptPubKey != nil)
		if i == index && scriptPubKey != nil {
			h = append(h, utils.VarInt(uint64(len(scriptPubKey)))...)
			h = append(h, scriptPubKey...)
		} else {
			h = append(h, s...)
		}
	}

	h = append(h, utils.VarInt(uint64(len(bt.GetOutputs())))...)
	for _, out := range bt.GetOutputs() {
		h = append(h, out.ToBytes()...)
	}

	lt := make([]byte, 4)
	binary.LittleEndian.PutUint32(lt, bt.Locktime)
	h = append(h, lt...)

	return h
}

// Sign is used to sign the transaction. It takes a Signed interface as a
// parameter so that different signing implementations can be used to sign
// the transaction - for example internal/local or external signing.
func (bt *Transaction) Sign(s Signer) error {
	signedTx, err := s.Sign(bt)
	if err != nil {
		return err
	}
	*bt = *signedTx
	return nil
}

// NewSigningPayloadFromTx creates a new SigningPayload from a Transaction and a SIGHASH type.
func NewSigningPayloadFromTx(bt *Transaction, sigType uint32) (*SigningPayload, error) {
	p := NewSigningPayload()
	for idx, i := range bt.Inputs {
		if i.PreviousTxSatoshis == 0 {
			return nil, errors.New("signing service error - error getting sighashes - Inputs need to have a PreviousTxSatoshis set to be signable")
		}

		if i.PreviousTxScript == nil {
			return nil, errors.New("signing service error - error getting sighashes - Inputs need to have a PreviousScript to be signable")

		}

		sighash := GetSighashForInput(bt, sigType, uint32(idx))
		pkh, _ := i.PreviousTxScript.GetPublicKeyHash() // if not P2PKH, pkh will just be nil
		p.AddItem(hex.EncodeToString(pkh), sighash)     // and the SigningItem will have PublicKeyHash = ""
	}
	return p, nil
}

// ApplySignatures applies the signatures passed in through SigningPayload parameter to the transaction inputs
// The signing payload from the signing service should contain a signing item for each of the tx inputs.
// If the TX input does not belong to us, its signature will be blank unless its owner has already signed it.
// If the signing payload contains a signature for a given input, we apply that to the tx regardless of whether we own it or not.
func (bt *Transaction) ApplySignatures(signingPayload *SigningPayload, sigType uint32) error {
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

// GetSighashPayload assembles a payload of sighases for this TX, to be submitted to signing service.
func (bt *Transaction) GetSighashPayload(sigType uint32) (*SigningPayload, error) {
	signingPayload, err := NewSigningPayloadFromTx(bt, sigType)
	if err != nil {
		return nil, err
	}
	return signingPayload, nil
}

// ApplySignaturesWithoutP2PKHCheck applies signatures without checking if the input previous script equals
// to a P2PKH script matching the private key (see func SignWithoutP2PKHCheck below)
func (bt *Transaction) ApplySignaturesWithoutP2PKHCheck(signingPayload *SigningPayload, sigType uint32) error {
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
func (bt *Transaction) SignWithoutP2PKHCheck(privateKey *bsvec.PrivateKey, sigType uint32) error {
	if sigType == 0 {
		sigType = SighashAllForkID
	}

	payload, err := bt.GetSighashPayload(sigType)
	if err != nil {
		return err
	}

	// todo abstract into separate function. used in InternalSigner.
	// loops through signing items for each input and signs accordingly
	for _, signingItem := range *payload {
		h, err := hex.DecodeString(signingItem.SigHash)
		if err != nil {
			return err
		}
		sig, err := privateKey.Sign(utils.ReverseBytes(h))
		if err != nil {
			return err
		}
		pubkey := privateKey.PubKey().SerializeCompressed()
		signingItem.PublicKey = hex.EncodeToString(pubkey)
		signingItem.Signature = hex.EncodeToString(sig.Serialize())
	}

	err = bt.ApplySignaturesWithoutP2PKHCheck(payload, sigType)
	if err != nil {
		return err
	}
	return nil
}
