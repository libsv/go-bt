package bt

import (
	"encoding/binary"
	"errors"

	"github.com/libsv/go-bk/crypto"
	"github.com/libsv/go-bt/v2/sighash"
)

// CalcInputSignatureHash serialised the transaction and returns the hash digest
// to be signed. BitCoin (SV) uses a different signature hashing algorithm
// after the UAHF fork for replay protection.
//
// see https://github.com/bitcoin-sv/bitcoin-sv/blob/master/doc/abc/replay-protected-sighash.md#digest-algorithm
func (tx *Tx) CalcInputSignatureHash(inputNumber uint32, sigHashFlag sighash.Flag) ([]byte, error) {
	buf, err := tx.CalcInputPreimage(inputNumber, sigHashFlag)
	if err != nil {
		return nil, err
	}

	return crypto.Sha256d(buf), nil
}

// CalcInputPreimage serialises the transaction based on the input index and the SIGHASH flag
// and returns the preimage before double hashing (SHA256d).
//
// see https://github.com/bitcoin-sv/bitcoin-sv/blob/master/doc/abc/replay-protected-sighash.md#digest-algorithm
func (tx *Tx) CalcInputPreimage(inputNumber uint32, sigHashFlag sighash.Flag) ([]byte, error) {

	if tx.InputIdx(int(inputNumber)) == nil {
		return nil, errors.New("specified input does not exist")
	}
	in := tx.InputIdx(int(inputNumber))

	if len(in.PreviousTxID()) == 0 {
		return nil, errors.New("'PreviousTxID' not supplied")
	}
	if in.PreviousTxScript == nil {
		return nil, errors.New("'PreviousTxScript' not supplied")
	}

	hashPreviousOuts := make([]byte, 32)
	hashSequence := make([]byte, 32)
	hashOutputs := make([]byte, 32)

	if sigHashFlag&sighash.AnyOneCanPay == 0 {
		// This will be executed in the usual BSV case (where sigHashType = SighashAllForkID)
		hashPreviousOuts = tx.getPreviousOutHash()
	}

	if sigHashFlag&sighash.AnyOneCanPay == 0 &&
		(sigHashFlag&31) != sighash.Single &&
		(sigHashFlag&31) != sighash.None {
		// This will be executed in the usual BSV case (where sigHashType = SighashAllForkID)
		hashSequence = tx.getSequenceHash()
	}

	if (sigHashFlag&31) != sighash.Single && (sigHashFlag&31) != sighash.None {
		// This will be executed in the usual BSV case (where sigHashType = SighashAllForkID)
		hashOutputs = tx.getOutputsHash(-1)
	} else if (sigHashFlag&31) == sighash.Single && inputNumber < uint32(tx.OutputCount()) {
		// This will *not* be executed in the usual BSV case (where sigHashType = SighashAllForkID)
		hashOutputs = tx.getOutputsHash(int32(inputNumber))
	}

	buf := make([]byte, 0)

	// Version
	v := make([]byte, 4)
	binary.LittleEndian.PutUint32(v, tx.Version)
	buf = append(buf, v...)

	// Input previousOuts/nSequence (none/all, depending on flags)
	buf = append(buf, hashPreviousOuts...)
	buf = append(buf, hashSequence...)

	//  outpoint (32-byte hash + 4-byte little endian)
	buf = append(buf, ReverseBytes(in.PreviousTxID())...)
	oi := make([]byte, 4)
	binary.LittleEndian.PutUint32(oi, in.PreviousTxOutIndex)
	buf = append(buf, oi...)

	// scriptCode of the input (serialised as scripts inside CTxOuts)
	buf = append(buf, VarInt(uint64(len(*in.PreviousTxScript)))...)
	buf = append(buf, *in.PreviousTxScript...)

	// value of the output spent by this input (8-byte little endian)
	sat := make([]byte, 8)
	binary.LittleEndian.PutUint64(sat, in.PreviousTxSatoshis)
	buf = append(buf, sat...)

	// nSequence of the input (4-byte little endian)
	seq := make([]byte, 4)
	binary.LittleEndian.PutUint32(seq, in.SequenceNumber)
	buf = append(buf, seq...)

	// Outputs (none/one/all, depending on flags)
	buf = append(buf, hashOutputs...)

	// LockTime
	lt := make([]byte, 4)
	binary.LittleEndian.PutUint32(lt, tx.LockTime)
	buf = append(buf, lt...)

	// sighashType
	// writer.writeUInt32LE(sighashType >>> 0)
	st := make([]byte, 4)
	binary.LittleEndian.PutUint32(st, uint32(sigHashFlag)>>0)
	buf = append(buf, st...)

	return buf, nil
}

func (tx *Tx) getPreviousOutHash() []byte {
	buf := make([]byte, 0)

	for _, in := range tx.Inputs {
		buf = append(buf, ReverseBytes(in.PreviousTxID())...)
		oi := make([]byte, 4)
		binary.LittleEndian.PutUint32(oi, in.PreviousTxOutIndex)
		buf = append(buf, oi...)
	}

	return crypto.Sha256d(buf)
}

func (tx *Tx) getSequenceHash() []byte {
	buf := make([]byte, 0)

	for _, in := range tx.Inputs {
		oi := make([]byte, 4)
		binary.LittleEndian.PutUint32(oi, in.SequenceNumber)
		buf = append(buf, oi...)
	}

	return crypto.Sha256d(buf)
}

func (tx *Tx) getOutputsHash(n int32) []byte {
	buf := make([]byte, 0)

	if n == -1 {
		for _, out := range tx.Outputs {
			buf = append(buf, out.BytesForSigHash()...)
		}
	} else {
		buf = append(buf, tx.Outputs[n].BytesForSigHash()...)
	}

	return crypto.Sha256d(buf)
}
