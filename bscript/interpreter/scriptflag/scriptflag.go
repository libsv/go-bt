// Package scriptflag comment
package scriptflag

// Flag is a bitmask defining additional operations or tests that will be
// done when executing a script pair.
type Flag uint32

const (
	// Bip16 defines whether the bip16 threshold has passed and thus
	// pay-to-script hash transactions will be fully validated.
	Bip16 Flag = 1 << iota

	// StrictMultiSig defines whether to verify the stack item
	// used by CHECKMULTISIG is zero length.
	StrictMultiSig

	// DiscourageUpgradableNops defines whether to verify that
	// NOP1 through NOP10 are reserved for future soft-fork upgrades.  This
	// flag must not be used for consensus critical code nor applied to
	// blocks as this flag is only for stricter standard transaction
	// checks.  This flag is only applied when the above opcodes are
	// executed.
	DiscourageUpgradableNops

	// VerifyCheckLockTimeVerify defines whether to verify that
	// a transaction output is spendable based on the locktime.
	// This is BIP0065.
	VerifyCheckLockTimeVerify

	// VerifyCheckSequenceVerify defines whether to allow execution
	// pathways of a script to be restricted based on the age of the output
	// being spent.  This is BIP0112.
	VerifyCheckSequenceVerify

	// VerifyCleanStack defines that the stack must contain only
	// one stack element after evaluation and that the element must be
	// true if interpreted as a boolean.  This is rule 6 of BIP0062.
	// This flag should never be used without the Bip16 flag.
	VerifyCleanStack

	// VerifyDERSignatures defines that signatures are required
	// to comply with the DER format.
	VerifyDERSignatures

	// VerifyLowS defines that signatures are required to comply with
	// the DER format and whose S value is <= order / 2.  This is rule 5
	// of BIP0062.
	VerifyLowS

	// VerifyMinimalData defines that signatures must use the smallest
	// push operator. This is both rules 3 and 4 of BIP0062.
	VerifyMinimalData

	// VerifyNullFail defines that signatures must be empty if
	// a CHECKSIG or CHECKMULTISIG operation fails.
	VerifyNullFail

	// VerifySigPushOnly defines that signature scripts must contain
	// only pushed data.  This is rule 2 of BIP0062.
	VerifySigPushOnly

	// EnableSighashForkID defined that signature scripts have forkid
	// enabled.
	EnableSighashForkID

	// VerifyStrictEncoding defines that signature scripts and
	// public keys must follow the strict encoding requirements.
	VerifyStrictEncoding

	// VerifyBip143SigHash defines that signature hashes should
	// be calculated using the bip0143 signature hashing algorithm.
	VerifyBip143SigHash

	// UTXOAfterGenesis defines that the utxo was created after
	// genesis.
	UTXOAfterGenesis

	// VerifyMinimalIf defines the enforcement of any conditional statement using the
	// minimum required data.
	VerifyMinimalIf
)

// HasFlag returns whether the Flags has the passed flag set.
func (s Flag) HasFlag(flag Flag) bool {
	return s&flag == flag
}

// HasAny returns true if any of the passed in flags are present.
func (s Flag) HasAny(flags ...Flag) bool {
	for _, f := range flags {
		if s&f == f {
			return true
		}
	}

	return false
}

// AddFlag adds the passed flag to Flags
func (s *Flag) AddFlag(flag Flag) {
	*s |= flag
}
