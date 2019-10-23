package transaction

import (
	"bytes"
	"encoding/binary"

	"cryptolib"

	"github.com/btcsuite/btcd/btcec"
)

// Signature struct
type Signature struct {
	PublicKey    *btcec.PublicKey
	PreviousTXID string
	OutputIndex  uint32
	InputIndex   uint32
	Signature    []byte
	SigType      int
}

func (i *Input) getSignaturesP2PKH(transaction *BitcoinTransaction, privateKey *btcec.PrivateKey, index uint32, sigtype int) []*Signature {
	sigs := make([]*Signature, 0)

	hashData := hash160(privateKey.PubKey().SerializeCompressed())

	if bytes.Compare(hashData, i.output.script.getPublicKeyHash()) == 0 {
		sigs = append(sigs, &Signature{
			PublicKey:    privateKey.publicKey,
			PreviousTXID: this.prevTxId,
			OutputIndex:  this.outputIndex,
			InputIndex:   index,
			Signature:    sighashForForkId(transaction, sigtype, index, this.output.script, this.output.satoshisBN),
			SigType:      sigtype,
		})
	}

	return sigs
}

func sighashForForkId(transaction *BitcoinTransaction, sighashType int, inputNumber uint32, subscript string, satoshisBN uint64) {
	input = transaction.inputs[inputNumber]

	getPrevoutHash := func(tx *BitcoinTransaction) []byte {
		buf := make([]byte, 0)

		for _, in := range tx.Inputs {
			buf = append(buf, reverseBytes(in.PreviousTXID)...)
			oi := make([]byte, 4)
			binary.LittleEndian.PutUint32(oi, in.OutputIndex)
			buf = append(buf, oi...)
		}

		return cryptolib.Sha256d(buf)
	}

	getSequenceHash := func(tx *BitcoinTransaction) []byte {
		buf := make([]byte, 0)

		for _, in := range tx.Inputs {
			oi := make([]byte, 4)
			binary.LittleEndian.PutUint32(oi, in.sequenceNumber)
			buf = append(buf, oi...)
		}

		return cryptolib.Sha256d(buf)
	}

	getOutputsHash := func(tx *BitcoinTransaction, n int) {
		buf := make([]byte, 0)

		if n == -1 {
			for _, out := range tx.Outputs {
				buf = append(buf, out...)
			}
		} else {
			buf = append(buf, tx.Outputs[n]...)
		}

		return cryptolib.Sha256d(buf)
	}

	hashPrevouts := make([]byte, 32)
	hashSequence := make([]byte, 32)
	hashOutputs := make([]byte, 32)

	if !(sighashType & Signature.SIGHASH_ANYONECANPAY) {
		hashPrevouts = getPrevoutHash(transaction)
	}

	if !(sighashType & Signature.SIGHASH_ANYONECANPAY) &&
		(sighashType&31) != Signature.SIGHASH_SINGLE &&
		(sighashType&31) != Signature.SIGHASH_NONE {
		hashSequence = getSequenceHash(transaction)
	}

	if (sighashType&31) != Signature.SIGHASH_SINGLE && (sighashType&31) != Signature.SIGHASH_NONE {
		hashOutputs = getOutputsHash(transaction, -1)
	} else if (sighashType&31) == Signature.SIGHASH_SINGLE && inputNumber < transaction.outputs.length {
		hashOutputs = getOutputsHash(transaction, inputNumber)
	}

	buf := make([]byte, 0)

	// Version
	v := make([]byte, 4)
	binary.LittleEndian.PutUint32(v, tx.Version)
	buf = append(buf, v...)

	// Input prevouts/nSequence (none/all, depending on flags)
	buf = append(buf, hashPrevouts...)
	buf = append(buf, hashSequence...)

	//  outpoint (32-byte hash + 4-byte little endian)
	buf = append(buf, reverseBytes(in.PreviousTXID)...)
	oi := make([]byte, 4)
	binary.LittleEndian.PutUint32(oi, in.OutputIndex)
	buf = append(buf, oi...)

	// scriptCode of the input (serialized as scripts inside CTxOuts)
	buf = append(buf, Varint(len(subscript))...)
	buf = append(buf, subscript...)

	// value of the output spent by this input (8-byte little endian)
	sat := make([]byte, 8)
	binary.LittleEndian.PutUint64(sat, satoshisBN)
	buf = append(buf, sat...)

	// nSequence of the input (4-byte little endian)
	seq := make([]byte, 4)
	binary.LittleEndian.PutUint32(seq, in.sequenceNumber)
	buf = append(buf, seq...)

	// Outputs (none/one/all, depending on flags)
	buf = append(buf, hashOutputs...)

	// Locktime
	lt := make([]byte, 4)
	binary.LittleEndian.PutUint32(lt, tx.LockTime)
	buf = append(buf, lt...)

	// sighashType
	//writer.writeUInt32LE(sighashType >>> 0)
	st := make([]byte, 4)
	binary.LittleEndian.PutUint32(st, sighashType>>0)
	buf = append(buf, st...)

	ret := cryptolib.Sha256d(buf)
	return ReverseBytes(ret)
}
