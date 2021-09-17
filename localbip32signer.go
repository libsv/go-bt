package bt

import (
	"context"

	"github.com/libsv/go-bk/bip32"
	"github.com/libsv/go-bt/v2/sighash"
)

type LocalBip32SignerDeriver struct {
	MasterPrivateKey *bip32.ExtendedKey
}

func (b *LocalBip32SignerDeriver) DeriveBip32Signer(derivationPath string) (AutoSigner, error) {
	derivKey, err := b.MasterPrivateKey.DeriveChildFromPath(derivationPath)
	if err != nil {
		return nil, err
	}
	return &LocalBip32Signer{derivKey: derivKey}, nil
}

type LocalBip32Signer struct {
	derivKey *bip32.ExtendedKey
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
	privKey, err := l.derivKey.ECPrivKey()
	if err != nil {
		return
	}

	sig, err := privKey.Sign(hash)
	if err != nil {
		return
	}
	publicKey, err = l.PublicKey(ctx)
	if err != nil {
		return
	}
	signature = sig.Serialise()
	return
}

func (l *LocalBip32Signer) PublicKey(ctx context.Context) (publicKey []byte, err error) {
	pubKey, err := l.derivKey.ECPubKey()
	if err != nil {
		return nil, err
	}
	return pubKey.SerialiseCompressed(), nil
}
