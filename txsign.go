package bt

import (
	"encoding/hex"
	"fmt"

	"github.com/libsv/go-bt/bscript"
	"github.com/libsv/go-bt/crypto"
	"github.com/libsv/go-bt/sighash"
)

// Sign is used to sign the transaction at a specific input index.
// It takes a Signed interface as a parameter so that different
// signing implementations can be used to sign the transaction -
// for example internal/local or external signing (hardware wallet).
func (tx *Tx) Sign(s Signer, index uint32, shf sighash.Flag) error {
	if shf == 0 {
		shf = sighash.AllForkID
	}

	pubkey, sig, err := s.Sign(tx, index, shf)
	if err != nil {
		return err
	}

	return tx.ApplyP2PKHUnlockingScript(index, pubkey, sig, shf)
}

// SignHash is used to sign the transaction at a specific input index.
// It takes a Signed interface as a parameter so that different
// signing implementations can be used to sign the transaction -
// for example internal/local or external signing (hardware wallet).
//
// SignHash will only
// take the final signature hash to be signed so will need to trust that
// it is getting the right hash to sign as there no way to verify that
// it is signing the right hash.
func (tx *Tx) SignHash(s Signer, index uint32, shf sighash.Flag) error {
	if shf == 0 {
		shf = sighash.AllForkID
	}

	sh, err := tx.CalcInputSignatureHash(index, shf)
	if err != nil {
		return err
	}

	pubkey, sig, err := s.SignHash(sh)
	if err != nil {
		return err
	}

	return tx.ApplyP2PKHUnlockingScript(index, pubkey, sig, shf)
}

// ApplyP2PKHUnlockingScript applies a script to the transaction at a specific index in
// unlocking script field.
func (tx *Tx) ApplyP2PKHUnlockingScript(index uint32, pubKey []byte, sig []byte, shf sighash.Flag) error {
	uls, err := bscript.NewP2PKHUnlockingScript(pubKey, sig, shf)
	if err != nil {
		return err
	}

	return tx.ApplyUnlockingScript(index, uls)
}

// ApplyUnlockingScript applies a script to the transaction at a specific index in
// unlocking script field.
func (tx *Tx) ApplyUnlockingScript(index uint32, s *bscript.Script) error {
	if tx.Inputs[index] != nil {
		tx.Inputs[index].UnlockingScript = s
		return nil
	}

	return fmt.Errorf("no input at index %d", index)
}

// SignAuto is used to automatically check which P2PKH inputs are
// able to be signed (match the public key) and then sign them.
// It takes a Signed interface as a parameter so that different
// signing implementations can be used to sign the transaction -
// for example internal/local or external signing.
func (tx *Tx) SignAuto(s AutoSigner) (inputsSigned []int, err error) {

	shf := sighash.AllForkID // use SIGHASHALLFORFORKID to sign automatically

	for i, in := range tx.Inputs {
		pubKeyHash, _ := in.PreviousTxScript.PublicKeyHash() // doesn't matter if returns error (not p2pkh)
		pubKeyHashStr := hex.EncodeToString(pubKeyHash)

		pubKeyHashStrFromPriv := hex.EncodeToString(crypto.Hash160(s.PublicKey()))

		// check if able to sign (public key matches pubKeyHash in script)
		if pubKeyHashStr == pubKeyHashStrFromPriv {
			if err = tx.Sign(s, uint32(i), shf); err != nil {
				return
			}
			inputsSigned = append(inputsSigned, i)
		}
	}

	return
}
