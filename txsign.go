package bt

import (
	"context"
	"errors"
	"fmt"

	"github.com/libsv/go-bt/v2/bscript"
	"github.com/libsv/go-bt/v2/sighash"
)

type signerFunc func(context.Context, Signer, uint32, sighash.Flag) error

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

// SignAll is used to sign all inputs. It currently only supports the signing P2PKH.
// It takes a Signed interface as a parameter so that different
// signing implementations can be used to sign the transaction -
// for example internal/local or external signing.
func (tx *Tx) SignAll(ctx context.Context, sg SignerGetter) error {
	shf := sighash.AllForkID // use SIGHASHALLFORFORKID to sign automatically
	// TODO: add support for other script types
	signerStrats := map[string]signerFunc{
		bscript.ScriptTypePubKeyHash: tx.Sign,
	}

	for i, in := range tx.Inputs {
		// TODO: add support for other script types
		stratFn, ok := signerStrats[in.PreviousTxScript.ScriptType()]
		if !ok {
			return errors.New("unsupported script type")
		}
		s, err := sg.Signer(ctx, in.PreviousTxScript)
		if err != nil {
			return err
		}
		if err := stratFn(ctx, s, uint32(i), shf); err != nil {
			return err
		}
	}

	return nil
}
