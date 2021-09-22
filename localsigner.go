package bt

import (
	"context"

	"github.com/libsv/go-bk/bec"
	"github.com/libsv/go-bt/v2/bscript"
	"github.com/libsv/go-bt/v2/sighash"
)

// LocalSigner implements the Signer interface. It is used to sign Tx Inputs locally
// using a bkec PrivateKey.
type LocalSigner struct {
	PrivateKey *bec.PrivateKey
}

// LocalSignerCreator implements the SignerCreator interface. It is used to sign Tx Inputs locally
// using a bkec PrivateKey.
type LocalSignerCreator struct {
	PrivateKey *bec.PrivateKey
}

// Create builds a new *bt.LocalSigner with the same private key as the calling *bt.LocalSignerCreator
func (lsc *LocalSignerCreator) Create(ctx context.Context, lockingScript *bscript.Script) (Signer, error) {
	return &LocalSigner{PrivateKey: lsc.PrivateKey}, nil
}

// Sign a transaction at a given input index using the PrivateKey passed in through the
// InternalSigner struct.
func (is *LocalSigner) Sign(ctx context.Context, unsignedTx *Tx, index uint32,
	shf sighash.Flag) (publicKey []byte, signature []byte, err error) {

	if shf == 0 {
		shf = sighash.AllForkID
	}

	var sh []byte
	if sh, err = unsignedTx.CalcInputSignatureHash(index, shf); err != nil {
		return
	}

	return is.SignHash(ctx, sh)
}

// SignHash a transaction at a given a hash digest using the PrivateKey passed in through the
// InternalSigner struct.
func (is *LocalSigner) SignHash(ctx context.Context, hash []byte) (publicKey, signature []byte, err error) {
	sig, err := is.PrivateKey.Sign(hash)
	if err != nil {
		return
	}

	publicKey = is.PrivateKey.PubKey().SerialiseCompressed()
	signature = sig.Serialise()
	return
}
