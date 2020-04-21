package transaction

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"github.com/libsv/libsv/crypto"
	"github.com/libsv/libsv/script"
	"github.com/libsv/libsv/utils"
)

// SigningItem contains the metadata needed to sign a transaction.
type SigningItem struct {
	PublicKeyHash string `json:"publicKeyHash"`
	SigHash       string `json:"sigHash"`
	PublicKey     string `json:"publicKey,omitempty"`
	Signature     string `json:"signature,omitempty"`
}

// SigningPayload type
type SigningPayload []*SigningItem

// NewSigningPayload creates a new SigningPayload.
func NewSigningPayload() *SigningPayload {
	sp := make([]*SigningItem, 0)
	p := SigningPayload(sp)
	return &p
}

// NewSigningPayloadFromTx creates a new SigningPayload from a BitcoinTransaction and a SIGHASH type.
func NewSigningPayloadFromTx(bt *BitcoinTransaction, sigType uint32) (*SigningPayload, error) {
	p := NewSigningPayload()
	for idx, input := range bt.Inputs {
		if input.PreviousTxSatoshis == 0 {
			return nil, errors.New("signing service error - error getting sighashes - Inputs need to have a PreviousTxSatoshis set to be signable")
		}

		if input.PreviousTxScript == nil {
			return nil, errors.New("signing service error - error getting sighashes - Inputs need to have a PreviousScript to be signable")

		}

		sighash := GetSighashForInput(bt, sigType, uint32(idx))
		pkh, _ := input.PreviousTxScript.GetPublicKeyHash() // if not P2PKH, pkh will just be nil
		p.AddItem(hex.EncodeToString(pkh), sighash)         // and the SigningItem will have PublicKeyHash = ""
	}
	return p, nil
}

// AddItem appends a new SigningItem to the SigningPayload array.
func (sp *SigningPayload) AddItem(publicKeyHash string, sigHash string) {
	si := &SigningItem{
		PublicKeyHash: publicKeyHash,
		SigHash:       sigHash,
	}

	*sp = append(*sp, si)
}

// GetSighashForInput function
func GetSighashForInput(transaction *BitcoinTransaction, sighashType uint32, inputNumber uint32) string {

	input := transaction.Inputs[inputNumber]

	getPrevoutHash := func(tx *BitcoinTransaction) []byte {
		buf := make([]byte, 0)

		for _, in := range tx.Inputs {
			buf = append(buf, utils.ReverseBytes(in.PreviousTxHash[:])...)
			oi := make([]byte, 4)
			binary.LittleEndian.PutUint32(oi, in.PreviousTxOutIndex)
			buf = append(buf, oi...)
		}

		return crypto.Sha256d(buf)
	}

	getSequenceHash := func(tx *BitcoinTransaction) []byte {
		buf := make([]byte, 0)

		for _, in := range tx.Inputs {
			oi := make([]byte, 4)
			binary.LittleEndian.PutUint32(oi, in.SequenceNumber)
			buf = append(buf, oi...)
		}

		return crypto.Sha256d(buf)
	}

	getOutputsHash := func(tx *BitcoinTransaction, n int32) []byte {
		buf := make([]byte, 0)

		if n == -1 {
			for _, out := range tx.Outputs {
				buf = append(buf, out.GetBytesForSigHash()...)
			}
		} else {
			buf = append(buf, tx.Outputs[n].GetBytesForSigHash()...)
		}

		return crypto.Sha256d(buf)
	}

	hashPrevouts := make([]byte, 32)
	hashSequence := make([]byte, 32)
	hashOutputs := make([]byte, 32)

	if sighashType&SighashAnyoneCanPay == 0 {
		// This will be executed in the usual BSV case (where sighashType = SighashAllForkID)
		hashPrevouts = getPrevoutHash(transaction)
	}

	if sighashType&SighashAnyoneCanPay == 0 &&
		(sighashType&31) != SighashSingle &&
		(sighashType&31) != SighashNone {
		// This will be executed in the usual BSV case (where sighashType = SighashAllForkID)
		hashSequence = getSequenceHash(transaction)
	}

	if (sighashType&31) != SighashSingle && (sighashType&31) != SighashNone {
		// This will be executed in the usual BSV case (where sighashType = SighashAllForkID)
		hashOutputs = getOutputsHash(transaction, -1)
	} else if (sighashType&31) == SighashSingle && inputNumber < uint32(len(transaction.Outputs)) {
		// This will *not* be executed in the usual BSV case (where sighashType = SighashAllForkID)
		hashOutputs = getOutputsHash(transaction, int32(inputNumber))
	}

	buf := make([]byte, 0)

	// Version
	v := make([]byte, 4)
	binary.LittleEndian.PutUint32(v, transaction.Version)
	buf = append(buf, v...)

	// Input prevouts/nSequence (none/all, depending on flags)
	buf = append(buf, hashPrevouts...)
	buf = append(buf, hashSequence...)

	//  outpoint (32-byte hash + 4-byte little endian)
	buf = append(buf, utils.ReverseBytes(input.PreviousTxHash[:])...)
	oi := make([]byte, 4)
	binary.LittleEndian.PutUint32(oi, input.PreviousTxOutIndex)
	buf = append(buf, oi...)

	// scriptCode of the input (serialized as scripts inside CTxOuts)
	buf = append(buf, utils.VarInt(uint64(len(*input.PreviousTxScript)))...)
	buf = append(buf, *input.PreviousTxScript...)

	// value of the output spent by this input (8-byte little endian)
	sat := make([]byte, 8)
	binary.LittleEndian.PutUint64(sat, input.PreviousTxSatoshis)
	buf = append(buf, sat...)

	// nSequence of the input (4-byte little endian)
	seq := make([]byte, 4)
	binary.LittleEndian.PutUint32(seq, input.SequenceNumber)
	buf = append(buf, seq...)

	// Outputs (none/one/all, depending on flags)
	buf = append(buf, hashOutputs...)

	// Locktime
	lt := make([]byte, 4)
	binary.LittleEndian.PutUint32(lt, transaction.Locktime)
	buf = append(buf, lt...)

	// sighashType
	// writer.writeUInt32LE(sighashType >>> 0)
	st := make([]byte, 4)
	binary.LittleEndian.PutUint32(st, sighashType>>0)
	buf = append(buf, st...)
	ret := crypto.Sha256d(buf)
	return hex.EncodeToString(utils.ReverseBytes(ret))
}

