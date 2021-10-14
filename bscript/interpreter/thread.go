package interpreter

import (
	"math/big"

	"github.com/libsv/go-bk/bec"
	"github.com/libsv/go-bt/v2"
	"github.com/libsv/go-bt/v2/bscript"
	"github.com/libsv/go-bt/v2/bscript/interpreter/errs"
	"github.com/libsv/go-bt/v2/bscript/interpreter/scriptflag"
	"github.com/libsv/go-bt/v2/sighash"
)

// halfOrder is used to tame ECDSA malleability (see BIP0062).
var halfOrder = new(big.Int).Rsh(bec.S256().N, 1)

type thread struct {
	dstack stack // data stack
	astack stack // alt stack

	elseStack boolStack

	cfg config

	debug *Debugger

	scripts         []ParsedScript
	condStack       []int
	savedFirstStack [][]byte // stack from first script for bip16 scripts

	scriptParser OpcodeParser
	scriptIdx    int
	scriptOff    int
	lastCodeSep  int

	tx         *bt.Tx
	inputIdx   int
	prevOutput *bt.Output

	numOps int

	flags scriptflag.Flag
	bip16 bool // treat execution as pay-to-script-hash

	afterGenesis            bool
	earlyReturnAfterGenesis bool
}

func createThread(opts *execOpts) (*thread, error) {
	th := &thread{
		scriptParser: &DefaultOpcodeParser{
			ErrorOnCheckSig: opts.Tx == nil || opts.PreviousTxOut == nil,
		},
		debug: NewDebugger(),
		cfg:   &beforeGenesisConfig{},
	}

	if err := th.apply(opts); err != nil {
		return nil, err
	}

	return th, nil
}

// execOpts are the params required for building an Engine
//
// Raw *bscript.Scripts can be supplied as LockingScript and UnlockingScript, or
// a Tx, an input index, and a previous output.
//
// If checksig operaitons are to be executed without a Tx or a PreviousTxOut supplied,
// the engine will return an ErrInvalidParams on execute.
type execOpts struct {
	LockingScript   *bscript.Script
	UnlockingScript *bscript.Script
	PreviousTxOut   *bt.Output
	Tx              *bt.Tx
	InputIdx        int
	Flags           scriptflag.Flag
	Debugger        *Debugger
}

func (o execOpts) validate() error {
	// The provided transaction input index must refer to a valid input.
	if o.InputIdx < 0 || (o.Tx != nil && o.InputIdx > o.Tx.InputCount()-1) {
		return errs.NewError(
			errs.ErrInvalidIndex,
			"transaction input index %d is negative or >= %d", o.InputIdx, len(o.Tx.Inputs),
		)
	}

	outputHasLockingScript := o.PreviousTxOut != nil && o.PreviousTxOut.LockingScript != nil
	txHasUnlockingScript := o.Tx != nil && o.Tx.Inputs != nil && len(o.Tx.Inputs) > 0 &&
		o.Tx.Inputs[o.InputIdx] != nil && o.Tx.Inputs[o.InputIdx].UnlockingScript != nil
	// If no locking script was provided
	if o.LockingScript == nil && !outputHasLockingScript {
		return errs.NewError(errs.ErrInvalidParams, "no locking script provided")
	}

	// If no unlocking script was provided
	if o.UnlockingScript == nil && !txHasUnlockingScript {
		return errs.NewError(errs.ErrInvalidParams, "no unlocking script provided")
	}

	// If both a locking script and previous output were provided, make sure the scripts match
	if o.LockingScript != nil && outputHasLockingScript {
		if !o.LockingScript.Equals(o.PreviousTxOut.LockingScript) {
			return errs.NewError(
				errs.ErrInvalidParams,
				"locking script does not match the previous outputs locking script",
			)
		}
	}

	// If both a unlocking script and an input were provided, make sure the scripts match
	if o.UnlockingScript != nil && txHasUnlockingScript {
		if !o.UnlockingScript.Equals(o.Tx.Inputs[o.InputIdx].UnlockingScript) {
			return errs.NewError(
				errs.ErrInvalidParams,
				"unlocking script does not match the unlocking script of the requested input",
			)
		}
	}

	return nil
}

// hasFlag returns whether the script engine instance has the passed flag set.
func (t *thread) hasFlag(flag scriptflag.Flag) bool {
	return t.flags.HasFlag(flag)
}

