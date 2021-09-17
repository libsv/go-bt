package bt

import (
	"context"
	"encoding/hex"

	"github.com/libsv/go-bk/crypto"
	"github.com/libsv/go-bt/v2/sighash"
)

func (tx *Tx) Bip32SignAuto(ctx context.Context, b Bip32SignerDeriver, fn Bip32PathGetterFunc) (inputsSigned []int, err error) {
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
