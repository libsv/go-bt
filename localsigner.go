package bt

import (
	"github.com/libsv/go-bt/bec"
	"github.com/libsv/go-bt/sighash"
)

// LocalSigner implements the Signer interface. It is used to sign a Tx locally
// using a bec PrivateKey.
type LocalSigner struct {
	PrivateKey *bec.PrivateKey
}

// Sign a transaction at a given input index using the PrivateKey passed in through the
// InternalSigner struct.
func (is *LocalSigner) Sign(unsignedTx *Tx, index uint32,
	shf sighash.Flag) (publicKey []byte, signature []byte, err error) {

	if shf == 0 {
		shf = sighash.AllForkID
	}

	var sh []byte
	if sh, err = unsignedTx.CalcInputSignatureHash(index, shf); err != nil {
		return
	}

	return is.SignHash(sh)
}

// SignHash a transaction at a given a hash digest using the PrivateKey passed in through the
// InternalSigner struct.
func (is *LocalSigner) SignHash(hash []byte) (publicKey []byte, signature []byte, err error) {

	sig, err := is.PrivateKey.Sign(hash)
	if err != nil {
		return
	}

	publicKey = is.PrivateKey.PubKey().SerialiseCompressed()
	signature = sig.Serialise()
	return
}

// PublicKey returns the public key which will be used to sign.
func (is *LocalSigner) PublicKey() (publicKey []byte) {
	return is.PrivateKey.PubKey().SerialiseCompressed()
}