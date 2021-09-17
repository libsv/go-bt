package bt

import (
	"context"
)

// Bip32PathGetterFunc interfaces the retreival of a derived path for signing
// transacations in accordance with bip32
type Bip32PathGetterFunc func(ctx context.Context, lockingScript string) (path string, err error)

// Bip32SignerDeriver interfaces the building of signers drived from a path.
type Bip32SignerDeriver interface {
	DeriveBip32Signer(derivationPath string) (AutoSigner, error)
}