func (t *thread) hasAny(ff ...scriptflag.Flag) bool {
	return t.flags.HasAny(ff...)
}

func (t *thread) addFlag(flag scriptflag.Flag) {
	t.flags.AddFlag(flag)
}

// isBranchExecuting returns whether the current conditional branch is
// actively executing. For example, when the data stack has an OP_FALSE on it
// and an OP_IF is encountered, the branch is inactive until an OP_ELSE or
// OP_ENDIF is encountered.  It properly handles nested conditionals.
func (t *thread) isBranchExecuting() bool {
	return len(t.condStack) == 0 || t.condStack[len(t.condStack)-1] == opCondTrue
}

// executeOpcode performs execution on the passed opcode. It takes into account
// whether it is hidden by conditionals, but some rules still must be
// tested in this case.
func (t *thread) executeOpcode(pop ParsedOpcode) error {
	if len(pop.Data) > t.cfg.MaxScriptElementSize() {
		return errs.NewError(errs.ErrElementTooBig,
			"element size %d exceeds max allowed size %d", len(pop.Data), t.cfg.MaxScriptElementSize())
	}

	exec := t.shouldExec(pop)

	// Disabled opcodes are fail on program counter.
	if pop.IsDisabled() && (!t.afterGenesis || exec) {
		return errs.NewError(errs.ErrDisabledOpcode, "attempt to execute disabled opcode %s", pop.Name())
	}

	// Always-illegal opcodes are fail on program counter.
	if pop.AlwaysIllegal() && !t.afterGenesis {
		return errs.NewError(errs.ErrReservedOpcode, "attempt to execute reserved opcode %s", pop.Name())
	}

	// Note that this includes OP_RESERVED which counts as a push operation.
	if pop.op.val > bscript.Op16 {
		t.numOps++
		if t.numOps > t.cfg.MaxOps() {
			return errs.NewError(errs.ErrTooManyOperations, "exceeded max operation limit of %d", t.cfg.MaxOps())
		}

	}

	if len(pop.Data) > t.cfg.MaxScriptElementSize() {
		return errs.NewError(errs.ErrElementTooBig,
			"element size %d exceeds max allowed size %d", len(pop.Data), t.cfg.MaxScriptElementSize())
	}

	// Nothing left to do when this is not a conditional opcode, and it is
	// not in an executing branch.
	if !t.isBranchExecuting() && !pop.IsConditional() {
		return nil
	}

	// Ensure all executed data push opcodes use the minimal encoding when
	// the minimal data verification flag is set.
	if t.dstack.verifyMinimalData && t.isBranchExecuting() && pop.op.val <= bscript.OpPUSHDATA4 && exec {
		if err := pop.enforceMinimumDataPush(); err != nil {
			return err
		}
	}

	// If we have already reached an OP_RETURN, we don't execute the next comment, unless it is a conditional,
	// in which case we need to evaluate it as to check for correct if/else balances
	if !exec && !pop.IsConditional() {
		return nil
	}

	return pop.op.exec(&pop, t)
}

// validPC returns an error if the current script position is valid for
// execution, nil otherwise.
func (t *thread) validPC() error {
	if t.scriptIdx >= len(t.scripts) {
		return errs.NewError(errs.ErrInvalidProgramCounter,
			"past input scripts %v:%v %v:xxxx", t.scriptIdx, t.scriptOff, len(t.scripts))
	}
	if t.scriptOff >= len(t.scripts[t.scriptIdx]) {
		return errs.NewError(errs.ErrInvalidProgramCounter, "past input scripts %v:%v %v:%04d", t.scriptIdx, t.scriptOff,
			t.scriptIdx, len(t.scripts[t.scriptIdx]))
	}
	return nil
}

// CheckErrorCondition returns nil if the running script has ended and was
// successful, leaving a true boolean on the stack.  An error otherwise,
// including if the script has not finished.
func (t *thread) CheckErrorCondition(finalScript bool) error {
	if t.dstack.Depth() < 1 {
		return errs.NewError(errs.ErrEmptyStack, "stack empty at end of script execution")
	}

	if finalScript && t.hasFlag(scriptflag.VerifyCleanStack) && t.dstack.Depth() != 1 {
		return errs.NewError(errs.ErrCleanStack, "stack contains %d unexpected items", t.dstack.Depth()-1)
	}

	v, err := t.dstack.PopBool()
	if err != nil {
		return err
	}
	if !v {
		return errs.NewError(errs.ErrEvalFalse, "false stack entry at end of script execution")
	}

	if finalScript {
		t.debug.afterSuccess()
	}
	return nil
}

