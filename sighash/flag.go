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

	// Currently, all BitCoin (SV) transactions require an additional SIGHASH flag (after UAHF)

	AllForkID          Flag = 0x1 | 0x40
	NoneForkID         Flag = 0x2 | 0x40
	SingleForkID       Flag = 0x3 | 0x40
	AnyOneCanPayForkID Flag = 0x80 | 0x40

	// ForkID is the replay protected signature hash flag
	// used by the Uahf hardfork.

	ForkID Flag = 0x40

	// Mask defines the number of bits of the hash type which is used
	// to identify which outputs are signed.
	Mask = 0x1f
)

// Has returns true if contains the provided flag.
func (f Flag) Has(shf Flag) bool {
	return f&shf == shf
}

// HasWithMask returns true if contains the provided flag masked
func (f Flag) HasWithMask(shf Flag) bool {
	return f&Mask == shf
}

func (f Flag) String() string {
	switch f {
	case All:
		return "ALL"
	case None:
		return "NONE"
	case Single:
		return "SINGLE"
	case All | AnyOneCanPay:
		return "ALL|ANYONECANPAY"
	case None | AnyOneCanPay:
		return "NONE|ANYONECANPAY"
	case Single | AnyOneCanPay:
		return "SINGLE|ANYONECANPAY"
	case AllForkID:
		return "ALL|FORKID"
	case NoneForkID:
		return "NONE|FORKID"
	case SingleForkID:
		return "SINGLE|FORKID"
	case AllForkID | AnyOneCanPay:
		return "ALL|FORKID|ANYONECANPAY"
	case NoneForkID | AnyOneCanPay:
		return "NONE|FORKID|ANYONECANPAY"
	case SingleForkID | AnyOneCanPay:
		return "SINGLE|FORKID|ANYONECANPAY"
	}

	return "ALL"
}
