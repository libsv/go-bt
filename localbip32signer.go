package bt

import (
	"github.com/libsv/go-bk/bip32"
)

type LocalBip32SignerDeriver struct {
	MasterPrivateKey *bip32.ExtendedKey
}

func (b *LocalBip32SignerDeriver) DeriveBip32Signer(derivationPath string) (AutoSigner, error) {
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
