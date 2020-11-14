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
func (is *InternalSigner) Sign(index uint32, unsignedTx *Tx) (*Tx, error) {
	if is.SigHashFlag == 0 {
		is.SigHashFlag = sighash.AllForkID
	}

	sh, err := unsignedTx.GetInputSignatureHash(index, is.SigHashFlag)
	if err != nil {
		return nil, err
	}

	var sig *bsvec.Signature
	sig, err = is.PrivateKey.Sign(ReverseBytes(sh)) // little endian sign
	if err != nil {
		return nil, err
	}

	var s *bscript.Script
	s, err = bscript.NewP2PKHUnlockingScript(
		is.PrivateKey.PubKey().SerializeCompressed(),
		sig.Serialize(),
		is.SigHashFlag,
	)
	if err != nil {
		return nil, err
	}

	err = unsignedTx.ApplyUnlockingScript(index, s)
	if err != nil {
		return nil, err
	}

	return unsignedTx, nil
}

// SignAuto goes through each input of the transaction and automatically
// signs the P2PKH inputs that it is able to sign using the specific
// PrivateKey passed in through the InternalSigner struct.
func (is *InternalSigner) SignAuto(unsignedTx *Tx) (*Tx, error) {
	if is.SigHashFlag == 0 {
		is.SigHashFlag = sighash.AllForkID
	}

	for i, in := range unsignedTx.Inputs {

		pubKeyHash, _ := in.PreviousTxScript.GetPublicKeyHash() // doesn't matter if returns error (not p2pkh)
		pubKeyHashStr := hex.EncodeToString(pubKeyHash)

		pubKeyHashStrFromPriv := hex.EncodeToString(crypto.Hash160(is.PrivateKey.PubKey().SerializeCompressed()))

		// check if able to sign (public key matches pubKeyHash in script)
		if pubKeyHashStr == pubKeyHashStrFromPriv {
			// todo: not sure if the tx value should be used or not? @mrz
			_, err := is.Sign(uint32(i), unsignedTx)
			if err != nil {
				return nil, err
			}
		}
	}

	return unsignedTx, nil
}
