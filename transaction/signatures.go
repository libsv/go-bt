package transaction

import (
	"github.com/btcsuite/btcd/btcec"
)

// Signature struct
type Signature struct {
	PublicKey    *btcec.PublicKey
	PreviousTXID string
	OutputIndex  uint32
	InputIndex   uint32
	Signature    []byte
	SigType      uint32
}

/*

P2PKH - Pay to public key hash
------------------------------
Locking script:   OP_DUP OP_HASH160 <Public Key Hash> OP_EQUAL OP_CHECKSIG
Unlocking script: <Signature> <Public Key>

P2PK - Pay to public key
------------------------
Locking script:   <Public Key A> OP_CHECKSIG
Unlocking script: <Signature from Private Key A>

P2MS - Multisignature (M of N)
------------------------------
Locking script:   M <Public Key 1> <Public Key 2> ... <Public Key N> N OP_CHECKMULTISIG
Unlocking script: OP_0 <Signature B> <Signature C>

P2SK - Pay to script hash (M or N)
----------------------------------
Locking script:   OP_HASH160 <20-byte hash of redeem script> OP_EQUAL
Redeem script:    M PubKey1 PubKey2 PubKey3 PubKey4 PubKey5 N OP_CHECKMULTISIG
Unlocking script: <Sig1> <Sig2> <redeem script>
*/

// The five standard types of transaction scripts are:
// P2PKH: pay-to-public-key-hash - OP_DUP OP_HASH160 <Public Key Hash> OP_EQUAL OP_CHECKSIG
// PK:    public-key
// MS:    multi-signature (limited to 15 keys)
// P2SH:  pay-to-script-hash
// OP_RETURN: Not implemented here

func getInputType(in *Input) {
	// P2SH
	// P2PKH
	// P2PK
}

func getSignatures(transaction *BitcoinTransaction, privateKeys []*btcec.PrivateKey, sigtype uint32) []*Signature {
	sigs := make([]*Signature, 0)

	for _, privateKey := range privateKeys {
		for idx, input := range transaction.Inputs {
			sigs = append(sigs, getSignatureForInput(input, transaction, privateKey, uint32(idx), sigtype)...)
		}
	}

	return sigs
}

func getSignatureForInput(input *Input, transaction *BitcoinTransaction, privateKey *btcec.PrivateKey, index uint32, sigtype uint32) []*Signature {
	sigs := make([]*Signature, 0)

	// hashData := hash160(privateKey.PubKey().SerializeCompressed())

	// if bytes.Compare(hashData, transaction.GetInputs()[index].GetInputScript()) == 0 {
	// 	sigs = append(sigs, &Signature{
	// 		PublicKey:    privateKey.PubKey(),
	// 		PreviousTXID: input.previousTxHash,
	// 		OutputIndex:  input.previousTxOutIndex,
	// 		InputIndex:   idx,
	// 		Signature:    sighashForForkId(transaction, sigtype, index, this.output.script, this.output.satoshis),
	// 		SigType:      sigtype,
	// 	})
	// }

	return sigs
}

// func sighashForForkID(transaction *BitcoinTransaction, sighashType uint32, inputNumber uint32, subscript string, satoshis uint64) {
// 	input = transaction.inputs[inputNumber]

// 	getPrevoutHash := func(tx *BitcoinTransaction) []byte {
// 		buf := make([]byte, 0)

// 		for _, in := range tx.Inputs {
// 			buf = append(buf, reverseBytes(in.PreviousTXID)...)
// 			oi := make([]byte, 4)
// 			binary.LittleEndian.PutUint32(oi, in.OutputIndex)
// 			buf = append(buf, oi...)
// 		}

// 		return cryptolib.Sha256d(buf)
// 	}

// 	getSequenceHash := func(tx *BitcoinTransaction) []byte {
// 		buf := make([]byte, 0)

// 		for _, in := range tx.Inputs {
// 			oi := make([]byte, 4)
// 			binary.LittleEndian.PutUint32(oi, in.sequenceNumber)
// 			buf = append(buf, oi...)
// 		}

// 		return cryptolib.Sha256d(buf)
// 	}

// 	getOutputsHash := func(tx *BitcoinTransaction, n int) {
// 		buf := make([]byte, 0)

