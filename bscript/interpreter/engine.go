// Copyright (c) 2013-2018 The btcsuite developers
// Copyright (c) 2015-2018 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package interpreter

import (
	"math/big"

	"github.com/libsv/go-bk/bec"
)

// ScriptFlags is a bitmask defining additional operations or tests that will be
// done when executing a script pair.
type ScriptFlags uint32

const (
	// ScriptBip16 defines whether the bip16 threshold has passed and thus
	// pay-to-script hash transactions will be fully validated.
	ScriptBip16 ScriptFlags = 1 << iota

	// ScriptStrictMultiSig defines whether to verify the stack item
	// used by CHECKMULTISIG is zero length.
	ScriptStrictMultiSig

	// ScriptDiscourageUpgradableNops defines whether to verify that
	// NOP1 through NOP10 are reserved for future soft-fork upgrades.  This
	// flag must not be used for consensus critical code nor applied to
	// blocks as this flag is only for stricter standard transaction
	// checks.  This flag is only applied when the above opcodes are
	// executed.
	ScriptDiscourageUpgradableNops

	// ScriptVerifyCheckLockTimeVerify defines whether to verify that
	// a transaction output is spendable based on the locktime.
	// This is BIP0065.
	ScriptVerifyCheckLockTimeVerify

	// ScriptVerifyCheckSequenceVerify defines whether to allow execution
	// pathways of a script to be restricted based on the age of the output
	// being spent.  This is BIP0112.
	ScriptVerifyCheckSequenceVerify

	// ScriptVerifyCleanStack defines that the stack must contain only
	// one stack element after evaluation and that the element must be
	// true if interpreted as a boolean.  This is rule 6 of BIP0062.
	// This flag should never be used without the ScriptBip16 flag.
	ScriptVerifyCleanStack

	// ScriptVerifyDERSignatures defines that signatures are required
	// to compily with the DER format.
	ScriptVerifyDERSignatures

	// ScriptVerifyLowS defines that signtures are required to comply with
	// the DER format and whose S value is <= order / 2.  This is rule 5
	// of BIP0062.
	ScriptVerifyLowS

	// ScriptVerifyMinimalData defines that signatures must use the smallest
	// push operator. This is both rules 3 and 4 of BIP0062.
	ScriptVerifyMinimalData

	// ScriptVerifyNullFail defines that signatures must be empty if
	// a CHECKSIG or CHECKMULTISIG operation fails.
	ScriptVerifyNullFail

	// ScriptVerifySigPushOnly defines that signature scripts must contain
	// only pushed data.  This is rule 2 of BIP0062.
	ScriptVerifySigPushOnly

	// ScriptEnableSighashForkID defined that signature scripts have forkid
	// enabled.
	ScriptEnableSighashForkID

	// ScriptVerifyStrictEncoding defines that signature scripts and
	// public keys must follow the strict encoding requirements.
	ScriptVerifyStrictEncoding

	// ScriptVerifyBip143SigHash defines that signature hashes should
	// be calculated using the bip0143 signature hashing algorithm.
	ScriptVerifyBip143SigHash

	// ScriptUTXOAfterGenesis defines that the utxo was created after
	// genesis.
	ScriptUTXOAfterGenesis

	// ScriptVerifyMinimalIf defines the enforcement of any conditional statement using the
	// minimum required data.
	ScriptVerifyMinimalIf
)

// HasFlag returns whether the ScriptFlags has the passed flag set.
func (s ScriptFlags) HasFlag(flag ScriptFlags) bool {
	return s&flag == flag
}

// AddFlag adds the passed flag to ScriptFlags
func (s *ScriptFlags) AddFlag(flag ScriptFlags) {
	*s |= flag
}

// halforder is used to tame ECDSA malleability (see BIP0062).
var halfOrder = new(big.Int).Rsh(bec.S256().N, 1)

// Engine is the virtual machine that executes scripts.
type Engine interface {
	Execute(ExecutionParams) error
}

type engine struct{}

// NewEngine returns a new script engine for the provided locking script
// (of a previous transaction out), transaction, and input index.  The
// flags modify the behaviour of the script engine according to the
// description provided by each flag.
func NewEngine() Engine {
	return &engine{}
}

// Execute will execute all scripts in the script engine and return either nil
// for successful validation or an error if one occurred.
func (e *engine) Execute(params ExecutionParams) error {
	th := &thread{
		scriptParser: &parser{},
		cfg:          &beforeGenesisConfig{},
	}

	if err := th.apply(params); err != nil {
		return err
	}

	return th.execute()
}