// Step will execute the next instruction and move the program counter to the
// next opcode in the script, or the next script if the current has ended.  Step
// will return true in the case that the last opcode was successfully executed.
//
// The result of calling Step or any other method is undefined if an error is
// returned.
func (t *thread) Step() (bool, error) {
	// Verify that it is pointing to a valid script address.
	if err := t.validPC(); err != nil {
		return true, err
	}

	opcode := t.scripts[t.scriptIdx][t.scriptOff]

	t.debug.beforeExecuteOpcode()
	// Execute the opcode while taking into account several things such as
	// disabled opcodes, illegal opcodes, maximum allowed operations per
	// script, maximum script element sizes, and conditionals.
	if err := t.executeOpcode(opcode); err != nil {
		if ok := errs.IsErrorCode(err, errs.ErrOK); ok {
			// If returned early, move onto the next script
			t.shiftScript()
			return t.scriptIdx >= len(t.scripts), nil
		}
		return true, err
	}
	t.debug.afterExecuteOpcode()

	t.scriptOff++

	// The number of elements in the combination of the data and alt stacks
	// must not exceed the maximum number of stack elements allowed.
	combinedStackSize := t.dstack.Depth() + t.astack.Depth()
	if combinedStackSize > int32(t.cfg.MaxStackSize()) {
		return false, errs.NewError(errs.ErrStackOverflow,
			"combined stack size %d > max allowed %d", combinedStackSize, t.cfg.MaxStackSize())
	}

	if t.scriptOff < len(t.scripts[t.scriptIdx]) {
		return false, nil
	}

	// Prepare for next instruction.
	// Illegal to have an `if' that straddles two scripts.
	if len(t.condStack) != 0 {
		return false, errs.NewError(errs.ErrUnbalancedConditional, "end of script reached in conditional execution")
	}

	// Alt stack doesn't persist.
	_ = t.astack.DropN(t.astack.Depth())

	// Move onto the next script
	t.shiftScript()

	if t.bip16 && !t.afterGenesis && t.scriptIdx <= 2 {
		switch t.scriptIdx {
		case 1:
			t.savedFirstStack = t.GetStack()
		case 2:
			// Put us past the end for CheckErrorCondition()
			// Check script ran successfully and pull the script
			// out of the first stack and execute that.
			if err := t.CheckErrorCondition(false); err != nil {
				return false, err
			}

			script := t.savedFirstStack[len(t.savedFirstStack)-1]
			pops, err := t.scriptParser.Parse(bscript.NewFromBytes(script))
			if err != nil {
				return false, err
			}

			t.scripts = append(t.scripts, pops)

			// Set stack to be the stack from first script minus the
			// script itself
			t.SetStack(t.savedFirstStack[:len(t.savedFirstStack)-1])
		}
	}

	// there are zero length scripts in the wild
	if t.scriptIdx < len(t.scripts) && t.scriptOff >= len(t.scripts[t.scriptIdx]) {
		t.scriptIdx++
	}

	t.lastCodeSep = 0
	if t.scriptIdx >= len(t.scripts) {
		return true, nil
	}

	return false, nil
}

