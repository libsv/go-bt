package bt

import (
	"context"

	"github.com/libsv/go-bt/v2/bscript"
	"github.com/libsv/go-bt/v2/sighash"
)

// UnlockerParams params used for unlocking an input with a `bt.Unlocker`.
type UnlockerParams struct {
	// InputIdx the input to be unlocked. [DEFAULT 0]
	InputIdx uint32
	// SigHashFlags the be applied [DEFAULT ALL|FORKID]
	SigHashFlags sighash.Flag
	// TODO: add previous tx script and sats here instead of in
	// input (and potentially remove from input) - see issue #143
}

// Unlocker interface to allow custom implementations of different unlocking mechanisms.
// Implement the Unlocker function as shown in LocalUnlocker, for example.
type Unlocker interface {
	UnlockingScript(ctx context.Context, tx *Tx, up UnlockerParams) (uscript *bscript.Script, err error)
}

// UnlockerGetter interfaces getting an unlocker for a given output/locking script.
type UnlockerGetter interface {
	Unlocker(ctx context.Context, lockingScript *bscript.Script) (Unlocker, error)
}
