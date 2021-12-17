package bt

import (
	"context"

	"github.com/libsv/go-bt/v2/sighash"
)

// UnlockInput is used to unlock the transaction at a specific input index.
// It takes an Unlocker interface as a parameter so that different
// unlocking implementations can be used to unlock the transaction -
// for example local or external unlocking (hardware wallet), or
// signature/nonsignature based.
func (tx *Tx) UnlockInput(ctx context.Context, unlocker Unlocker, params UnlockerParams) error {
	if unlocker == nil {
		return ErrNoUnlocker
	}

	if params.SigHashFlags == 0 {
		params.SigHashFlags = sighash.AllForkID
	}

	uscript, err := unlocker.UnlockingScript(ctx, tx, params)
	if err != nil {
		return err
	}

	return tx.InsertInputUnlockingScript(params.InputIdx, uscript)
}

// UnlockAllInputs is used to sign all inputs. It takes an UnlockerGetter interface
// as a parameter so that different unlocking implementations can
// be used to sign the transaction - for example local/external
// signing, or P2PKH/contract signing.
//
// Given this signs inputs and outputs, sighash `ALL|FORKID` is used.
func (tx *Tx) UnlockAllInputs(ctx context.Context, ug UnlockerGetter) error {
	for i, in := range tx.Inputs {
		u, err := ug.Unlocker(ctx, in.PreviousTxScript)
		if err != nil {
			return err
		}

		if err = tx.UnlockInput(ctx, u, UnlockerParams{
			InputIdx:     uint32(i),
			SigHashFlags: sighash.AllForkID, // use SIGHASHALLFORFORKID to sign automatically
		}); err != nil {
			return err
		}
	}

	return nil
}
