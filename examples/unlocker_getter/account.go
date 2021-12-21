package main

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"errors"

	"github.com/libsv/go-bk/bip32"
	"github.com/libsv/go-bk/chaincfg"
	"github.com/libsv/go-bt/v2"
	"github.com/libsv/go-bt/v2/bscript"
	"github.com/libsv/go-bt/v2/unlocker"
)

// account for creating destination scripts and stores these scripts with their derivations.
type account struct {
	// masterPrivKey of which all locking scripts and private keys are generated from.
	masterPrivKey *bip32.ExtendedKey
	// counter for deriving paths for public/private keys
	counter uint64
	// scriptToPathMap for mapping a locking script hex to its derivation path.
	scriptToPathMap map[string]string
}

func newAccount() *account {
	// Generate the master private key.
	seed, err := bip32.GenerateSeed(bip32.RecommendedSeedLen)
	if err != nil {
		panic(err)
	}
	privKey, err := bip32.NewMaster(seed, &chaincfg.MainNet)
	if err != nil {
		panic(err)
	}

	return &account{
		masterPrivKey:   privKey,
		scriptToPathMap: make(map[string]string, 0),
	}
}

func (a *account) createDestination() *bscript.Script {
	// generate a new path until and increment the counter.
	path := bip32.DerivePath(a.counter)
	a.counter++

	// Derive a public key from the new derivation path.
	pubKey, err := a.masterPrivKey.DerivePublicKeyFromPath(path)
	if err != nil {
		panic(err)
	}

	// Create a locking script from this public key.
	s, err := bscript.NewP2PKHFromPubKeyBytes(pubKey)
	if err != nil {
		panic(err)
	}

	// Store the locking script and its path for later use.
	a.scriptToPathMap[s.String()] = path

	return s
}

// Unlocker get the correct unlocker for a given locking script.
func (a *account) Unlocker(ctx context.Context, lockingScript *bscript.Script) (bt.Unlocker, error) {
	// Retrieve the path for the given locking script.
	path, ok := a.scriptToPathMap[lockingScript.String()]
	if !ok {
		panic(errors.New("oh no"))
	}

	// Derive a private key from the stored derivation path. This private key will be pair to
	// the public key we derived when this locking script was created.
	extPrivKey, err := a.masterPrivKey.DeriveChildFromPath(path)
	if err != nil {
		panic(err)
	}

	// Convert the extended key into an eliptic curve private key.
	privKey, err := extPrivKey.ECPrivKey()
	if err != nil {
		panic(err)
	}

	return &unlocker.Simple{
		PrivateKey: privKey,
	}, nil
}

func randUint64() uint64 {
	var bb [8]byte
	if _, err := rand.Read(bb[:]); err != nil {
		panic(err)
	}

	return binary.LittleEndian.Uint64(bb[:])
}