func (t *thread) apply(opts *execOpts) error {
	if err := opts.validate(); err != nil {
		return err
	}

	if opts.Debugger != nil {
		t.debug = opts.Debugger
		t.debug.attach(t)
	}
	t.dstack.debug = t.debug
	t.astack.debug = t.debug

	if opts.UnlockingScript == nil {
		opts.UnlockingScript = opts.Tx.Inputs[opts.InputIdx].UnlockingScript
	}
	if opts.LockingScript == nil {
		opts.LockingScript = opts.PreviousTxOut.LockingScript
	}

	t.tx = opts.Tx
	t.flags = opts.Flags
	t.inputIdx = opts.InputIdx
	t.prevOutput = opts.PreviousTxOut

	// The clean stack flag (ScriptVerifyCleanStack) is not allowed without
	// the pay-to-script-hash (P2SH) evaluation (ScriptBip16).
	//
	// Recall that evaluating a P2SH script without the flag set results in
	// non-P2SH evaluation which leaves the P2SH inputs on the stack.
	// Thus, allowing the clean stack flag without the P2SH flag would make
	// it possible to have a situation where P2SH would not be a soft fork
	// when it should be.
	if t.hasFlag(scriptflag.EnableSighashForkID) {
		t.addFlag(scriptflag.VerifyStrictEncoding)
	}

	t.elseStack = &nopBoolStack{}
	if t.hasFlag(scriptflag.UTXOAfterGenesis) {
		t.elseStack = &stack{debug: NewDebugger()}
		t.afterGenesis = true
		t.cfg = &afterGenesisConfig{}
	}

	uls := opts.UnlockingScript
	ls := opts.LockingScript

	// When both the signature script and public key script are empty the
	// result is necessarily an error since the stack would end up being
	// empty which is equivalent to a false top element.  Thus, just return
	// the relevant error now as an optimization.
	if (uls == nil || len(*uls) == 0) && (ls == nil || len(*ls) == 0) {
		return errs.NewError(errs.ErrEvalFalse, "false stack entry at end of script execution")
	}

	if t.hasFlag(scriptflag.VerifyCleanStack) && !t.hasFlag(scriptflag.Bip16) {
		return errs.NewError(errs.ErrInvalidFlags, "invalid scriptflag combination")
	}

	if len(*uls) > t.cfg.MaxScriptSize() {
		return errs.NewError(
			errs.ErrScriptTooBig,
			"unlocking script size %d is larger than the max allowed size %d",
			len(*uls),
			t.cfg.MaxScriptSize(),
		)
	}
	if len(*ls) > t.cfg.MaxScriptSize() {
		return errs.NewError(
			errs.ErrScriptTooBig,
			"locking script size %d is larger than the max allowed size %d",
			len(*uls),
			t.cfg.MaxScriptSize(),
		)
	}

	// The engine stores the scripts in parsed form using a slice.  This
	// allows multiple scripts to be executed in sequence.  For example,
	// with a pay-to-script-hash transaction, there will be ultimately be
	// a third script to execute.
	t.scripts = make([]ParsedScript, 2)
	for i, script := range []*bscript.Script{uls, ls} {
		pscript, err := t.scriptParser.Parse(script)
		if err != nil {
			return err
		}

		t.scripts[i] = pscript
	}

	// The signature script must only contain data pushes when the
	// associated flag is set.
	if t.hasFlag(scriptflag.VerifySigPushOnly) && !t.scripts[0].IsPushOnly() {
		return errs.NewError(errs.ErrNotPushOnly, "signature script is not push only")
	}

	// Advance the program counter to the public key script if the signature
	// script is empty since there is nothing to execute for it in that
	// case.
	if len(*uls) == 0 {
		t.scriptIdx++
	}

	if t.hasFlag(scriptflag.Bip16) && ls.IsP2SH() {
		// Only accept input scripts that push data for P2SH.
		if !t.scripts[0].IsPushOnly() {
			return errs.NewError(errs.ErrNotPushOnly, "pay to script hash is not push only")
		}
		t.bip16 = true
	}

	if t.hasFlag(scriptflag.VerifyMinimalData) {
		t.dstack.verifyMinimalData = true
		t.astack.verifyMinimalData = true
	}

	if t.tx != nil {
		t.tx.InputIdx(t.inputIdx).PreviousTxScript = t.prevOutput.LockingScript
		t.tx.InputIdx(t.inputIdx).PreviousTxSatoshis = t.prevOutput.Satoshis
	}

	return nil
}

func (t *thread) execute() error {
	if err := func() error {
		defer t.debug.afterExecution()
		for {
			done, err := t.Step()
			if err != nil {
				return err
			}

			if done {
				return nil
			}
		}
	}(); err != nil {
		return err
	}

	return t.CheckErrorCondition(true)
}

// GetStack returns the contents of the primary stack as an array. where the
// last item in the array is the top of the stack.
func (t *thread) GetStack() [][]byte {
	return getStack(&t.dstack)
}

