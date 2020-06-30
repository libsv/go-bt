package transaction

import (
	"encoding/binary"
	"encoding/hex"
	"errors"

	"github.com/libsv/libsv/crypto"
	"github.com/libsv/libsv/script"
	"github.com/libsv/libsv/utils"
)

// SigHashType represents hash type bits at the end of a signature.
type SigHashType uint32

// SIGHASH type bits from the end of a signature.
// see: https://wiki.bitcoinsv.io/index.php/SIGHASH_flags
const (
	SigHashOld          SigHashType = 0x0
	SigHashAll          SigHashType = 0x1
	SigHashNone         SigHashType = 0x2
	SigHashSingle       SigHashType = 0x3
	SigHashAnyOneCanPay SigHashType = 0x80

	// Currently all BitCoin (SV) transactions require an additional SIGHASH flag (after UAHF)
	SigHashAllForkID          SigHashType = 0x1 | 0x00000040
	SigHashNoneForkID         SigHashType = 0x2 | 0x00000040
	SigHashSingleForkID       SigHashType = 0x3 | 0x00000040
	SigHashAnyOneCanPayForkID SigHashType = 0x80 | 0x00000040

	// SigHashForkID is the replay protected signature hash flag
	// used by the Uahf hardfork.
	SigHashForkID SigHashType = 0x40

	// sigHashMask defines the number of bits of the hash type which is used
	// to identify which outputs are signed.
	sigHashMask = 0x1f
)

// GetInputSignatureHash serializes the transaction based on the sig TODO:
func (bt *Transaction) GetInputSignatureHash(inputNumber uint32, sigHashType SigHashType) ([]byte, error) {
	in := bt.Inputs[inputNumber]

	if bt.IsCoinbase() {
		bt.Inputs[inputNumber].PreviousTxScript = &script.Script{}

	} else {
		if in.PreviousTxID == "" {
			return nil, errors.New("'PreviousTxID' not supplied")
		}
		if in.PreviousTxScript == nil {
			return nil, errors.New("'PreviousTxScript' not supplied")
		}
	}

	getPrevoutHash := func(tx *Transaction) []byte {
		buf := make([]byte, 0)

		for _, in := range tx.Inputs {
			txid, _ := hex.DecodeString(in.PreviousTxID[:])
			buf = append(buf, utils.ReverseBytes(txid)...)
			oi := make([]byte, 4)
			binary.LittleEndian.PutUint32(oi, in.PreviousTxOutIndex)
			buf = append(buf, oi...)
		}

		return crypto.Sha256d(buf)
	}

	getSequenceHash := func(tx *Transaction) []byte {
		buf := make([]byte, 0)

		for _, in := range tx.Inputs {
			oi := make([]byte, 4)
			binary.LittleEndian.PutUint32(oi, in.SequenceNumber)
			buf = append(buf, oi...)
		}

		return crypto.Sha256d(buf)
	}

	getOutputsHash := func(tx *Transaction, n int32) []byte {
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

	if sigHashType&SigHashAnyOneCanPay == 0 {
		// This will be executed in the usual BSV case (where sigHashType = SighashAllForkID)
		hashPrevouts = getPrevoutHash(bt)
	}

	if sigHashType&SigHashAnyOneCanPay == 0 &&
		(sigHashType&31) != SighashSingle &&
		(sigHashType&31) != SighashNone {
		// This will be executed in the usual BSV case (where sigHashType = SighashAllForkID)
		hashSequence = getSequenceHash(bt)
	}

	if (sigHashType&31) != SighashSingle && (sigHashType&31) != SighashNone {
		// This will be executed in the usual BSV case (where sigHashType = SighashAllForkID)
		hashOutputs = getOutputsHash(bt, -1)
	} else if (sigHashType&31) == SighashSingle && inputNumber < uint32(len(bt.Outputs)) {
		// This will *not* be executed in the usual BSV case (where sigHashType = SighashAllForkID)
		hashOutputs = getOutputsHash(bt, int32(inputNumber))
	}

	buf := make([]byte, 0)

	// Version
	v := make([]byte, 4)
	binary.LittleEndian.PutUint32(v, bt.Version)
	buf = append(buf, v...)

	// Input prevouts/nSequence (none/all, depending on flags)
	buf = append(buf, hashPrevouts...)
	buf = append(buf, hashSequence...)

	//  outpoint (32-byte hash + 4-byte little endian)
	txid, _ := hex.DecodeString(in.PreviousTxID[:])
	buf = append(buf, utils.ReverseBytes(txid)...)
	oi := make([]byte, 4)
	binary.LittleEndian.PutUint32(oi, in.PreviousTxOutIndex)
	buf = append(buf, oi...)

	// scriptCode of the input (serialized as scripts inside CTxOuts)
	buf = append(buf, utils.VarInt(uint64(len(*in.PreviousTxScript)))...)
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

	// Locktime
	lt := make([]byte, 4)
	binary.LittleEndian.PutUint32(lt, bt.Locktime)
	buf = append(buf, lt...)

	// sighashType
	// writer.writeUInt32LE(sighashType >>> 0)
	st := make([]byte, 4)
	binary.LittleEndian.PutUint32(st, uint32(sigHashType)>>0)
	buf = append(buf, st...)
	ret := crypto.Sha256d(buf)
	return utils.ReverseBytes(ret), nil
}
