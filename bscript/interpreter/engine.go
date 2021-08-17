// Copyright (c) 2013-2018 The btcsuite developers
// Copyright (c) 2015-2018 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package interpreter

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/libsv/go-bk/bec"
	"github.com/libsv/go-bt/v2"
	"github.com/libsv/go-bt/v2/bscript"
	"github.com/libsv/go-bt/v2/sighash"
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
	// This flag should never be used without the ScriptBip16 flag nor the
	// ScriptVerifyWitness flag.
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

	// ScriptVerifyStrictEncoding defines that signature scripts and
	// public keys must follow the strict encoding requirements.
	ScriptVerifyStrictEncoding

	// ScriptVerifyBip143SigHash defines that signature hashes should
	// be calculated using the bip0143 signature hashing algorithm.
	ScriptVerifyBip143SigHash

	// ScriptAfterGenesis defines that the utxo was created after
	// genesis
	ScriptAfterGenesis

	// ScriptVerifyMinimalIf
	ScriptVerifyMinimalIf
)

// HasFlag returns whether the ScriptFlags has the passed flag set.
func (scriptFlags ScriptFlags) HasFlag(flag ScriptFlags) bool {
	return scriptFlags&flag == flag
}

const (
	// MaxStackSize is the maximum combined height of stack and alt stack
	// during execution.
	MaxStackSize = 1000

	// MaxScriptSize is the maximum allowed length of a raw script.
	MaxScriptSize = 110000
)

// halforder is used to tame ECDSA malleability (see BIP0062).
var halfOrder = new(big.Int).Rsh(bec.S256().N, 1)

// Engine is the virtual machine that executes scripts.
type Engine struct {
	dstack stack // data stack
	astack stack // alt stack

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

	numOps    int
	sigCache  SigCache
	hashCache *TxSigHashes

	flags        ScriptFlags
	bip16        bool // treat execution as pay-to-script-hash
	afterGenesis bool
}

// EngineParams are the params required for building an Engine
type EngineParams struct {
	PreviousTxOut *bt.Output
	Tx            *bt.Tx
	InputIdx      int
}

// NewEngine returns a new script engine for the provided public key script,
// transaction, and input index.  The flags modify the behaviour of the script
// engine according to the description provided by each flag.
func NewEngine(opts EngineParams, oo ...EngineOptFunc) (*Engine, error) {
	// The clean stack flag (ScriptVerifyCleanStack) is not allowed without
	// either the pay-to-script-hash (P2SH) evaluation (ScriptBip16)
	// flag or the Segregated Witness (ScriptVerifyWitness) flag.
	//
	// Recall that evaluating a P2SH script without the flag set results in
	// non-P2SH evaluation which leaves the P2SH inputs on the stack.
	// Thus, allowing the clean stack flag without the P2SH flag would make
	// it possible to have a situation where P2SH would not be a soft fork
	// when it should be.
	vm := &Engine{
		prevOutput: opts.PreviousTxOut,
		tx:         opts.Tx,
		inputIdx:   opts.InputIdx,
	}

	for _, o := range oo {
		o(vm)
	}

	if vm.scriptParser == nil {
		WithDefaultParser()(vm)
	}

	if vm.sigCache == nil {
		WithNopSignatureCache()(vm)
	}

	// The provided transaction input index must refer to a valid input.
	if vm.inputIdx < 0 || vm.inputIdx > vm.tx.InputCount()-1 {
		return nil, scriptError(
			ErrInvalidIndex,
			"transaction input index %d is negative or >= %d", opts.InputIdx, len(opts.Tx.Inputs),
		)
	}

	uls := vm.tx.Inputs[opts.InputIdx].UnlockingScript
	ls := vm.prevOutput.LockingScript

	// When both the signature script and public key script are empty the
	// result is necessarily an error since the stack would end up being
	// empty which is equivalent to a false top element.  Thus, just return
	// the relevant error now as an optimization.
	if (uls == nil || len(*uls) == 0) && (ls == nil || len(*ls) == 0) {
		return nil, scriptError(ErrEvalFalse, "false stack entry at end of script execution")
	}

	if vm.hasFlag(ScriptVerifyCleanStack) && (!vm.hasFlag(ScriptBip16)) {
		return nil, scriptError(ErrInvalidFlags, "invalid flags combination")
	}

	// The engine stores the scripts in parsed form using a slice.  This
	// allows multiple scripts to be executed in sequence.  For example,
	// with a pay-to-script-hash transaction, there will be ultimately be
	// a third script to execute.
	scripts := []*bscript.Script{uls, ls}
	vm.scripts = make([]ParsedScript, len(scripts))
	for i, script := range scripts {
		if len(*script) > MaxScriptSize {
			return nil, scriptError(
				ErrScriptTooBig,
				"script size %d is larger than max allowed size %d", len(*script), MaxScriptSize,
			)
		}

		var err error
		if vm.scripts[i], err = vm.scriptParser.Parse(script); err != nil {
			return nil, err
		}
	}

	// The signature script must only contain data pushes when the
	// associated flag is set.
	if vm.hasFlag(ScriptVerifySigPushOnly) && !vm.scripts[0].IsPushOnly() {
		return nil, scriptError(ErrNotPushOnly, "signature script is not push only")
	}

	// Advance the program counter to the public key script if the signature
	// script is empty since there is nothing to execute for it in that
	// case.
	if len(*scripts[0]) == 0 {
		vm.scriptIdx++
	}

	if vm.hasFlag(ScriptBip16) && ls.IsP2SH() {
		// Only accept input scripts that push data for P2SH.
		if !vm.scripts[0].IsPushOnly() {
			return nil, scriptError(ErrNotPushOnly, "pay to script hash is not push only")
		}
		vm.bip16 = true
	}
	if vm.hasFlag(ScriptVerifyMinimalData) {
		vm.dstack.verifyMinimalData = true
		vm.astack.verifyMinimalData = true
	}

	if vm.hasFlag(ScriptAfterGenesis) {
		vm.afterGenesis = true
	}

	vm.tx.InputIdx(vm.inputIdx).PreviousTxScript = vm.prevOutput.LockingScript
	vm.tx.InputIdx(vm.inputIdx).PreviousTxSatoshis = vm.prevOutput.Satoshis

	return vm, nil
}

