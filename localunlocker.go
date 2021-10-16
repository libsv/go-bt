package bt

import (
	"context"
	"errors"

	"github.com/libsv/go-bk/bec"
	"github.com/libsv/go-bt/v2/bscript"
	"github.com/libsv/go-bt/v2/sighash"
)

// LocalUnlockerGetter implements the UnlockerGetter interface. It unlocks a Tx locally,
// using a bec PrivateKey.
type LocalUnlockerGetter struct {
	PrivateKey *bec.PrivateKey
}

// Unlocker builds a new *bt.LocalUnlocker with the same private key
// as the calling *bt.LocalUnlockerGetter.
func (lg *LocalUnlockerGetter) Unlocker(ctx context.Context, lockingScript *bscript.Script) (Unlocker, error) {
	return &LocalUnlocker{PrivateKey: lg.PrivateKey}, nil
}

// LocalUnlocker implements the unlocker interface. It is used to unlock a tx locally using a
// bec Private Key.
type LocalUnlocker struct {
	PrivateKey *bec.PrivateKey
}

// Unlock a transaction at a given input using the PrivateKey passed in through the LocalUnlocker
// struct.
// Unlock generates, applies, and returns an ECDSA signature for the provided hash digest using the private key
// as well as the public key corresponding to the private key used. The produced
// signature is deterministic (same message and same key yield the same signature) and
// canonical in accordance with RFC6979 and BIP0062.
func (lu *LocalUnlocker) Unlock(ctx context.Context, tx *Tx, idx uint32, shf sighash.Flag) error {
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

	switch tx.Inputs[idx].PreviousTxScript.ScriptType() {
	case bscript.ScriptTypePubKeyHash:
		return tx.ApplyP2PKHUnlockingScript(idx, pubKey, signature, shf)
	}

	return errors.New("currently only p2pkh supported")
}
