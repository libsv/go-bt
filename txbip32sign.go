package bt

import (
	"context"
	"encoding/hex"

	"github.com/libsv/go-bk/crypto"
	"github.com/libsv/go-bt/v2/sighash"
)

func (tx *Tx) AutoSignBip32(ctx context.Context, b Bip32SignerBuilder, fn Bip32PathGetterFunc) error {
	shf := sighash.AllForkID

	for i, in := range tx.Inputs {
		pubKeyHash, _ := in.PreviousTxScript.PublicKeyHash()
		pubKeyHashStr := hex.EncodeToString(pubKeyHash)

		path, err := fn(ctx, in.PreviousTxScript.String())
		if err != nil {
			return err
		}

		signer := b.NewSigner()

		pubKey, err := signer.PublicKey(ctx, path)
		if err != nil {
			return err
		}

		pubKeyHashStrFromPriv := hex.EncodeToString(crypto.Hash160(pubKey))
		if pubKeyHashStr == pubKeyHashStrFromPriv {
			if err = tx.Sign(ctx, signer, uint32(i), shf); err != nil {
				return err
			}
		}
	}
	return nil
}
