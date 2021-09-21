package bt

import (
	"context"
	"errors"

	"github.com/libsv/go-bt/v2/bscript"
	"github.com/libsv/go-bt/v2/sighash"
)

// SignAllBip32 is used to automatically check which P2PKH Inputs are
// able to be signed (match a derivated public key) and then sign them.
// It takes a Bip32SignerDeriver interface as a parameter so that different
// signing implementations can be used to sign the transaction -
// for example internal/local or external signing.
// It also takes Bip32PathGetterFunc with is expected to return a correct derivation,
// path for a given input, to allow matching and signing.
func (tx *Tx) SignAllBip32(ctx context.Context, b Bip32SignerDeriver, get Bip32PathGetterFunc) error {
	if get == nil {
		return errors.New("Bip32PathGetterFunc not provided")
	}
	// TODO: add support for other script types
	signerStrats := map[string]signerFunc{
		bscript.ScriptTypePubKeyHash: tx.Sign,
	}

	shf := sighash.AllForkID

	for i, in := range tx.Inputs {
		fn, ok := signerStrats[in.PreviousTxScript.ScriptType()]
		if !ok {
			return errors.New("unsupported script type")
		}

		path, err := get(ctx, in.PreviousTxScript.String())
		if err != nil {
			return err
		}

		signer, err := b.DeriveBip32Signer(path)
		if err != nil {
			return err
		}

		if err = fn(ctx, signer, uint32(i), shf); err != nil {
			return err
		}
	}

	return nil
}
