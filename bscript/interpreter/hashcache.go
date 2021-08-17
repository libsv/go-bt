// Copyright (c) 2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package interpreter

import (
	"sync"

	"github.com/libsv/go-bk/crypto"
	"github.com/libsv/go-bt/v2"
)

// TxSigHashes houses the partial set of sighashes introduced within BIP0143.
// This partial set of sighashes may be re-used within each input across a
// transaction when validating all inputs. As a result, validation complexity
// for SigHashAll can be reduced by a polynomial factor.
type TxSigHashes struct {
	HashPrevOuts [32]byte
	HashSequence [32]byte
	HashOutputs  [32]byte
}

// NewTxSigHashes computes, and returns the cached sighashes of the given
// transaction.
func NewTxSigHashes(tx *bt.Tx) *TxSigHashes {
	return &TxSigHashes{
		HashPrevOuts: sha256dh(tx.PreviousOutHash()),
		HashSequence: sha256dh(tx.SequenceHash()),
		HashOutputs:  sha256dh(tx.OutputsHash(-1)),
	}
}

// HashCache houses a set of partial sighashes keyed by txid. The set of partial
// sighashes are those introduced within BIP0143 by the new more efficient
// sighash digest calculation algorithm. Using this threadsafe shared cache,
// multiple goroutines can safely re-use the pre-computed partial sighashes
// speeding up validation time amongst all inputs found within a block.
type HashCache struct {
	sigHashes map[[32]byte]*TxSigHashes

	sync.RWMutex
}

// NewHashCache returns a new instance of the HashCache given a maximum number
// of entries which may exist within it at anytime.
func NewHashCache(maxSize uint) *HashCache {
	return &HashCache{
		sigHashes: make(map[[32]byte]*TxSigHashes, maxSize),
	}
}

// AddSigHashes computes, then adds the partial sighashes for the passed
// transaction.
func (h *HashCache) AddSigHashes(tx *bt.Tx) {
	h.Lock()
	defer h.Unlock()
	var hash [32]byte
	copy(hash[:], tx.TxIDBytes())
	h.sigHashes[hash] = NewTxSigHashes(tx)
}

// ContainsHashes returns true if the partial sighashes for the passed
// transaction currently exist within the HashCache, and false otherwise.
func (h *HashCache) ContainsHashes(txid []byte) bool {
	h.RLock()
	defer h.RUnlock()

	var hash [32]byte
	copy(hash[:], txid)

	_, found := h.sigHashes[hash]

	return found
}

// GetSigHashes possibly returns the previously cached partial sighashes for
// the passed transaction. This function also returns an additional boolean
// value indicating if the sighashes for the passed transaction were found to
// be present within the HashCache.
func (h *HashCache) GetSigHashes(txid []byte) (*TxSigHashes, bool) {
	h.RLock()
	defer h.RUnlock()

	var hash [32]byte
	copy(hash[:], txid)

	item, found := h.sigHashes[hash]

	return item, found
}

// PurgeSigHashes removes all partial sighashes from the HashCache belonging to
// the passed transaction.
func (h *HashCache) PurgeSigHashes(txid []byte) {
	h.Lock()
	defer h.Unlock()

	var hash [32]byte
	copy(hash[:], txid)

	delete(h.sigHashes, hash)
}

func sha256dh(b []byte) [32]byte {
	var h [32]byte
	copy(h[:], crypto.Sha256d(b))

	return h
}