// hasFlag returns whether the script engine instance has the passed flag set.
func (vm *Engine) hasFlag(flag ScriptFlags) bool {
	return vm.flags.HasFlag(flag)
}

// isBranchExecuting returns whether or not the current conditional branch is
// actively executing. For example, when the data stack has an OP_FALSE on it
// and an OP_IF is encountered, the branch is inactive until an OP_ELSE or
// OP_ENDIF is encountered.  It properly handles nested conditionals.
func (vm *Engine) isBranchExecuting() bool {
	if len(vm.condStack) == 0 {
		return true
	}
	return vm.condStack[len(vm.condStack)-1] == OpCondTrue
}

// executeOpcode performs execution on the passed opcode. It takes into account
// whether or not it is hidden by conditionals, but some rules still must be
// tested in this case.
func (vm *Engine) executeOpcode(pop ParsedOp) error {
	// Disabled opcodes are fail on program counter.
	if pop.IsDisabled() {
		return scriptError(ErrDisabledOpcode, "attempt to execute disabled opcode %s", pop.Name())
	}

	// Always-illegal opcodes are fail on program counter.
	if pop.AlwaysIllegal() {
		return scriptError(ErrReservedOpcode, "attempt to execute reserved opcode %s", pop.Name())
	}

	// Note that this includes OP_RESERVED which counts as a push operation.
	if pop.Op.val > bscript.Op16 {
		vm.numOps++
		if vm.numOps > bscript.MaxOps {
			return scriptError(ErrTooManyOperations, "exceeded max operation limit of %d", bscript.MaxOps)
		}

	} else if len(pop.Data) > bscript.MaxScriptElementSize {
		return scriptError(ErrElementTooBig,
			"element size %d exceeds max allowed size %d", len(pop.Data), bscript.MaxScriptElementSize)
	}

	// Nothing left to do when this is not a conditional opcode and it is
	// not in an executing branch.
	if !vm.isBranchExecuting() && !pop.IsConditional() {
		return nil
	}

	// Ensure all executed data push opcodes use the minimal encoding when
	// the minimal data verification flag is set.
	if vm.dstack.verifyMinimalData && vm.isBranchExecuting() && pop.Op.val <= bscript.OpPUSHDATA4 {
		if err := pop.EnforceMinimumDataPush(); err != nil {
			return err
		}
	}

	return pop.Op.exec(&pop, vm)
}

// disasm is a helper function to produce the output for DisasmPC and
// DisasmScript.  It produces the opcode prefixed by the program counter at the
// provided position in the script.  It does no error checking and leaves that
// to the caller to provide a valid offset.
func (vm *Engine) disasm(scriptIdx int, scriptOff int) string {
	return fmt.Sprintf("%02x:%04x: %s", scriptIdx, scriptOff, vm.scripts[scriptIdx][scriptOff].print(false))
}

