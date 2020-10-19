package sig

import (
	"encoding/hex"

	"github.com/bitcoinsv/bsvd/bsvec"

	"github.com/libsv/libsv/bt"
	"github.com/libsv/libsv/bt/sig/sighash"
	"github.com/libsv/libsv/crypto"
	"github.com/libsv/libsv/script"
	"github.com/libsv/libsv/utils"
)

// InternalSigner implements the Signer interface. It is used to sign a Tx locally
// given a PrivateKey and SIGHASH type.
type InternalSigner struct {
	PrivateKey  *bsvec.PrivateKey
	SigHashFlag sighash.Flag
}

// Sign a transaction at a given input index using the PrivateKey passed in through the
// InternalSigner struct.
func (is *InternalSigner) Sign(index uint32, unsignedTx *bt.Tx) (*bt.Tx, error) {
	if is.SigHashFlag == 0 {
		is.SigHashFlag = sighash.AllForkID
	}

	sh, err := unsignedTx.GetInputSignatureHash(index, is.SigHashFlag)
	if err != nil {
		return nil, err
	}

	sig, err := is.PrivateKey.Sign(utils.ReverseBytes(sh)) // little endian sign
	if err != nil {
		return nil, err
	}

	s, err := script.NewP2PKHUnlockingScript(is.PrivateKey.PubKey().SerializeCompressed(), sig.Serialize(), is.SigHashFlag)

	err = unsignedTx.ApplyUnlockingScript(index, s)
	if err != nil {
		return nil, err
	}

	return unsignedTx, nil
}

// SignAuto goes through each input of the transaction and automatically
// signs the P2PKH inputs that it is able to sign using the specific
// PrivateKey passed in through the InternalSigner struct.
func (is *InternalSigner) SignAuto(unsignedTx *bt.Tx) (*bt.Tx, error) {
	if is.SigHashFlag == 0 {
		is.SigHashFlag = sighash.AllForkID
	}

	for i, in := range unsignedTx.Inputs {

		pubKeyHash, _ := in.PreviousTxScript.GetPublicKeyHash() // doesn't matter if returns error (not p2pkh)
		pubKeyHashStr := hex.EncodeToString(pubKeyHash)

		pubKeyHashStrFromPriv := hex.EncodeToString(crypto.Hash160(is.PrivateKey.PubKey().SerializeCompressed()))

		// check if able to sign (public key matches pubKeyHash in script)
		if pubKeyHashStr == pubKeyHashStrFromPriv {
			is.Sign(uint32(i), unsignedTx)
		}
	}

	return unsignedTx, nil
}
