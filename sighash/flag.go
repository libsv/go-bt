package sighash

// Flag represents hash type bits at the end of a signature.
type Flag uint8

// SIGHASH type bits from the end of a signature.
// see: https://wiki.bitcoinsv.io/index.php/SIGHASH_flags
const (
	Old          Flag = 0x0
	All          Flag = 0x1
	None         Flag = 0x2
	Single       Flag = 0x3
	AnyOneCanPay Flag = 0x80

	// Currently all BitCoin (SV) transactions require an additional SIGHASH flag (after UAHF)
	AllForkID          Flag = 0x1 | 0x40
	NoneForkID         Flag = 0x2 | 0x40
	SingleForkID       Flag = 0x3 | 0x40
	AnyOneCanPayForkID Flag = 0x80 | 0x40

	// SigHashForkID is the replay protected signature hash flag
	// used by the Uahf hardfork.
	ForkID Flag = 0x40

	// sigHashMask defines the number of bits of the hash type which is used
	// to identify which outputs are signed.
	Mask = 0x1f
)