// validPC returns an error if the current script position is valid for
// execution, nil otherwise.
func (vm *Engine) validPC() error {
	if vm.scriptIdx >= len(vm.scripts) {
		return scriptError(ErrInvalidProgramCounter,
			"past input scripts %v:%v %v:xxxx", vm.scriptIdx, vm.scriptOff, len(vm.scripts))
	}
	if vm.scriptOff >= len(vm.scripts[vm.scriptIdx]) {
		return scriptError(ErrInvalidProgramCounter, "past input scripts %v:%v %v:%04d", vm.scriptIdx, vm.scriptOff,
			vm.scriptIdx, len(vm.scripts[vm.scriptIdx]))
	}
	return nil
}

// curPC returns either the current script and offset, or an error if the
// position isn't valid.
func (vm *Engine) curPC() (script int, off int, err error) {
	err = vm.validPC()
	if err != nil {
		return 0, 0, err
	}
	return vm.scriptIdx, vm.scriptOff, nil
}

// DisasmPC returns the string for the disassembly of the opcode that will be
// next to execute when Step() is called.
func (vm *Engine) DisasmPC() (string, error) {
	scriptIdx, scriptOff, err := vm.curPC()
	if err != nil {
		return "", err
	}
	return vm.disasm(scriptIdx, scriptOff), nil
}

// DisasmScript returns the disassembly string for the script at the requested
// offset index.  Index 0 is the signature script and 1 is the public key
// script.
func (vm *Engine) DisasmScript(idx int) (string, error) {
	if idx >= len(vm.scripts) {
		return "", scriptError(ErrInvalidIndex, "script index %d >= total scripts %d", idx, len(vm.scripts))
	}

	//var disstr string
	var b strings.Builder
	for i := range vm.scripts[idx] {
		b.WriteString(vm.disasm(idx, i))
		b.WriteRune('\n')
	}
	return b.String(), nil
}

// CheckErrorCondition returns nil if the running script has ended and was
// successful, leaving a a true boolean on the stack.  An error otherwise,
// including if the script has not finished.
func (vm *Engine) CheckErrorCondition(finalScript bool) error {
	// Check execution is actually done.  When pc is past the end of script
	// array there are no more scripts to run.
	if vm.scriptIdx < len(vm.scripts) {
		return scriptError(ErrScriptUnfinished, "error check when script unfinished")
	}
	if finalScript && vm.hasFlag(ScriptVerifyCleanStack) && vm.dstack.Depth() != 1 {
		return scriptError(ErrCleanStack, "stack contains %d unexpected items", vm.dstack.Depth()-1)
	}
	if vm.dstack.Depth() < 1 {
		return scriptError(ErrEmptyStack, "stack empty at end of script execution")
	}

	v, err := vm.dstack.PopBool()
	if err != nil {
		return err
	}
	if !v {
		log.Tracef("%v", newLogClosure(func() string {
			dis0, _ := vm.DisasmScript(0)
			dis1, _ := vm.DisasmScript(1)
			return fmt.Sprintf("scripts failed: script0: %s\nscript1: %s\n", dis0, dis1)
		}))

		return scriptError(ErrEvalFalse, "false stack entry at end of script execution")
	}

	return nil
}

// Step will execute the next instruction and move the program counter to the
// next opcode in the script, or the next script if the current has ended.  Step
// will return true in the case that the last opcode was successfully executed.
//
// The result of calling Step or any other method is undefined if an error is
// returned.
func (vm *Engine) Step() (done bool, err error) {
	// Verify that it is pointing to a valid script address.
	if err = vm.validPC(); err != nil {
		return true, err
	}

	opcode := vm.scripts[vm.scriptIdx][vm.scriptOff]
	vm.scriptOff++

	// Execute the opcode while taking into account several things such as
	// disabled opcodes, illegal opcodes, maximum allowed operations per
	// script, maximum script element sizes, and conditionals.
	if err = vm.executeOpcode(opcode); err != nil {
		return true, err
	}

	// The number of elements in the combination of the data and alt stacks
	// must not exceed the maximum number of stack elements allowed.
	combinedStackSize := vm.dstack.Depth() + vm.astack.Depth()
	if combinedStackSize > MaxStackSize {
		return false, scriptError(ErrStackOverflow,
			"combined stack size %d > max allowed %d", combinedStackSize, MaxStackSize)
	}

	if vm.scriptOff < len(vm.scripts[vm.scriptIdx]) {
		return false, nil
	}

	// Prepare for next instruction.
	// Illegal to have an `if' that straddles two scripts.
	if len(vm.condStack) != 0 {
		return false, scriptError(ErrUnbalancedConditional, "end of script reached in conditional execution")
	}

	// Alt stack doesn't persist.
	_ = vm.astack.DropN(vm.astack.Depth())

	vm.numOps = 0 // number of ops is per script.
	vm.scriptOff = 0
	vm.scriptIdx++
	if vm.scriptIdx == 1 && vm.bip16 {
		vm.savedFirstStack = vm.GetStack()
	}

	if vm.scriptIdx == 2 && vm.bip16 {
		// Put us past the end for CheckErrorCondition()
		// Check script ran successfully and pull the script
		// out of the first stack and execute that.
		if err := vm.CheckErrorCondition(false); err != nil {
			return false, err
		}

		script := vm.savedFirstStack[len(vm.savedFirstStack)-1]
		pops, err := vm.scriptParser.Parse(bscript.NewFromBytes(script))
		if err != nil {
			return false, err
		}
		vm.scripts = append(vm.scripts, pops)

		// Set stack to be the stack from first script minus the
		// script itself
		vm.SetStack(vm.savedFirstStack[:len(vm.savedFirstStack)-1])
	}

	// there are zero length scripts in the wild
	if vm.scriptIdx < len(vm.scripts) && vm.scriptOff >= len(vm.scripts[vm.scriptIdx]) {
		vm.scriptIdx++
	}

	vm.lastCodeSep = 0
	if vm.scriptIdx >= len(vm.scripts) {
		return true, nil
	}

	return false, nil
}

