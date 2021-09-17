package bt

import (
	"context"
	"encoding/hex"

	"github.com/libsv/go-bk/crypto"
	"github.com/libsv/go-bt/v2/sighash"
)

// Bip32SignAuto is used to automatically check which P2PKH Inputs are
// able to be signed (match a derivated public key) and then sign them.
// It takes a Bip32SignerDeriver interface as a parameter so that different
// signing implementations can be used to sign the transaction -
// for example internal/local or external signing.
// It also takes Bip32PathGetterFunc with is expected to return a correct derivation,
// path for a given input, to allow matching and signing.
func (tx *Tx) Bip32SignAuto(ctx context.Context, b Bip32SignerDeriver,
	fn Bip32PathGetterFunc) (inputsSigned []int, err error) {
	shf := sighash.AllForkID

	for i, in := range tx.Inputs {
		pubKeyHash, _ := in.PreviousTxScript.PublicKeyHash()
		pubKeyHashStr := hex.EncodeToString(pubKeyHash)

		path, err := fn(ctx, in.PreviousTxScript.String())
		if err != nil {
			return nil, err
		}

		signer, err := b.DeriveBip32Signer(path)
		if err != nil {
			return nil, err
		}

		pubKey, err := signer.PublicKey(ctx)
		if err != nil {
			return nil, err
		}

		pubKeyHashStrFromPriv := hex.EncodeToString(crypto.Hash160(pubKey))
		if pubKeyHashStr == pubKeyHashStrFromPriv {
			if err = tx.Sign(ctx, signer, uint32(i), shf); err != nil {
				return nil, err
			}
			inputsSigned = append(inputsSigned, i)
		}
	}
	return
}