// GetSighashForInputValidation comment todo
func GetSighashForInputValidation(transaction *BitcoinTransaction, sighashType uint32, inputNumber uint32, previousTxOutIndex uint32, previousTxSatoshis uint64, previousTxScript *script.Script) string {

	input := transaction.Inputs[inputNumber]

	getPrevoutHash := func(tx *BitcoinTransaction) []byte {
		buf := make([]byte, 0)

		for _, in := range tx.Inputs {
			buf = append(buf, utils.ReverseBytes(in.PreviousTxHash[:])...)
			oi := make([]byte, 4)
			binary.LittleEndian.PutUint32(oi, previousTxOutIndex)
			buf = append(buf, oi...)
		}

		return crypto.Sha256d(buf)
	}

	getSequenceHash := func(tx *BitcoinTransaction) []byte {
		buf := make([]byte, 0)

		for _, in := range tx.Inputs {
			oi := make([]byte, 4)
			binary.LittleEndian.PutUint32(oi, in.SequenceNumber)
			buf = append(buf, oi...)
		}

		return crypto.Sha256d(buf)
	}

	getOutputsHash := func(tx *BitcoinTransaction, n int32) []byte {
		buf := make([]byte, 0)

		if n == -1 {
			for _, out := range tx.Outputs {
				buf = append(buf, out.GetBytesForSigHash()...)
			}
		} else {
			buf = append(buf, tx.Outputs[n].GetBytesForSigHash()...)
		}

		return crypto.Sha256d(buf)
	}

	hashPrevouts := make([]byte, 32)
	hashSequence := make([]byte, 32)
	hashOutputs := make([]byte, 32)

	if sighashType&SighashAnyoneCanPay == 0 {
		// This will be executed in the usual BSV case (where sighashType = SighashAllForkID)
		hashPrevouts = getPrevoutHash(transaction)
	}

	if sighashType&SighashAnyoneCanPay == 0 &&
		(sighashType&31) != SighashSingle &&
		(sighashType&31) != SighashNone {
		// This will be executed in the usual BSV case (where sighashType = SighashAllForkID)
		hashSequence = getSequenceHash(transaction)
	}

	if (sighashType&31) != SighashSingle && (sighashType&31) != SighashNone {
		// This will be executed in the usual BSV case (where sighashType = SighashAllForkID)
		hashOutputs = getOutputsHash(transaction, -1)
	} else if (sighashType&31) == SighashSingle && inputNumber < uint32(len(transaction.Outputs)) {
		// This will *not* be executed in the usual BSV case (where sighashType = SighashAllForkID)
		hashOutputs = getOutputsHash(transaction, int32(inputNumber))
	}

	buf := make([]byte, 0)

	// Version
	v := make([]byte, 4)
	binary.LittleEndian.PutUint32(v, transaction.Version)
	buf = append(buf, v...)

	// Input prevouts/nSequence (none/all, depending on flags)
	buf = append(buf, hashPrevouts...)
	buf = append(buf, hashSequence...)

	//  outpoint (32-byte hash + 4-byte little endian)
	buf = append(buf, utils.ReverseBytes(input.PreviousTxHash[:])...)
	oi := make([]byte, 4)
	binary.LittleEndian.PutUint32(oi, previousTxOutIndex)
	buf = append(buf, oi...)

	// scriptCode of the input (serialized as scripts inside CTxOuts)
	buf = append(buf, utils.VarInt(uint64(len(*previousTxScript)))...)
	buf = append(buf, *previousTxScript...)

	// value of the output spent by this input (8-byte little endian)
	sat := make([]byte, 8)
	binary.LittleEndian.PutUint64(sat, previousTxSatoshis)
	buf = append(buf, sat...)

	// nSequence of the input (4-byte little endian)
	seq := make([]byte, 4)
	binary.LittleEndian.PutUint32(seq, input.SequenceNumber)
	buf = append(buf, seq...)

	// Outputs (none/one/all, depending on flags)
	buf = append(buf, hashOutputs...)

	// Locktime
	lt := make([]byte, 4)
	binary.LittleEndian.PutUint32(lt, transaction.Locktime)
	buf = append(buf, lt...)

	// sighashType
	// writer.writeUInt32LE(sighashType >>> 0)
	st := make([]byte, 4)
	binary.LittleEndian.PutUint32(st, sighashType>>0)
	buf = append(buf, st...)
	ret := crypto.Sha256d(buf)
	return hex.EncodeToString(utils.ReverseBytes(ret))
}
