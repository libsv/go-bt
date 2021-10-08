package bt

import (
	"context"

	"github.com/libsv/go-bk/bec"
	"github.com/libsv/go-bt/v2/bscript"
	"github.com/libsv/go-bt/v2/sighash"
)

// LocalP2PKHUnlockerGetter implements the UnlockerGetter interface. It unlocks a Tx locally,
// using a bkec PrivateKey.
type LocalP2PKHUnlockerGetter struct {
	PrivateKey *bec.PrivateKey
}

// Unlocker builds a new *bt.LocalP2PKHUnlocker with the same private key as the calling *bt.LocalP2PKHUnlockerGetter.
func (lg *LocalP2PKHUnlockerGetter) Unlocker(ctx context.Context, lockingScript *bscript.Script) (Unlocker, error) {
	return &LocalP2PKHUnlocker{PrivateKey: lg.PrivateKey}, nil
}

// LocalP2PKHUnlocker implements the unlocker interface. It is used to unlock a tx locally using a
// bkec Private Key.
type LocalP2PKHUnlocker struct {
	PrivateKey *bec.PrivateKey
}

// Unlock a transaction at a given input using the PrivateKey passed in through the LocalP2PKHUnlocker
// struct.
// Unlock generates and applies an ECDSA signature for the provided hash digest using the private key
// as well as the public key corresponding to the private key used. The produced
// signature is deterministic (same message and same key yield the same signature) and
// canonical in accordance with RFC6979 and BIP0062.
func (lu *LocalP2PKHUnlocker) Unlock(ctx context.Context, tx *Tx, idx uint32, shf sighash.Flag) error {
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

	return tx.ApplyP2PKHUnlockingScript(idx, pubKey, signature, shf)
}