// SetStack sets the contents of the primary stack to the contents of the
// provided array where the last item in the array will be the top of the stack.
func (t *thread) SetStack(data [][]byte) {
	setStack(&t.dstack, data)
}

// subScript returns the script since the last OP_CODESEPARATOR.
func (t *thread) subScript() ParsedScript {
	return t.scripts[t.scriptIdx][t.lastCodeSep:]
}

// checkHashTypeEncoding returns whether the passed hashtype adheres to
// the strict encoding requirements if enabled.
func (t *thread) checkHashTypeEncoding(shf sighash.Flag) error {
	if !t.hasFlag(scriptflag.VerifyStrictEncoding) {
		return nil
	}

	sigHashType := shf & ^sighash.AnyOneCanPay
	if t.hasFlag(scriptflag.VerifyBip143SigHash) {
		sigHashType ^= sighash.ForkID
		if shf&sighash.ForkID == 0 {
			return errs.NewError(errs.ErrInvalidSigHashType, "hash type does not contain uahf forkID 0x%x", shf)
		}
	}

	if !sigHashType.Has(sighash.ForkID) {
		if sigHashType < sighash.All || sigHashType > sighash.Single {
			return errs.NewError(errs.ErrInvalidSigHashType, "invalid hash type 0x%x", shf)
		}
		return nil
	}

	if sigHashType < sighash.AllForkID || sigHashType > sighash.SingleForkID {
		return errs.NewError(errs.ErrInvalidSigHashType, "invalid hash type 0x%x", shf)
	}

	if !t.hasFlag(scriptflag.EnableSighashForkID) && shf.Has(sighash.ForkID) {
		return errs.NewError(errs.ErrIllegalForkID, "fork id sighash set without flag")
	}
	if t.hasFlag(scriptflag.EnableSighashForkID) && !shf.Has(sighash.ForkID) {
		return errs.NewError(errs.ErrIllegalForkID, "fork id sighash not set with flag")
	}

	return nil
}

// checkPubKeyEncoding returns whether the passed public key adheres to
// the strict encoding requirements if enabled.
func (t *thread) checkPubKeyEncoding(pubKey []byte) error {
	if !t.hasFlag(scriptflag.VerifyStrictEncoding) {
		return nil
	}

	if len(pubKey) == 33 && (pubKey[0] == 0x02 || pubKey[0] == 0x03) {
		// Compressed
		return nil
	}
	if len(pubKey) == 65 && pubKey[0] == 0x04 {
		// Uncompressed
		return nil
	}

	return errs.NewError(errs.ErrPubKeyType, "unsupported public key type")
}