// Execute will execute all scripts in the script engine and return either nil
// for successful validation or an error if one occurred.
func (vm *Engine) Execute() (err error) {
	var done bool
	for !done {
		log.Tracef("%v", newLogClosure(func() string {
			var dis string
			if dis, err = vm.DisasmPC(); err != nil {
				return fmt.Sprintf("stepping (%v)", err)
			}
			return fmt.Sprintf("stepping %v", dis)
		}))

		if done, err = vm.Step(); err != nil {
			return err
		}

		log.Tracef("%v", newLogClosure(func() string {
			var dstr, astr string

			// if we're tracing, dump the stacks.
			if vm.dstack.Depth() != 0 {
				dstr = "Stack:\n" + vm.dstack.String()
			}
			if vm.astack.Depth() != 0 {
				astr = "AltStack:\n" + vm.astack.String()
			}

			return dstr + astr
		}))
	}

	return vm.CheckErrorCondition(true)
}

// subScript returns the script since the last OP_CODESEPARATOR.
func (vm *Engine) subScript() ParsedScript {
	return vm.scripts[vm.scriptIdx][vm.lastCodeSep:]
}

// checkHashTypeEncoding returns whether or not the passed hashtype adheres to
// the strict encoding requirements if enabled.
func (vm *Engine) checkHashTypeEncoding(shf sighash.Flag) error {
	if !vm.hasFlag(ScriptVerifyStrictEncoding) {
		return nil
	}

	sigHashType := shf & ^sighash.AnyOneCanPay
	if vm.hasFlag(ScriptVerifyBip143SigHash) {
		sigHashType ^= sighash.ForkID
		if shf&sighash.ForkID == 0 {
			return scriptError(ErrInvalidSigHashType, "hash type does not contain uahf forkID 0x%x", shf)
		}
	}

	if sigHashType < sighash.All || sigHashType > sighash.Single {
		return scriptError(ErrInvalidSigHashType, "invalid hash type 0x%x", shf)
	}
	return nil
}

