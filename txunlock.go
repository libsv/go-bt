package bt

import (
	"context"

	"github.com/libsv/go-bt/v2/sighash"
)

// UnlockInputParams params used for unlocking an input with a `bt.Unlocker`.
type UnlockInputParams struct {
	// Unlocker to be used. [REQUIRED]
	Unlocker Unlocker
	// InputIdx the input to be unlocked. [DEFAULT 0]
	InputIdx uint32
	// SigHashFlags the be applied [DEFAULT ALL|FORKID]
	SigHashFlags sighash.Flag
}

// UnlockInput is used to unlock the transaction at a specific input index.
// It takes an Unlocker interface as a parameter so that different
// unlocking implementations can be used to unlock the transaction -
// for example local or external unlocking (hardware wallet), or
// signature/nonsignature based.
func (tx *Tx) UnlockInput(ctx context.Context, params UnlockInputParams) error {
	if params.Unlocker == nil {
		return ErrNoUnlocker
	}

	if params.SigHashFlags == 0 {
		params.SigHashFlags = sighash.AllForkID
	}

	return params.Unlocker.Unlock(ctx, tx, params.InputIdx, params.SigHashFlags)
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

		if err = tx.UnlockInput(ctx, UnlockInputParams{
			Unlocker:     u,
			InputIdx:     uint32(i),
			SigHashFlags: sighash.AllForkID, // use SIGHASHALLFORFORKID to sign automatically
		}); err != nil {
			return err
		}
	}

	return nil
}
