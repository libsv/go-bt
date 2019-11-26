package transaction

import (
	"encoding/binary"
	"encoding/hex"
	"errors"

	"bitbucket.org/simon_ordish/cryptolib"
)

// SigningItem struct
type SigningItem struct {
	PublicKeyHash string `json:"publicKeyHash"`
	SigHash       string `json:"sigHash"`
	PublicKey     string `json:"publicKey,omitempty"`
	Signature     string `json:"signature,omitempty"`
}

// SigningPayload type
type SigningPayload []*SigningItem

// NewSigningPayload comment
func NewSigningPayload() *SigningPayload {
	sp := make([]*SigningItem, 0)
	p := SigningPayload(sp)
	return &p
}

// NewSigningPayloadFromTx create a signinbg payload for a transaction.
func NewSigningPayloadFromTx(bt *BitcoinTransaction) (*SigningPayload, error) {
	p := NewSigningPayload()
	for idx, input := range bt.Inputs {
		if input.PreviousTxSatoshis == 0 {
			return nil, errors.New("Error getting sighashes - Inputs need to have a PreviousTxSatoshis set to be signable")
		}

		if input.PreviousTxScript == nil {
			return nil, errors.New("Error getting sighashes - Inputs need to have a PreviousScript to be signable")

		}

		sighash := getSighash(bt, SighashAllForkID, uint32(idx), *input.PreviousTxScript, input.PreviousTxSatoshis)
		pkh, err := input.PreviousTxScript.GetPublicKeyHash()
		if err != nil {
			return nil, err
		}
		p.AddItem(hex.EncodeToString(pkh), sighash)
	}
	return p, nil
}

// AddItem function
func (sp *SigningPayload) AddItem(publicKeyHash string, sigHash string) {
	si := &SigningItem{
		PublicKeyHash: publicKeyHash,
		SigHash:       sigHash,
	}

	*sp = append(*sp, si)
}

func getSighash(transaction *BitcoinTransaction, sighashType uint32, inputNumber uint32, subscript Script, satoshis uint64) string {

	input := transaction.Inputs[inputNumber]

	getPrevoutHash := func(tx *BitcoinTransaction) []byte {
		buf := make([]byte, 0)

		for _, in := range tx.Inputs {
			buf = append(buf, cryptolib.ReverseBytes(in.PreviousTxHash[:])...)
			oi := make([]byte, 4)
			binary.LittleEndian.PutUint32(oi, in.PreviousTxOutIndex)
			buf = append(buf, oi...)
		}

		return cryptolib.Sha256d(buf)
	}

	getSequenceHash := func(tx *BitcoinTransaction) []byte {
		buf := make([]byte, 0)

		for _, in := range tx.Inputs {
			oi := make([]byte, 4)
			binary.LittleEndian.PutUint32(oi, in.SequenceNumber)
			buf = append(buf, oi...)
		}

		return cryptolib.Sha256d(buf)
	}

	getOutputsHash := func(tx *BitcoinTransaction, n int32) []byte {
		buf := make([]byte, 0)

		if n == -1 {
			for _, out := range tx.Outputs {
				buf = append(buf, out.getBytesForSigHash()...)
			}
		} else {
			buf = append(buf, tx.Outputs[n].getBytesForSigHash()...)
		}

		return cryptolib.Sha256d(buf)
	}

	hashPrevouts := make([]byte, 32)
	hashSequence := make([]byte, 32)
	hashOutputs := make([]byte, 32)

	if sighashType&SighashAnyoneCanPay == 0 {
		hashPrevouts = getPrevoutHash(transaction)
	}

	if sighashType&SighashAnyoneCanPay == 0 &&
		(sighashType&31) != SighashSingle &&
		(sighashType&31) != SighashNone {
		hashSequence = getSequenceHash(transaction)
	}

	if (sighashType&31) != SighashSingle && (sighashType&31) != SighashNone {
		hashOutputs = getOutputsHash(transaction, -1)
	} else if (sighashType&31) == SighashSingle && inputNumber < uint32(len(transaction.Outputs)) {
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
	buf = append(buf, cryptolib.ReverseBytes(input.PreviousTxHash[:])...)
	oi := make([]byte, 4)
	binary.LittleEndian.PutUint32(oi, input.PreviousTxOutIndex)
	buf = append(buf, oi...)

	// scriptCode of the input (serialized as scripts inside CTxOuts)
	buf = append(buf, cryptolib.VarInt(uint64(len(subscript)))...)
	buf = append(buf, subscript...)

	// value of the output spent by this input (8-byte little endian)
	sat := make([]byte, 8)
	binary.LittleEndian.PutUint64(sat, satoshis)
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
	//writer.writeUInt32LE(sighashType >>> 0)
	st := make([]byte, 4)
	binary.LittleEndian.PutUint32(st, sighashType>>0)
	buf = append(buf, st...)
	ret := cryptolib.Sha256d(buf)
	return hex.EncodeToString(cryptolib.ReverseBytes(ret))
}
