package bt

import (
	"context"

	"github.com/libsv/go-bt/v2/bscript"
	"github.com/libsv/go-bt/v2/sighash"
)

// Unlocker interface to allow custom implementations of different unlocking mechanisms.
// Implement the Unlocker function as shown in LocalUnlocker, for example.
type Unlocker interface {
	Unlock(ctx context.Context, tx *Tx, idx uint32, shf sighash.Flag) error
}

// UnlockerGetter interfaces getting an unlocker for a given output/locking script.
type UnlockerGetter interface {
	Unlocker(ctx context.Context, lockingScript *bscript.Script) (Unlocker, error)
}