// checkSignatureEncoding returns whether the passed signature adheres to
// the strict encoding requirements if enabled.
func (t *thread) checkSignatureEncoding(sig []byte) error {
	if !t.hasAny(scriptflag.VerifyDERSignatures, scriptflag.VerifyLowS, scriptflag.VerifyStrictEncoding) {
		return nil
	}

	// The format of a DER encoded signature is as follows:
	//
	// 0x30 <total length> 0x02 <length of R> <R> 0x02 <length of S> <S>
	//   - 0x30 is the ASN.1 identifier for a sequence
	//   - Total length is 1 byte and specifies length of all remaining data
	//   - 0x02 is the ASN.1 identifier that specifies an integer follows
	//   - Length of R is 1 byte and specifies how many bytes R occupies
	//   - R is the arbitrary length big-endian encoded number which
	//     represents the R value of the signature.  DER encoding dictates
	//     that the value must be encoded using the minimum possible number
	//     of bytes.  This implies the first byte can only be null if the
	//     highest bit of the next byte is set in order to prevent it from
	//     being interpreted as a negative number.
	//   - 0x02 is once again the ASN.1 integer identifier
	//   - Length of S is 1 byte and specifies how many bytes S occupies
	//   - S is the arbitrary length big-endian encoded number which
	//     represents the S value of the signature.  The encoding rules are
	//     identical as those for R.
	const (
		asn1SequenceID = 0x30
		asn1IntegerID  = 0x02

		// minSigLen is the minimum length of a DER encoded signature and is
		// when both R and S are 1 byte each.
		//
		// 0x30 + <1-byte> + 0x02 + 0x01 + <byte> + 0x2 + 0x01 + <byte>
		minSigLen = 8

		// maxSigLen is the maximum length of a DER encoded signature and is
		// when both R and S are 33 bytes each.  It is 33 bytes because a
		// 256-bit integer requires 32 bytes and an additional leading null byte
		// might be required if the high bit is set in the value.
		//
		// 0x30 + <1-byte> + 0x02 + 0x21 + <33 bytes> + 0x2 + 0x21 + <33 bytes>
		maxSigLen = 72

		// sequenceOffset is the byte offset within the signature of the
		// expected ASN.1 sequence identifier.
		sequenceOffset = 0

		// dataLenOffset is the byte offset within the signature of the expected
		// total length of all remaining data in the signature.
		dataLenOffset = 1

		// rTypeOffset is the byte offset within the signature of the ASN.1
		// identifier for R and is expected to indicate an ASN.1 integer.
		rTypeOffset = 2

		// rLenOffset is the byte offset within the signature of the length of
		// R.
		rLenOffset = 3

		// rOffset is the byte offset within the signature of R.
		rOffset = 4
	)

	// The signature must adhere to the minimum and maximum allowed length.
	sigLen := len(sig)
	if sigLen < minSigLen {
		return errs.NewError(errs.ErrSigTooShort, "malformed signature: too short: %d < %d", sigLen, minSigLen)
	}
	if sigLen > maxSigLen {
		return errs.NewError(errs.ErrSigTooLong, "malformed signature: too long: %d > %d", sigLen, maxSigLen)
	}

	// The signature must start with the ASN.1 sequence identifier.
	if sig[sequenceOffset] != asn1SequenceID {
		return errs.NewError(errs.ErrSigInvalidSeqID, "malformed signature: format has wrong type: %#x", sig[sequenceOffset])
	}

	// The signature must indicate the correct amount of data for all elements
	// related to R and S.
	if int(sig[dataLenOffset]) != sigLen-2 {
		return errs.NewError(errs.ErrSigInvalidDataLen,
			"malformed signature: bad length: %d != %d",
			sig[dataLenOffset], sigLen-2,
		)
	}

	// Calculate the offsets of the elements related to S and ensure S is inside
	// the signature.
	//
	// rLen specifies the length of the big-endian encoded number which
	// represents the R value of the signature.
	//
	// sTypeOffset is the offset of the ASN.1 identifier for S and, like its R
	// counterpart, is expected to indicate an ASN.1 integer.
	//
	// sLenOffset and sOffset are the byte offsets within the signature of the
	// length of S and S itself, respectively.
	rLen := int(sig[rLenOffset])
	sTypeOffset := rOffset + rLen
	sLenOffset := sTypeOffset + 1
	if sTypeOffset >= sigLen {
		return errs.NewError(errs.ErrSigMissingSTypeID, "malformed signature: S type indicator missing")
	}
	if sLenOffset >= sigLen {
		return errs.NewError(errs.ErrSigMissingSLen, "malformed signature: S length missing")
	}

	// The lengths of R and S must match the overall length of the signature.
	//
	// sLen specifies the length of the big-endian encoded number which
	// represents the S value of the signature.
	sOffset := sLenOffset + 1
	sLen := int(sig[sLenOffset])
	if sOffset+sLen != sigLen {
		return errs.NewError(errs.ErrSigInvalidSLen, "malformed signature: invalid S length")
	}

	// R elements must be ASN.1 integers.
	if sig[rTypeOffset] != asn1IntegerID {
		return errs.NewError(errs.ErrSigInvalidRIntID,
			"malformed signature: R integer marker: %#x != %#x", sig[rTypeOffset], asn1IntegerID)
	}

	// Zero-length integers are not allowed for R.
	if rLen == 0 {
		return errs.NewError(errs.ErrSigZeroRLen, "malformed signature: R length is zero")
	}

	// R must not be negative.
	if sig[rOffset]&0x80 != 0 {
		return errs.NewError(errs.ErrSigNegativeR, "malformed signature: R is negative")
	}

	// Null bytes at the start of R are not allowed, unless R would otherwise be
	// interpreted as a negative number.
	if rLen > 1 && sig[rOffset] == 0x00 && sig[rOffset+1]&0x80 == 0 {
		return errs.NewError(errs.ErrSigTooMuchRPadding, "malformed signature: R value has too much padding")
	}

	// S elements must be ASN.1 integers.
	if sig[sTypeOffset] != asn1IntegerID {
		return errs.NewError(errs.ErrSigInvalidSIntID,
			"malformed signature: S integer marker: %#x != %#x", sig[sTypeOffset], asn1IntegerID)
	}

	// Zero-length integers are not allowed for S.
	if sLen == 0 {
		return errs.NewError(errs.ErrSigZeroSLen, "malformed signature: S length is zero")
	}

	// S must not be negative.
	if sig[sOffset]&0x80 != 0 {
		return errs.NewError(errs.ErrSigNegativeS, "malformed signature: S is negative")
	}

	// Null bytes at the start of S are not allowed, unless S would otherwise be
	// interpreted as a negative number.
	if sLen > 1 && sig[sOffset] == 0x00 && sig[sOffset+1]&0x80 == 0 {
		return errs.NewError(errs.ErrSigTooMuchSPadding, "malformed signature: S value has too much padding")
	}

	// Verify the S value is <= half the order of the curve.  This check is done
	// because when it is higher, the complement modulo the order can be used
	// instead which is a shorter encoding by 1 byte.  Further, without
	// enforcing this, it is possible to replace a signature in a valid
	// transaction with the complement while still being a valid signature that
	// verifies.  This would result in changing the transaction hash and thus is
	// a source of malleability.
	if t.hasFlag(scriptflag.VerifyLowS) {
		sValue := new(big.Int).SetBytes(sig[sOffset : sOffset+sLen])
		if sValue.Cmp(halfOrder) > 0 {
			return errs.NewError(errs.ErrSigHighS, "signature is not canonical due to unnecessarily high S value")
		}
	}
	return nil
}

