package transaction

import (
	"bytes"
	"cryptolib"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"log"

	"github.com/btcsuite/btcd/btcec"
)

// Signature struct
type Signature struct {
	PublicKey    *btcec.PublicKey
	PreviousTXID [32]byte
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

	for idx, input := range transaction.Inputs {
		// Get the value of the previous input
		s, _ := hex.DecodeString("76a91403ececf2d12a7f614aef4c82ecf13c303bd9975d88ac")
		input.script = NewScript(s)
		input.previousTxAmount = 4998000000

		for _, privateKey := range privateKeys {
			sig := getSignatureForInput(input, transaction, privateKey, uint32(idx), sigtype)
			fmt.Printf("SIG: %x\n", sig[0].Signature)

			sigs = append(sigs, sig...)
		}
	}

	return sigs
}

func getSignatureForInput(input *Input, transaction *BitcoinTransaction, privateKey *btcec.PrivateKey, index uint32, sigtype uint32) []*Signature {
	sigs := make([]*Signature, 0)

	hashData := hash160(privateKey.PubKey().SerializeCompressed())
	fmt.Printf("%x\n", hashData)

	if bytes.Compare(hashData, input.script.getPublicKeyHash()) == 0 {
		sighash := sighashForForkID(transaction, sigtype, index, *input.script, input.previousTxAmount)
		fmt.Printf("SIGHASH: %x\n", sighash)

		s, err := privateKey.Sign(cryptolib.ReverseBytes(sighash))
		if err != nil {
			log.Printf("ERROR [%v]", err)
		}

		signature := s.Serialize()
		signature = append(signature, byte(sigtype))

		if err != nil {
			log.Printf("ERROR [%v]", err)
		} else {
			sigs = append(sigs, &Signature{
				PublicKey:    privateKey.PubKey(),
				PreviousTXID: input.previousTxHash,
				OutputIndex:  input.previousTxOutIndex,
				InputIndex:   index,
				Signature:    signature,
				SigType:      sigtype,
			})
		}
	}

	return sigs
}

func sighashForForkID(transaction *BitcoinTransaction, sighashType uint32, inputNumber uint32, subscript Script, satoshis uint64) []byte {

	input := transaction.Inputs[inputNumber]

	getPrevoutHash := func(tx *BitcoinTransaction) []byte {
		buf := make([]byte, 0)

		for _, in := range tx.Inputs {
			buf = append(buf, cryptolib.ReverseBytes(in.previousTxHash[:])...)
			oi := make([]byte, 4)
			binary.LittleEndian.PutUint32(oi, in.previousTxOutIndex)
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
		fmt.Printf("%x\n", hashPrevouts)
	}

	if sighashType&SighashAnyoneCanPay == 0 &&
		(sighashType&31) != SighashSingle &&
		(sighashType&31) != SighashNone {
		hashSequence = getSequenceHash(transaction)
		fmt.Printf("%x\n", hashSequence)
	}

	if (sighashType&31) != SighashSingle && (sighashType&31) != SighashNone {
		hashOutputs = getOutputsHash(transaction, -1)
	} else if (sighashType&31) == SighashSingle && inputNumber < uint32(len(transaction.Outputs)) {
		hashOutputs = getOutputsHash(transaction, int32(inputNumber))
	}
	fmt.Printf("%x\n", hashOutputs)

	buf := make([]byte, 0)

	// Version
	v := make([]byte, 4)
	binary.LittleEndian.PutUint32(v, transaction.Version)
	buf = append(buf, v...)

	// Input prevouts/nSequence (none/all, depending on flags)
	buf = append(buf, hashPrevouts...)
	buf = append(buf, hashSequence...)

	//  outpoint (32-byte hash + 4-byte little endian)
	buf = append(buf, cryptolib.ReverseBytes(input.previousTxHash[:])...)
	oi := make([]byte, 4)
	binary.LittleEndian.PutUint32(oi, input.previousTxOutIndex)
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
	binary.LittleEndian.PutUint32(seq, input.sequenceNumber)
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
	return cryptolib.ReverseBytes(ret)
}
