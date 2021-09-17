package bt

import (
	"context"
)

type Bip32PathGetterFunc func(ctx context.Context, lockingScript string) (path string, err error)

type Bip32SignerBuilder interface {
	NewSigner() Bip32Signer
}

type Bip32Signer interface {
	Signer
	PublicKey(ctx context.Context, derivationPath string) (publicKey []byte, err error)
}