// getStack returns the contents of stack as a byte array bottom up
func getStack(stack *stack) [][]byte {
	array := make([][]byte, stack.Depth())
	for i := range array {
		// PeekByteArray can't fail due to overflow, already checked
		array[len(array)-i-1], _ = stack.PeekByteArray(int32(i))
	}
	return array
}

// setStack sets the stack to the contents of the array where the last item in
// the array is the top item in the stack.
func setStack(stack *stack, data [][]byte) {
	// This can not error. Only errors are for invalid arguments.
	_ = stack.DropN(stack.Depth())

	for i := range data {
		stack.PushByteArray(data[i])
	}
}

// shouldExec returns true if the engine should execute the passed in operation,
// based on its own internal state.
func (t *thread) shouldExec(pop ParsedOpcode) bool {
	if !t.afterGenesis {
		return true
	}
	cf := true
	for _, v := range t.condStack {
		if v == opCondFalse {
			cf = false
			break
		}
	}

	return cf && (!t.earlyReturnAfterGenesis || pop.op.val == bscript.OpRETURN)
}

func (t *thread) shiftScript() {
	t.numOps = 0
	t.scriptOff = 0
	t.scriptIdx++
	t.earlyReturnAfterGenesis = false
}

func (t *thread) state() *ThreadState {
	scriptIdx := t.scriptIdx
	offsetIdx := t.scriptOff
	if scriptIdx >= len(t.scripts) {
		scriptIdx = len(t.scripts) - 1
		offsetIdx = len(t.scripts[scriptIdx]) - 1
	}

	if offsetIdx >= len(t.scripts[scriptIdx]) {
		offsetIdx = len(t.scripts[scriptIdx]) - 1
	}
	ts := ThreadState{
		DStack:    make([][]byte, int(t.dstack.Depth())),
		AStack:    make([][]byte, int(t.astack.Depth())),
		Opcode:    t.scripts[scriptIdx][offsetIdx],
		Scripts:   make([]ParsedScript, len(t.scripts)),
		ScriptIdx: scriptIdx,
		OpcodeIdx: offsetIdx,
	}

	for i, dd := range t.dstack.stk {
		ts.DStack[i] = make([]byte, len(dd))
		copy(ts.DStack[i], dd)
	}

	for i, aa := range t.astack.stk {
		ts.AStack[i] = make([]byte, len(aa))
		copy(ts.AStack[i], aa)
	}

	for i, script := range t.scripts {
		ts.Scripts[i] = make(ParsedScript, len(script))
		copy(ts.Scripts[i], script)
	}

	return &ts
}