// 		if n == -1 {
// 			for _, out := range tx.Outputs {
// 				buf = append(buf, out...)
// 			}
// 		} else {
// 			buf = append(buf, tx.Outputs[n]...)
// 		}

// 		return cryptolib.Sha256d(buf)
// 	}

// 	hashPrevouts := make([]byte, 32)
// 	hashSequence := make([]byte, 32)
// 	hashOutputs := make([]byte, 32)

// 	if !(sighashType & Signature.SIGHASH_ANYONECANPAY) {
// 		hashPrevouts = getPrevoutHash(transaction)
// 	}

// 	if !(sighashType & Signature.SIGHASH_ANYONECANPAY) &&
// 		(sighashType&31) != Signature.SIGHASH_SINGLE &&
// 		(sighashType&31) != Signature.SIGHASH_NONE {
// 		hashSequence = getSequenceHash(transaction)
// 	}

// 	if (sighashType&31) != Signature.SIGHASH_SINGLE && (sighashType&31) != Signature.SIGHASH_NONE {
// 		hashOutputs = getOutputsHash(transaction, -1)
// 	} else if (sighashType&31) == Signature.SIGHASH_SINGLE && inputNumber < transaction.outputs.length {
// 		hashOutputs = getOutputsHash(transaction, inputNumber)
// 	}

// 	buf := make([]byte, 0)

// 	// Version
// 	v := make([]byte, 4)
// 	binary.LittleEndian.PutUint32(v, tx.Version)
// 	buf = append(buf, v...)

// 	// Input prevouts/nSequence (none/all, depending on flags)
// 	buf = append(buf, hashPrevouts...)
// 	buf = append(buf, hashSequence...)

// 	//  outpoint (32-byte hash + 4-byte little endian)
// 	buf = append(buf, reverseBytes(in.PreviousTXID)...)
// 	oi := make([]byte, 4)
// 	binary.LittleEndian.PutUint32(oi, in.OutputIndex)
// 	buf = append(buf, oi...)

// 	// scriptCode of the input (serialized as scripts inside CTxOuts)
// 	buf = append(buf, Varint(len(subscript))...)
// 	buf = append(buf, subscript...)

// 	// value of the output spent by this input (8-byte little endian)
// 	sat := make([]byte, 8)
// 	binary.LittleEndian.PutUint64(sat, satoshis)
// 	buf = append(buf, sat...)

// 	// nSequence of the input (4-byte little endian)
// 	seq := make([]byte, 4)
// 	binary.LittleEndian.PutUint32(seq, in.sequenceNumber)
// 	buf = append(buf, seq...)

// 	// Outputs (none/one/all, depending on flags)
// 	buf = append(buf, hashOutputs...)

// 	// Locktime
// 	lt := make([]byte, 4)
// 	binary.LittleEndian.PutUint32(lt, tx.LockTime)
// 	buf = append(buf, lt...)

// 	// sighashType
// 	//writer.writeUInt32LE(sighashType >>> 0)
// 	st := make([]byte, 4)
// 	binary.LittleEndian.PutUint32(st, sighashType>>0)
// 	buf = append(buf, st...)

// 	ret := cryptolib.Sha256d(buf)
// 	return ReverseBytes(ret)
// }

/*
def SignatureHashForkId(script, txTo, inIdx, hashtype, amount):

    hashPrevouts = 0
    hashSequence = 0
    hashOutputs = 0

    if not (hashtype & SIGHASH_ANYONECANPAY):
        serialize_prevouts = bytes()
        for i in txTo.vin:
            serialize_prevouts += i.prevout.serialize()
        hashPrevouts = uint256_from_str(hash256(serialize_prevouts))

    if (not (hashtype & SIGHASH_ANYONECANPAY) and (hashtype & 0x1f) != SIGHASH_SINGLE and (hashtype & 0x1f) != SIGHASH_NONE):
        serialize_sequence = bytes()
        for i in txTo.vin:
            serialize_sequence += struct.pack("<I", i.nSequence)
        hashSequence = uint256_from_str(hash256(serialize_sequence))

    if ((hashtype & 0x1f) != SIGHASH_SINGLE and (hashtype & 0x1f) != SIGHASH_NONE):
        serialize_outputs = bytes()
        for o in txTo.vout:
            serialize_outputs += o.serialize()
        hashOutputs = uint256_from_str(hash256(serialize_outputs))
    elif ((hashtype & 0x1f) == SIGHASH_SINGLE and inIdx < len(txTo.vout)):
        serialize_outputs = txTo.vout[inIdx].serialize()
        hashOutputs = uint256_from_str(hash256(serialize_outputs))

    ss = bytes()
    ss += struct.pack("<i", txTo.nVersion)
    ss += ser_uint256(hashPrevouts)
    ss += ser_uint256(hashSequence)
    ss += txTo.vin[inIdx].prevout.serialize()
    ss += ser_string(script)
    ss += struct.pack("<q", amount)
    ss += struct.pack("<I", txTo.vin[inIdx].nSequence)
    ss += ser_uint256(hashOutputs)
    ss += struct.pack("<i", txTo.nLockTime)
    ss += struct.pack("<I", hashtype)

    return hash256(ss)

*/

