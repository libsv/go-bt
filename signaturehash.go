package bt

import (
	"bytes"
	"encoding/binary"

	"github.com/libsv/go-bk/crypto"
	"github.com/libsv/go-bt/v2/bscript"
	"github.com/libsv/go-bt/v2/sighash"
)

// defaultHex is used to fix a bug in the original client (see if statement in the CalcInputSignatureHash func)
var defaultHex = []byte{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

type sigHashFunc func(inputIdx uint32, shf sighash.Flag) ([]byte, error)

// sigStrat will decide which tx serialisation to use.
// The legacy serialisation will be used for txs pre-fork
// whereas the new serialisation will be used for post-fork
// txs (and they should include the sighash_forkid flag).
func (tx *Tx) sigStrat(shf sighash.Flag) sigHashFunc {
	if shf.Has(sighash.ForkID) {
		return tx.CalcInputPreimage
	}
	return tx.CalcInputPreimageLegacy
}

// CalcInputSignatureHash serialised the transaction and returns the hash digest
// to be signed. BitCoin (SV) uses a different signature hashing algorithm
// after the UAHF fork for replay protection.
//
// see https://github.com/bitcoin-sv/bitcoin-sv/blob/master/doc/abc/replay-protected-sighash.md#digest-algorithm
func (tx *Tx) CalcInputSignatureHash(inputNumber uint32, sigHashFlag sighash.Flag) ([]byte, error) {
	sigHashFn := tx.sigStrat(sigHashFlag)
	buf, err := sigHashFn(inputNumber, sigHashFlag)
	if err != nil {
		return nil, err
	}

	// A bug in the original Satoshi client implementation means specifying
	// an index that is out of range results in a signature hash of 1 (as an
	// uint256 little endian).  The original intent appeared to be to
	// indicate failure, but unfortunately, it was never checked and thus is
	// treated as the actual signature hash.  This buggy behaviour is now
	// part of the consensus and a hard fork would be required to fix it.
	//
	// Due to this, if the tx signature returned matches this special case value,
	// we skip the double hashing as to not interfere.
	if bytes.Equal(defaultHex, buf) {
		return buf, nil
	}

	return crypto.Sha256d(buf), nil
}

// CalcInputPreimage serialises the transaction based on the input index and the SIGHASH flag
// and returns the preimage before double hashing (SHA256d).
//
// see https://github.com/bitcoin-sv/bitcoin-sv/blob/master/doc/abc/replay-protected-sighash.md#digest-algorithm
func (tx *Tx) CalcInputPreimage(inputNumber uint32, sigHashFlag sighash.Flag) ([]byte, error) {
	if tx.InputIdx(int(inputNumber)) == nil {
		return nil, ErrInputNoExist
	}
	in := tx.InputIdx(int(inputNumber))

	if len(in.PreviousTxID()) == 0 {
		return nil, ErrEmptyPreviousTxID
	}
	if in.PreviousTxScript == nil {
		return nil, ErrEmptyPreviousTxScript
	}

	hashPreviousOuts := make([]byte, 32)
	hashSequence := make([]byte, 32)
	hashOutputs := make([]byte, 32)

	if sigHashFlag&sighash.AnyOneCanPay == 0 {
		// This will be executed in the usual BSV case (where sigHashType = SighashAllForkID)
		hashPreviousOuts = tx.PreviousOutHash()
	}

	if sigHashFlag&sighash.AnyOneCanPay == 0 &&
		(sigHashFlag&31) != sighash.Single &&
		(sigHashFlag&31) != sighash.None {
		// This will be executed in the usual BSV case (where sigHashType = SighashAllForkID)
		hashSequence = tx.SequenceHash()
	}

	if (sigHashFlag&31) != sighash.Single && (sigHashFlag&31) != sighash.None {
		// This will be executed in the usual BSV case (where sigHashType = SighashAllForkID)
		hashOutputs = tx.OutputsHash(-1)
	} else if (sigHashFlag&31) == sighash.Single && inputNumber < uint32(tx.OutputCount()) {
		// This will *not* be executed in the usual BSV case (where sigHashType = SighashAllForkID)
		hashOutputs = tx.OutputsHash(int32(inputNumber))
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
	buf = append(buf, VarInt(uint64(len(*in.PreviousTxScript))).Bytes()...)
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

// CalcInputPreimageLegacy serialises the transaction based on the input index and the SIGHASH flag
// and returns the preimage before double hashing (SHA256d), in the legacy format.
//
// see https://wiki.bitcoinsv.io/index.php/Legacy_Sighash_Algorithm
func (tx *Tx) CalcInputPreimageLegacy(inputNumber uint32, shf sighash.Flag) ([]byte, error) {
	if tx.InputIdx(int(inputNumber)) == nil {
		return nil, ErrInputNoExist
	}
	in := tx.InputIdx(int(inputNumber))

	if len(in.PreviousTxID()) == 0 {
		return nil, ErrEmptyPreviousTxID
	}
	if in.PreviousTxScript == nil {
		return nil, ErrEmptyPreviousTxScript
	}

	// The SigHashSingle signature type signs only the corresponding input
	// and output (the output with the same index number as the input).
	//
	// Since transactions can have more inputs than outputs, this means it
	// is improper to use SigHashSingle on input indices that don't have a
	// corresponding output.
	//
	// A bug in the original Satoshi client implementation means specifying
	// an index that is out of range results in a signature hash of 1 (as an
	// uint256 little endian).  The original intent appeared to be to
	// indicate failure, but unfortunately, it was never checked and thus is
	// treated as the actual signature hash.  This buggy behaviour is now
	// part of the consensus and a hard fork would be required to fix it.
	//
	// Due to this, care must be taken by software that creates transactions
	// which make use of SigHashSingle because it can lead to an extremely
	// dangerous situation where the invalid inputs will end up signing a
	// hash of 1.  This in turn presents an opportunity for attackers to
	// cleverly construct transactions which can steal those coins provided
	// they can reuse signatures.
	if shf.HasWithMask(sighash.Single) && int(inputNumber) > len(tx.Outputs)-1 {
		return defaultHex, nil
	}

	txCopy := tx.Clone()

	for i := range txCopy.Inputs {
		if i == int(inputNumber) {
			txCopy.Inputs[i].PreviousTxScript = tx.Inputs[inputNumber].PreviousTxScript
		} else {
			txCopy.Inputs[i].UnlockingScript = &bscript.Script{}
			txCopy.Inputs[i].PreviousTxScript = &bscript.Script{}
		}
	}

	if shf.HasWithMask(sighash.None) {
		txCopy.Outputs = txCopy.Outputs[0:0]
		for i := range txCopy.Inputs {
			if i != int(inputNumber) {
				txCopy.Inputs[i].SequenceNumber = 0
			}
		}
	} else if shf.HasWithMask(sighash.Single) {
		txCopy.Outputs = txCopy.Outputs[:inputNumber+1]
		for i := 0; i < int(inputNumber); i++ {
			txCopy.Outputs[i].Satoshis = 18446744073709551615 // -1 but underflowed
			txCopy.Outputs[i].LockingScript = &bscript.Script{}
		}

		for i := range txCopy.Inputs {
			if i != int(inputNumber) {
				txCopy.Inputs[i].SequenceNumber = 0
			}
		}
	}

	if shf&sighash.AnyOneCanPay != 0 {
		txCopy.Inputs = txCopy.Inputs[inputNumber : inputNumber+1]
	}

	buf := make([]byte, 0)

	// Version
	v := make([]byte, 4)
	binary.LittleEndian.PutUint32(v, tx.Version)
	buf = append(buf, v...)

	buf = append(buf, VarInt(uint64(len(txCopy.Inputs))).Bytes()...)
	for _, in := range txCopy.Inputs {
		buf = append(buf, ReverseBytes(in.PreviousTxID())...)

		oi := make([]byte, 4)
		binary.LittleEndian.PutUint32(oi, in.PreviousTxOutIndex)
		buf = append(buf, oi...)

		buf = append(buf, VarInt(uint64(len(*in.PreviousTxScript))).Bytes()...)
		buf = append(buf, *in.PreviousTxScript...)

		sq := make([]byte, 4)
		binary.LittleEndian.PutUint32(sq, in.SequenceNumber)
		buf = append(buf, sq...)
	}

	buf = append(buf, VarInt(uint64(len(txCopy.Outputs))).Bytes()...)
	for _, out := range txCopy.Outputs {
		st := make([]byte, 8)
		binary.LittleEndian.PutUint64(st, out.Satoshis)
		buf = append(buf, st...)

		buf = append(buf, VarInt(uint64(len(*out.LockingScript))).Bytes()...)
		buf = append(buf, *out.LockingScript...)
	}

	// LockTime
	lt := make([]byte, 4)
	binary.LittleEndian.PutUint32(lt, tx.LockTime)
	buf = append(buf, lt...)

	sh := make([]byte, 4)
	binary.LittleEndian.PutUint32(sh, uint32(shf)>>0)
	return append(buf, sh...), nil
}

// OutputsHash returns a bytes slice of the requested output, used for generating
// the txs signature hash. If n is -1, it will create the byte slice from all outputs.
func (tx *Tx) OutputsHash(n int32) []byte {
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
