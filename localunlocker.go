package bt

import (
	"context"
	"errors"

	"github.com/libsv/go-bk/bec"
	"github.com/libsv/go-bt/v2/bscript"
	"github.com/libsv/go-bt/v2/sighash"
)

// LocalSignatureUnlockerGetter implements the UnlockerGetter interface. It unlocks a Tx locally,
// using a bkec PrivateKey.
type LocalSignatureUnlockerGetter struct {
	PrivateKey *bec.PrivateKey
}

// Unlocker builds a new *bt.LocalSignatureUnlocker with the same private key
// as the calling *bt.LocalSignatureUnlockerGetter.
func (lg *LocalSignatureUnlockerGetter) Unlocker(ctx context.Context, lockingScript *bscript.Script) (Unlocker, error) {
	return &LocalSignatureUnlocker{PrivateKey: lg.PrivateKey}, nil
}

// LocalSignatureUnlocker implements the unlocker interface. It is used to unlock a tx locally using a
// bkec Private Key.
type LocalSignatureUnlocker struct {
	PrivateKey *bec.PrivateKey
}

// Unlock a transaction at a given input using the PrivateKey passed in through the LocalSignatureUnlocker
// struct.
// Unlock generates and applies an ECDSA signature for the provided hash digest using the private key
// as well as the public key corresponding to the private key used. The produced
// signature is deterministic (same message and same key yield the same signature) and
// canonical in accordance with RFC6979 and BIP0062.
func (lu *LocalSignatureUnlocker) Unlock(ctx context.Context, tx *Tx, idx uint32, shf sighash.Flag) error {
	if shf == 0 {
		shf = sighash.AllForkID
	}

	sh, err := tx.CalcInputSignatureHash(idx, shf)
	if err != nil {
		return err
	}

	sig, err := lu.PrivateKey.Sign(sh)
	if err != nil {
		return err
	}

	pubKey := lu.PrivateKey.PubKey().SerialiseCompressed()
	signature := sig.Serialise()

	// TODO: support more script types
	switch tx.Inputs[idx].PreviousTxScript.ScriptType() {
	case bscript.ScriptTypePubKeyHash:
		return tx.ApplyP2PKHUnlockingScript(idx, pubKey, signature, shf)
	}

	return errors.New("currently only p2pkh supported")
}
