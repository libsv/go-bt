package bt

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/libsv/go-bk/crypto"
	"github.com/libsv/go-bt/v2/bscript"
	"github.com/libsv/go-bt/v2/sighash"
)

// Sign is used to sign the transaction at a specific input index.
// It takes a Signed interface as a parameter so that different
// signing implementations can be used to sign the transaction -
// for example internal/local or external signing (hardware wallet).
func (tx *Tx) Sign(ctx context.Context, s Signer, index uint32, shf sighash.Flag) error {
	if shf == 0 {
		shf = sighash.AllForkID
	}
	pubKey, sig, err := s.Sign(ctx, tx, index, shf)
	if err != nil {
		return err
	}
	return tx.ApplyP2PKHUnlockingScript(index, pubKey, sig, shf)
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
func (tx *Tx) SignHash(ctx context.Context, s Signer, index uint32, shf sighash.Flag) error {
	if shf == 0 {
		shf = sighash.AllForkID
	}

	sh, err := tx.CalcInputSignatureHash(index, shf)
	if err != nil {
		return err
	}

	pubKey, sig, err := s.SignHash(ctx, sh)
	if err != nil {
		return err
	}

	return tx.ApplyP2PKHUnlockingScript(index, pubKey, sig, shf)
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

// SignAuto is used to automatically check which P2PKH Inputs are
// able to be signed (match the public key) and then sign them.
// It takes a Signed interface as a parameter so that different
// signing implementations can be used to sign the transaction -
// for example internal/local or external signing.
func (tx *Tx) SignAuto(ctx context.Context, s AutoSigner) (inputsSigned []int, err error) {
	shf := sighash.AllForkID // use SIGHASHALLFORFORKID to sign automatically

	for i, in := range tx.Inputs {
		pubKeyHash, _ := in.PreviousTxScript.PublicKeyHash() // doesn't matter if returns error (not p2pkh)
		pubKeyHashStr := hex.EncodeToString(pubKeyHash)

		pubKey, err := s.PublicKey(ctx)
		if err != nil {
			return nil, err
		}
		pubKeyHashStrFromPriv := hex.EncodeToString(crypto.Hash160(pubKey))

		// check if able to sign (public key matches pubKeyHash in script)
		if pubKeyHashStr == pubKeyHashStrFromPriv {
			if err := tx.Sign(ctx, s, uint32(i), shf); err != nil {
				return nil, err
			}
			inputsSigned = append(inputsSigned, i)
		}
	}

	return
}
