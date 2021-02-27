package bt

import (
	"encoding/hex"

	"github.com/bitcoinsv/bsvd/bsvec"
	"github.com/libsv/go-bt/bscript"
	"github.com/libsv/go-bt/crypto"
	"github.com/libsv/go-bt/sighash"
)

// InternalSigner implements the Signer interface. It is used to sign a Tx locally
// given a PrivateKey and SIGHASH type.
type InternalSigner struct {
	PrivateKey  *bsvec.PrivateKey
	SigHashFlag sighash.Flag
}

// Sign a transaction at a given input index using the PrivateKey passed in through the
// InternalSigner struct.
func (is *InternalSigner) Sign(index uint32, unsignedTx *Tx) (signedTx *Tx, err error) {
	if is.SigHashFlag == 0 {
		is.SigHashFlag = sighash.AllForkID
	}

	// TODO: v2 put tx serialization in parent/general func (tx.Sign) so that the
	// functions like that implement the Signer interface only sign and don't do
	// tx input serialization as well. So this Sign func would probably need to take
	// Sign(sigdigest []bytes)
	var sh []byte
	if sh, err = unsignedTx.GetInputSignatureHash(index, is.SigHashFlag); err != nil {
		return
	}

	var sig *bsvec.Signature
	if sig, err = is.PrivateKey.Sign(ReverseBytes(sh)); err != nil { // little endian sign
		return
	}

	var s *bscript.Script
	if s, err = bscript.NewP2PKHUnlockingScript(
		is.PrivateKey.PubKey().SerializeCompressed(),
		sig.Serialize(),
		is.SigHashFlag,
	); err != nil {
		return
	}

	if err = unsignedTx.ApplyUnlockingScript(index, s); err != nil {
		return
	}
	signedTx = unsignedTx

	return
}

// SignAuto goes through each input of the transaction and automatically
// signs the P2PKH inputs that it is able to sign using the specific
// PrivateKey passed in through the InternalSigner struct.
func (is *InternalSigner) SignAuto(unsignedTx *Tx) (signedTx *Tx, inputsSigned []int, err error) {
	if is.SigHashFlag == 0 {
		is.SigHashFlag = sighash.AllForkID
	}

	for i, in := range unsignedTx.Inputs {
		pubKeyHash, _ := in.PreviousTxScript.GetPublicKeyHash() // doesn't matter if returns error (not p2pkh)

		pubKeyHashStr := hex.EncodeToString(pubKeyHash)

		pubKeyHashStrFromPriv := hex.EncodeToString(crypto.Hash160(is.PrivateKey.PubKey().SerializeCompressed()))

		// check if able to sign (public key matches pubKeyHash in script)
		if pubKeyHashStr == pubKeyHashStrFromPriv {

			if signedTx, err = is.Sign(uint32(i), unsignedTx); err != nil {
				return
			}

			inputsSigned = append(inputsSigned, i)
		}
	}

	if signedTx == nil {
		signedTx = unsignedTx
	}

	return
}