// checkPubKeyEncoding returns whether or not the passed public key adheres to
// the strict encoding requirements if enabled.
func (vm *Engine) checkPubKeyEncoding(pubKey []byte) error {
	if !vm.hasFlag(ScriptVerifyStrictEncoding) {
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

	return scriptError(ErrPubKeyType, "unsupported public key type")
}

// checkSignatureEncoding returns whether or not the passed signature adheres to
// the strict encoding requirements if enabled.
func (vm *Engine) checkSignatureEncoding(sig []byte) error {
	if !vm.hasFlag(ScriptVerifyDERSignatures) && !vm.hasFlag(ScriptVerifyLowS) && !vm.hasFlag(ScriptVerifyStrictEncoding) {
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
		// might required if the high bit is set in the value.
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
		return scriptError(ErrSigTooShort, "malformed signature: too short: %d < %d", sigLen, minSigLen)
	}
	if sigLen > maxSigLen {
		return scriptError(ErrSigTooLong, "malformed signature: too long: %d > %d", sigLen, maxSigLen)
	}

	// The signature must start with the ASN.1 sequence identifier.
	if sig[sequenceOffset] != asn1SequenceID {
		return scriptError(ErrSigInvalidSeqID, "malformed signature: format has wrong type: %#x", sig[sequenceOffset])
	}

	// The signature must indicate the correct amount of data for all elements
	// related to R and S.
	if int(sig[dataLenOffset]) != sigLen-2 {
		return scriptError(ErrSigInvalidDataLen, "malformed signature: bad length: %d != %d", sig[dataLenOffset], sigLen-2)
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
		return scriptError(ErrSigMissingSTypeID, "malformed signature: S type indicator missing")
	}
	if sLenOffset >= sigLen {
		return scriptError(ErrSigMissingSLen, "malformed signature: S length missing")
	}

	// The lengths of R and S must match the overall length of the signature.
	//
	// sLen specifies the length of the big-endian encoded number which
	// represents the S value of the signature.
	sOffset := sLenOffset + 1
	sLen := int(sig[sLenOffset])
	if sOffset+sLen != sigLen {
		return scriptError(ErrSigInvalidSLen, "malformed signature: invalid S length")
	}

	// R elements must be ASN.1 integers.
	if sig[rTypeOffset] != asn1IntegerID {
		return scriptError(ErrSigInvalidRIntID,
			"malformed signature: R integer marker: %#x != %#x", sig[rTypeOffset], asn1IntegerID)
	}

	// Zero-length integers are not allowed for R.
	if rLen == 0 {
		return scriptError(ErrSigZeroRLen, "malformed signature: R length is zero")
	}

	// R must not be negative.
	if sig[rOffset]&0x80 != 0 {
		return scriptError(ErrSigNegativeR, "malformed signature: R is negative")
	}

	// Null bytes at the start of R are not allowed, unless R would otherwise be
	// interpreted as a negative number.
	if rLen > 1 && sig[rOffset] == 0x00 && sig[rOffset+1]&0x80 == 0 {
		return scriptError(ErrSigTooMuchRPadding, "malformed signature: R value has too much padding")
	}

	// S elements must be ASN.1 integers.
	if sig[sTypeOffset] != asn1IntegerID {
		return scriptError(ErrSigInvalidSIntID,
			"malformed signature: S integer marker: %#x != %#x", sig[sTypeOffset], asn1IntegerID)
	}

	// Zero-length integers are not allowed for S.
	if sLen == 0 {
		return scriptError(ErrSigZeroSLen, "malformed signature: S length is zero")
	}

	// S must not be negative.
	if sig[sOffset]&0x80 != 0 {
		return scriptError(ErrSigNegativeS, "malformed signature: S is negative")
	}

	// Null bytes at the start of S are not allowed, unless S would otherwise be
	// interpreted as a negative number.
	if sLen > 1 && sig[sOffset] == 0x00 && sig[sOffset+1]&0x80 == 0 {
		return scriptError(ErrSigTooMuchSPadding, "malformed signature: S value has too much padding")
	}

	// Verify the S value is <= half the order of the curve.  This check is done
	// because when it is higher, the complement modulo the order can be used
	// instead which is a shorter encoding by 1 byte.  Further, without
	// enforcing this, it is possible to replace a signature in a valid
	// transaction with the complement while still being a valid signature that
	// verifies.  This would result in changing the transaction hash and thus is
	// a source of malleability.
	if vm.hasFlag(ScriptVerifyLowS) {
		sValue := new(big.Int).SetBytes(sig[sOffset : sOffset+sLen])
		if sValue.Cmp(halfOrder) > 0 {
			return scriptError(ErrSigHighS, "signature is not canonical due to unnecessarily high S value")
		}
	}

	return nil
}

// getStack returns the contents of stack as a byte array bottom up
func getStack(stack *stack) [][]byte {
	array := make([][]byte, stack.Depth())
	for i := range array {
		// PeekByteArry can't fail due to overflow, already checked
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

// GetStack returns the contents of the primary stack as an array. where the
// last item in the array is the top of the stack.
func (vm *Engine) GetStack() [][]byte {
	return getStack(&vm.dstack)
}

// SetStack sets the contents of the primary stack to the contents of the
// provided array where the last item in the array will be the top of the stack.
func (vm *Engine) SetStack(data [][]byte) {
	setStack(&vm.dstack, data)
}

// GetAltStack returns the contents of the alternate stack as an array where the
// last item in the array is the top of the stack.
func (vm *Engine) GetAltStack() [][]byte {
	return getStack(&vm.astack)
}

// SetAltStack sets the contents of the alternate stack to the contents of the
// provided array where the last item in the array will be the top of the stack.
func (vm *Engine) SetAltStack(data [][]byte) {
	setStack(&vm.astack, data)
}
