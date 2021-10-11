package bt

import (
	"context"
	"fmt"

	"github.com/libsv/go-bt/v2/bscript"
	"github.com/libsv/go-bt/v2/sighash"
)

// Unlock is used to unlock the transaction at a specific input index.
// It takes an Unlocker interface as a parameter so that different
// unlocking implementations can be used to unlock the transaction -
// for example local or external unlocking (hardware wallet), or
// signature/nonsignature based.
func (tx *Tx) Unlock(ctx context.Context, u Unlocker, idx uint32, shf sighash.Flag) error {
	if shf == 0 {
		shf = sighash.AllForkID
	}

	return u.Unlock(ctx, tx, idx, shf)
}

// UnlockAll is used to sign all inputs. It takes an UnlockerGetter interface
// as a parameter so that different unlocking implementations can
// be used to sign the transaction - for example local/external
// signing, or P2PKH/contract signing.
func (tx *Tx) UnlockAll(ctx context.Context, ug UnlockerGetter) error {
	shf := sighash.AllForkID // use SIGHASHALLFORFORKID to sign automatically

	for i, in := range tx.Inputs {
		u, err := ug.Unlocker(ctx, in.PreviousTxScript)
		if err != nil {
			return err
		}

		if err = tx.Unlock(ctx, u, uint32(i), shf); err != nil {
			return err
		}
	}

	return nil
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