/*
uint256 GetPrevoutHash(const CTransaction &txTo) {
    CHashWriter ss(SER_GETHASH, 0);
    for (size_t n = 0; n < txTo.vin.size(); n++) {
        ss << txTo.vin[n].prevout;
    }
    return ss.GetHash();
}

uint256 GetSequenceHash(const CTransaction &txTo) {
    CHashWriter ss(SER_GETHASH, 0);
    for (size_t n = 0; n < txTo.vin.size(); n++) {
        ss << txTo.vin[n].nSequence;
    }
    return ss.GetHash();
}

uint256 GetOutputsHash(const CTransaction &txTo) {
    CHashWriter ss(SER_GETHASH, 0);
    for (size_t n = 0; n < txTo.vout.size(); n++) {
        ss << txTo.vout[n];
    }
    return ss.GetHash();
}

} // namespace

PrecomputedTransactionData::PrecomputedTransactionData(
    const CTransaction &txTo) {
    hashPrevouts = GetPrevoutHash(txTo);
    hashSequence = GetSequenceHash(txTo);
    hashOutputs = GetOutputsHash(txTo);
}

uint256 SignatureHash(const CScript &scriptCode, const CTransaction &txTo,
                      unsigned int nIn, SigHashType sigHashType,
                      const Amount amount,
                      const PrecomputedTransactionData *cache, uint32_t flags) {
    if (sigHashType.hasForkId() && (flags & SCRIPT_ENABLE_SIGHASH_FORKID)) {
        uint256 hashPrevouts;
        uint256 hashSequence;
        uint256 hashOutputs;

        if (!sigHashType.hasAnyoneCanPay()) {
            hashPrevouts = cache ? cache->hashPrevouts : GetPrevoutHash(txTo);
        }

        if (!sigHashType.hasAnyoneCanPay() &&
            (sigHashType.getBaseType() != BaseSigHashType::SINGLE) &&
            (sigHashType.getBaseType() != BaseSigHashType::NONE)) {
            hashSequence = cache ? cache->hashSequence : GetSequenceHash(txTo);
        }

        if ((sigHashType.getBaseType() != BaseSigHashType::SINGLE) &&
            (sigHashType.getBaseType() != BaseSigHashType::NONE)) {
            hashOutputs = cache ? cache->hashOutputs : GetOutputsHash(txTo);
        } else if ((sigHashType.getBaseType() == BaseSigHashType::SINGLE) &&
                   (nIn < txTo.vout.size())) {
            CHashWriter ss(SER_GETHASH, 0);
            ss << txTo.vout[nIn];
            hashOutputs = ss.GetHash();
        }

        CHashWriter ss(SER_GETHASH, 0);
        // Version
        ss << txTo.nVersion;
        // Input prevouts/nSequence (none/all, depending on flags)
        ss << hashPrevouts;
        ss << hashSequence;
        // The input being signed (replacing the scriptSig with scriptCode +
        // amount). The prevout may already be contained in hashPrevout, and the
        // nSequence may already be contain in hashSequence.
        ss << txTo.vin[nIn].prevout;
        ss << scriptCode;
        ss << amount.GetSatoshis();
        ss << txTo.vin[nIn].nSequence;
        // Outputs (none/one/all, depending on flags)
        ss << hashOutputs;
        // Locktime
        ss << txTo.nLockTime;
        // Sighash type
        ss << sigHashType;

        return ss.GetHash();
    }

*/
