package bt

import (
	"context"

	"github.com/libsv/go-bk/bip32"
	"github.com/libsv/go-bt/v2/sighash"
)

type LocalBip32SignerBuilder struct {
	PrivateKey *bip32.ExtendedKey
}

func (b *LocalBip32SignerBuilder) NewSigner() Bip32Signer {
	return &LocalBip32Signer{privKey: b.PrivateKey}
}

type LocalBip32Signer struct {
	privKey *bip32.ExtendedKey
	path    string
}

func (l *LocalBip32Signer) Sign(ctx context.Context, unsignedTx *Tx, index uint32,
	shf sighash.Flag) (publicKey []byte, signature []byte, err error) {

	if shf == 0 {
		shf = sighash.AllForkID
	}

	var sh []byte
	if sh, err = unsignedTx.CalcInputSignatureHash(index, shf); err != nil {
		return
	}

	return l.SignHash(ctx, sh)
}

func (l *LocalBip32Signer) SignHash(ctx context.Context, hash []byte) (publicKey, signature []byte, err error) {
	privKey, err := l.privKey.ECPrivKey()
	if err != nil {
		return
	}
	sig, err := privKey.Sign(hash)
	if err != nil {
		return
	}

	publicKey, err = l.privKey.DerivePublicKeyFromPath(l.path)
	if err != nil {
		return
	}
	signature = sig.Serialise()
	return
}

func (l *LocalBip32Signer) PublicKey(ctx context.Context, derivationPath string) (publicKey []byte, err error) {
	l.path = derivationPath
	return l.privKey.DerivePublicKeyFromPath(derivationPath)
}
