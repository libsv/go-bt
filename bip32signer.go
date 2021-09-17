package bt

import (
	"context"
)

type Bip32PathGetterFunc func(ctx context.Context, lockingScript string) (path string, err error)

type Bip32SignerDeriver interface {
	DeriveBip32Signer(derivationPath string) (AutoSigner, error)
}
