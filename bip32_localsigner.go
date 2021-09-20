package bt

import (
	"github.com/libsv/go-bk/bip32"
)

// LocalBip32SignerDeriver implements the Bip32SignerDeriver interface. It is used to
// sign bip32 Tx Inputs locally using a *bip32.ExtendedKey.
type LocalBip32SignerDeriver struct {
	MasterPrivateKey *bip32.ExtendedKey
}

// DeriveBip32Signer derives a *bt.LocalSigner from the provided derivation path.
func (b *LocalBip32SignerDeriver) DeriveBip32Signer(derivationPath string) (Signer, error) {
	derivKey, err := b.MasterPrivateKey.DeriveChildFromPath(derivationPath)
	if err != nil {
		return nil, err
	}
	privKey, err := derivKey.ECPrivKey()
	if err != nil {
		return nil, err
	}

	return &LocalSigner{PrivateKey: privKey}, nil
}
