package bt

import (
	"context"

	"github.com/libsv/go-bt/v2/bscript"
	"github.com/libsv/go-bt/v2/sighash"
)

// Signer interface to allow custom implementations of different signing mechanisms.
// Implement the Sign function as shown in InternalSigner, for example. Sign generates
// and returns an ECDSA signature for the provided hash digest using the private key
// as well as the public key corresponding to the private key used. The produced
// signature is deterministic (same message and same key yield the same signature) and
// canonical in accordance with RFC6979 and BIP0062.
type Signer interface {
	Sign(ctx context.Context, unsignedTx *Tx, index uint32, shf sighash.Flag) (publicKey, signature []byte, err error)
	SignHash(ctx context.Context, hash []byte) (publicKey, signature []byte, err error)
}

// SignerGetter interfaces getting a signer for a given output/locking script.
type SignerGetter interface {
	Signer(ctx context.Context, lockingScript *bscript.Script) (Signer, error)
}
