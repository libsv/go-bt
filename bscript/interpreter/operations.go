package interpreter

import (
	"bytes"
	"crypto/sha1" //nolint:gosec // OP_SHA1 support requires this
	"crypto/sha256"
	"fmt"
	"hash"

	"github.com/libsv/go-bk/bec"
	"github.com/libsv/go-bk/crypto"
	"github.com/libsv/go-bt/v2"
	"github.com/libsv/go-bt/v2/bscript"
	"github.com/libsv/go-bt/v2/sighash"
	"golang.org/x/crypto/ripemd160"
)

// *******************************************
// Opcode implementation functions start here.
// *******************************************

// opcodeDisabled is a common handler for disabled opcodes.  It returns an
// appropriate error indicating the opcode is disabled.  While it would
// ordinarily make more sense to detect if the script contains any disabled
// opcodes before executing in an initial parse step, the consensus rules
// dictate the script doesn't fail until the program counter passes over a
// disabled opcode (even when they appear in a branch that is not executed).
func opcodeDisabled(op *ParsedOp, vm *Engine) error {
	return scriptError(ErrDisabledOpcode, "attempt to execute disabled opcode %s", op.Name())
}

// opcodeReserved is a common handler for all reserved opcodes.  It returns an
// appropriate error indicating the opcode is reserved.
func opcodeReserved(op *ParsedOp, vm *Engine) error {
	return scriptError(ErrReservedOpcode, "attempt to execute reserved opcode %s", op.Name())
}

// opcodeInvalid is a common handler for all invalid opcodes.  It returns an
// appropriate error indicating the opcode is invalid.
func opcodeInvalid(op *ParsedOp, vm *Engine) error {
	return scriptError(ErrReservedOpcode, "attempt to execute invalid opcode %s", op.Name())
}

// opcodeFalse pushes an empty array to the data stack to represent false.  Note
// that 0, when encoded as a number according to the numeric encoding consensus
// rules, is an empty array.
func opcodeFalse(op *ParsedOp, vm *Engine) error {
	vm.dstack.PushByteArray(nil)
	return nil
}

// opcodePushData is a common handler for the vast majority of opcodes that push
// raw data (bytes) to the data stack.
func opcodePushData(op *ParsedOp, vm *Engine) error {
	vm.dstack.PushByteArray(op.Data)
	return nil
}

// opcode1Negate pushes -1, encoded as a number, to the data stack.
func opcode1Negate(op *ParsedOp, vm *Engine) error {
	vm.dstack.PushInt(-1)
	return nil
}

// opcodeN is a common handler for the small integer data push opcodes.  It
// pushes the numeric value the opcode represents (which will be from 1 to 16)
// onto the data stack.
func opcodeN(op *ParsedOp, vm *Engine) error {
	// The opcodes are all defined consecutively, so the numeric value is
	// the difference.
	vm.dstack.PushInt(scriptNum((op.Op.val - (bscript.Op1 - 1))))
	return nil
}

// opcodeNop is a common handler for the NOP family of opcodes.  As the name
// implies it generally does nothing, however, it will return an error when
// the flag to discourage use of NOPs is set for select opcodes.
func opcodeNop(op *ParsedOp, vm *Engine) error {
	switch op.Op.val {
	case bscript.OpNOP1, bscript.OpNOP4, bscript.OpNOP5,
		bscript.OpNOP6, bscript.OpNOP7, bscript.OpNOP8, bscript.OpNOP9, bscript.OpNOP10:
		if vm.hasFlag(ScriptDiscourageUpgradableNops) {
			return scriptError(
				ErrDiscourageUpgradableNOPs,
				"bscript.OpNOP%d reserved for soft-fork upgrades",
				op.Op.val-(bscript.OpNOP1-1),
			)
		}
	}

	return nil
}

// popIfBool pops the top item off the stack and returns a bool
func popIfBool(vm *Engine) (bool, error) {
	return vm.dstack.PopBool()
}

// opcodeIf treats the top item on the data stack as a boolean and removes it.
//
// An appropriate entry is added to the conditional stack depending on whether
// the boolean is true and whether this if is on an executing branch in order
// to allow proper execution of further opcodes depending on the conditional
// logic.  When the boolean is true, the first branch will be executed (unless
// this opcode is nested in a non-executed branch).
//
// <expression> if [statements] [else [statements]] endif
//
// Note that, unlike for all non-conditional opcodes, this is executed even when
// it is on a non-executing branch so proper nesting is maintained.
//
// Data stack transformation: [... bool] -> [...]
// Conditional stack transformation: [...] -> [... OpCondValue]
func opcodeIf(op *ParsedOp, vm *Engine) error {
	condVal := OpCondFalse
	if vm.isBranchExecuting() {
		ok, err := popIfBool(vm)
		if err != nil {
			return err
		}

		if ok {
			condVal = OpCondTrue
		}
	} else {
		condVal = OpCondSkip
	}

	vm.condStack = append(vm.condStack, condVal)
	return nil
}

// opcodeNotIf treats the top item on the data stack as a boolean and removes
// it.
//
// An appropriate entry is added to the conditional stack depending on whether
// the boolean is true and whether this if is on an executing branch in order
// to allow proper execution of further opcodes depending on the conditional
// logic.  When the boolean is false, the first branch will be executed (unless
// this opcode is nested in a non-executed branch).
//
// <expression> notif [statements] [else [statements]] endif
//
// Note that, unlike for all non-conditional opcodes, this is executed even when
// it is on a non-executing branch so proper nesting is maintained.
//
// Data stack transformation: [... bool] -> [...]
// Conditional stack transformation: [...] -> [... OpCondValue]
func opcodeNotIf(op *ParsedOp, vm *Engine) error {
	condVal := OpCondFalse
	if vm.isBranchExecuting() {
		ok, err := popIfBool(vm)
		if err != nil {
			return err
		}

		if !ok {
			condVal = OpCondTrue
		}
	} else {
		condVal = OpCondSkip
	}

	vm.condStack = append(vm.condStack, condVal)
	return nil
}

// opcodeElse inverts conditional execution for other half of if/else/endif.
//
// An error is returned if there has not already been a matching bscript.OpIF.
//
// Conditional stack transformation: [... OpCondValue] -> [... !OpCondValue]
func opcodeElse(op *ParsedOp, vm *Engine) error {
	if len(vm.condStack) == 0 {
		return scriptError(ErrUnbalancedConditional,
			"encountered opcode %s with no matching opcode to begin conditional execution", op.Name())
	}

	conditionalIdx := len(vm.condStack) - 1
	switch vm.condStack[conditionalIdx] {
	case OpCondTrue:
		vm.condStack[conditionalIdx] = OpCondFalse
	case OpCondFalse:
		vm.condStack[conditionalIdx] = OpCondTrue
	case OpCondSkip:
		// Value doesn't change in skip since it indicates this opcode
		// is nested in a non-executed branch.
	}
	return nil
}

// opcodeEndif terminates a conditional block, removing the value from the
// conditional execution stack.
//
// An error is returned if there has not already been a matching bscript.OpIF.
//
// Conditional stack transformation: [... OpCondValue] -> [...]
func opcodeEndif(op *ParsedOp, vm *Engine) error {
	if len(vm.condStack) == 0 {
		return scriptError(ErrUnbalancedConditional,
			"encountered opcode %s with no matching opcode to begin conditional execution", op.Name())
	}

	vm.condStack = vm.condStack[:len(vm.condStack)-1]
	return nil
}

// abstractVerify examines the top item on the data stack as a boolean value and
// verifies it evaluates to true.  An error is returned either when there is no
// item on the stack or when that item evaluates to false.  In the latter case
// where the verification fails specifically due to the top item evaluating
// to false, the returned error will use the passed error code.
func abstractVerify(op *ParsedOp, vm *Engine, c ErrorCode) error {
	verified, err := vm.dstack.PopBool()
	if err != nil {
		return err
	}
	if !verified {
		return scriptError(c, "%s failed", op.Name())
	}

	return nil
}

// opcodeVerify examines the top item on the data stack as a boolean value and
// verifies it evaluates to true.  An error is returned if it does not.
func opcodeVerify(op *ParsedOp, vm *Engine) error {
	return abstractVerify(op, vm, ErrVerify)
}

// opcodeReturn returns an appropriate error since it is always an error to
// return early from a script.
func opcodeReturn(op *ParsedOp, vm *Engine) error {
	return scriptError(ErrEarlyReturn, "script returned early")
}

// verifyLockTime is a helper function used to validate locktimes.
func verifyLockTime(txLockTime, threshold, lockTime int64) error {
	// The lockTimes in both the script and transaction must be of the same
	// type.
	if !((txLockTime < threshold && lockTime < threshold) ||
		(txLockTime >= threshold && lockTime >= threshold)) {
		return scriptError(ErrUnsatisfiedLockTime,
			"mismatched locktime types -- tx locktime %d, stack locktime %d", txLockTime, lockTime)
	}

	if lockTime > txLockTime {
		return scriptError(ErrUnsatisfiedLockTime,
			"locktime requirement not satisfied -- locktime is greater than the transaction locktime: %d > %d",
			lockTime, txLockTime)
	}

	return nil
}

// opcodeCheckLockTimeVerify compares the top item on the data stack to the
// LockTime field of the transaction containing the script signature
// validating if the transaction outputs are spendable yet.  If flag
// ScriptVerifyCheckLockTimeVerify is not set, the code continues as if bscript.OpNOP2
// were executed.
func opcodeCheckLockTimeVerify(op *ParsedOp, vm *Engine) error {
	// If the ScriptVerifyCheckLockTimeVerify script flag is not set, treat
	// opcode as bscript.OpNOP2 instead.
	if !vm.hasFlag(ScriptVerifyCheckLockTimeVerify) {
		if vm.hasFlag(ScriptDiscourageUpgradableNops) {
			return scriptError(ErrDiscourageUpgradableNOPs, "bscript.OpNOP2 reserved for soft-fork upgrades")
		}

		return nil
	}

	// The current transaction locktime is a uint32 resulting in a maximum
	// locktime of 2^32-1 (the year 2106).  However, scriptNums are signed
	// and therefore a standard 4-byte scriptNum would only support up to a
	// maximum of 2^31-1 (the year 2038).  Thus, a 5-byte scriptNum is used
	// here since it will support up to 2^39-1 which allows dates beyond the
	// current locktime limit.
	//
	// PeekByteArray is used here instead of PeekInt because we do not want
	// to be limited to a 4-byte integer for reasons specified above.
	so, err := vm.dstack.PeekByteArray(0)
	if err != nil {
		return err
	}
	lockTime, err := makeScriptNum(so, vm.dstack.verifyMinimalData, 5)
	if err != nil {
		return err
	}

	// In the rare event that the argument needs to be < 0 due to some
	// arithmetic being done first, you can always use
	// 0 bscript.OpMAX bscript.OpCHECKLOCKTIMEVERIFY.
	if lockTime < 0 {
		return scriptError(ErrNegativeLockTime, "negative lock time: %d", lockTime)
	}

	// The lock time field of a transaction is either a block height at
	// which the transaction is finalised or a timestamp depending on if the
	// value is before the txscript.LockTimeThreshold.  When it is under the
	// threshold it is a block height.
	if err = verifyLockTime(int64(vm.tx.LockTime), LockTimeThreshold, int64(lockTime)); err != nil {
		return err
	}

	// The lock time feature can also be disabled, thereby bypassing
	// bscript.OpCHECKLOCKTIMEVERIFY, if every transaction input has been finalised by
	// setting its sequence to the maximum value (bt.MaxTxInSequenceNum).  This
	// condition would result in the transaction being allowed into the blockchain
	// making the opcode ineffective.
	//
	// This condition is prevented by enforcing that the input being used by
	// the opcode is unlocked (its sequence number is less than the max
	// value).  This is sufficient to prove correctness without having to
	// check every input.
	//
	// NOTE: This implies that even if the transaction is not finalised due to
	// another input being unlocked, the opcode execution will still fail when the
	// input being used by the opcode is locked.
	if vm.tx.Inputs[vm.inputIdx].SequenceNumber == bt.MaxTxInSequenceNum {
		return scriptError(ErrUnsatisfiedLockTime, "transaction input is finalised")
	}

	return nil
}

// opcodeCheckSequenceVerify compares the top item on the data stack to the
// LockTime field of the transaction containing the script signature
// validating if the transaction outputs are spendable yet.  If flag
// ScriptVerifyCheckSequenceVerify is not set, the code continues as if bscript.OpNOP3
// were executed.
func opcodeCheckSequenceVerify(op *ParsedOp, vm *Engine) error {
	// If the ScriptVerifyCheckSequenceVerify script flag is not set, treat
	// opcode as bscript.OpNOP3 instead.
	if !vm.hasFlag(ScriptVerifyCheckSequenceVerify) {
		if vm.hasFlag(ScriptDiscourageUpgradableNops) {
			return scriptError(ErrDiscourageUpgradableNOPs, "bscript.OpNOP3 reserved for soft-fork upgrades")
		}

		return nil
	}

	// The current transaction sequence is a uint32 resulting in a maximum
	// sequence of 2^32-1.  However, scriptNums are signed and therefore a
	// standard 4-byte scriptNum would only support up to a maximum of
	// 2^31-1.  Thus, a 5-byte scriptNum is used here since it will support
	// up to 2^39-1 which allows sequences beyond the current sequence
	// limit.
	//
	// PeekByteArray is used here instead of PeekInt because we do not want
	// to be limited to a 4-byte integer for reasons specified above.
	so, err := vm.dstack.PeekByteArray(0)
	if err != nil {
		return err
	}
	stackSequence, err := makeScriptNum(so, vm.dstack.verifyMinimalData, 5)
	if err != nil {
		return err
	}

	// In the rare event that the argument needs to be < 0 due to some
	// arithmetic being done first, you can always use
	// 0 bscript.OpMAX bscript.OpCHECKSEQUENCEVERIFY.
	if stackSequence < 0 {
		return scriptError(ErrNegativeLockTime, "negative sequence: %d", stackSequence)
	}

	sequence := int64(stackSequence)

	// To provide for future soft-fork extensibility, if the
	// operand has the disabled lock-time flag set,
	// CHECKSEQUENCEVERIFY behaves as a NOP.
	if sequence&int64(bt.SequenceLockTimeDisabled) != 0 {
		return nil
	}

	// Transaction version numbers not high enough to trigger CSV rules must
	// fail.
	if vm.tx.Version < 2 {
		return scriptError(ErrUnsatisfiedLockTime, "invalid transaction version: %d", vm.tx.Version)
	}

	// Sequence numbers with their most significant bit set are not
	// consensus constrained. Testing that the transaction's sequence
	// number does not have this bit set prevents using this property
	// to get around a CHECKSEQUENCEVERIFY check.
	txSequence := int64(vm.tx.Inputs[vm.inputIdx].SequenceNumber)
	if txSequence&int64(bt.SequenceLockTimeDisabled) != 0 {
		return scriptError(ErrUnsatisfiedLockTime,
			"transaction sequence has sequence locktime disabled bit set: 0x%x", txSequence)
	}

	// Mask off non-consensus bits before doing comparisons.
	lockTimeMask := int64(bt.SequenceLockTimeIsSeconds | bt.SequenceLockTimeMask)

	return verifyLockTime(txSequence&lockTimeMask, bt.SequenceLockTimeIsSeconds, sequence&lockTimeMask)
}

// opcodeToAltStack removes the top item from the main data stack and pushes it
// onto the alternate data stack.
//
// Main data stack transformation: [... x1 x2 x3] -> [... x1 x2]
// Alt data stack transformation:  [... y1 y2 y3] -> [... y1 y2 y3 x3]
func opcodeToAltStack(op *ParsedOp, vm *Engine) error {
	so, err := vm.dstack.PopByteArray()
	if err != nil {
		return err
	}

	vm.astack.PushByteArray(so)

	return nil
}

// opcodeFromAltStack removes the top item from the alternate data stack and
// pushes it onto the main data stack.
//
// Main data stack transformation: [... x1 x2 x3] -> [... x1 x2 x3 y3]
// Alt data stack transformation:  [... y1 y2 y3] -> [... y1 y2]
func opcodeFromAltStack(op *ParsedOp, vm *Engine) error {
	so, err := vm.astack.PopByteArray()
	if err != nil {
		return err
	}

	vm.dstack.PushByteArray(so)

	return nil
}

// opcode2Drop removes the top 2 items from the data stack.
//
// Stack transformation: [... x1 x2 x3] -> [... x1]
func opcode2Drop(op *ParsedOp, vm *Engine) error {
	return vm.dstack.DropN(2)
}

// opcode2Dup duplicates the top 2 items on the data stack.
//
// Stack transformation: [... x1 x2 x3] -> [... x1 x2 x3 x2 x3]
func opcode2Dup(op *ParsedOp, vm *Engine) error {
	return vm.dstack.DupN(2)
}

// opcode3Dup duplicates the top 3 items on the data stack.
//
// Stack transformation: [... x1 x2 x3] -> [... x1 x2 x3 x1 x2 x3]
func opcode3Dup(op *ParsedOp, vm *Engine) error {
	return vm.dstack.DupN(3)
}

// opcode2Over duplicates the 2 items before the top 2 items on the data stack.
//
// Stack transformation: [... x1 x2 x3 x4] -> [... x1 x2 x3 x4 x1 x2]
func opcode2Over(op *ParsedOp, vm *Engine) error {
	return vm.dstack.OverN(2)
}

// opcode2Rot rotates the top 6 items on the data stack to the left twice.
//
// Stack transformation: [... x1 x2 x3 x4 x5 x6] -> [... x3 x4 x5 x6 x1 x2]
func opcode2Rot(op *ParsedOp, vm *Engine) error {
	return vm.dstack.RotN(2)
}

// opcode2Swap swaps the top 2 items on the data stack with the 2 that come
// before them.
//
// Stack transformation: [... x1 x2 x3 x4] -> [... x3 x4 x1 x2]
func opcode2Swap(op *ParsedOp, vm *Engine) error {
	return vm.dstack.SwapN(2)
}

// opcodeIfDup duplicates the top item of the stack if it is not zero.
//
// Stack transformation (x1==0): [... x1] -> [... x1]
// Stack transformation (x1!=0): [... x1] -> [... x1 x1]
func opcodeIfDup(op *ParsedOp, vm *Engine) error {
	so, err := vm.dstack.PeekByteArray(0)
	if err != nil {
		return err
	}

	// Push copy of data iff it isn't zero
	if asBool(so) {
		vm.dstack.PushByteArray(so)
	}

	return nil
}

// opcodeDepth pushes the depth of the data stack prior to executing this
// opcode, encoded as a number, onto the data stack.
//
// Stack transformation: [...] -> [... <num of items on the stack>]
// Example with 2 items: [x1 x2] -> [x1 x2 2]
// Example with 3 items: [x1 x2 x3] -> [x1 x2 x3 3]
func opcodeDepth(op *ParsedOp, vm *Engine) error {
	vm.dstack.PushInt(scriptNum(vm.dstack.Depth()))
	return nil
}

// opcodeDrop removes the top item from the data stack.
//
// Stack transformation: [... x1 x2 x3] -> [... x1 x2]
func opcodeDrop(op *ParsedOp, vm *Engine) error {
	return vm.dstack.DropN(1)
}

// opcodeDup duplicates the top item on the data stack.
//
// Stack transformation: [... x1 x2 x3] -> [... x1 x2 x3 x3]
func opcodeDup(op *ParsedOp, vm *Engine) error {
	return vm.dstack.DupN(1)
}

// opcodeNip removes the item before the top item on the data stack.
//
// Stack transformation: [... x1 x2 x3] -> [... x1 x3]
func opcodeNip(op *ParsedOp, vm *Engine) error {
	return vm.dstack.NipN(1)
}

// opcodeOver duplicates the item before the top item on the data stack.
//
// Stack transformation: [... x1 x2 x3] -> [... x1 x2 x3 x2]
func opcodeOver(op *ParsedOp, vm *Engine) error {
	return vm.dstack.OverN(1)
}

// opcodePick treats the top item on the data stack as an integer and duplicates
// the item on the stack that number of items back to the top.
//
// Stack transformation: [xn ... x2 x1 x0 n] -> [xn ... x2 x1 x0 xn]
// Example with n=1: [x2 x1 x0 1] -> [x2 x1 x0 x1]
// Example with n=2: [x2 x1 x0 2] -> [x2 x1 x0 x2]
func opcodePick(op *ParsedOp, vm *Engine) error {
	val, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	return vm.dstack.PickN(val.Int32())
}

// opcodeRoll treats the top item on the data stack as an integer and moves
// the item on the stack that number of items back to the top.
//
// Stack transformation: [xn ... x2 x1 x0 n] -> [... x2 x1 x0 xn]
// Example with n=1: [x2 x1 x0 1] -> [x2 x0 x1]
// Example with n=2: [x2 x1 x0 2] -> [x1 x0 x2]
func opcodeRoll(op *ParsedOp, vm *Engine) error {
	val, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	return vm.dstack.RollN(val.Int32())
}

// opcodeRot rotates the top 3 items on the data stack to the left.
//
// Stack transformation: [... x1 x2 x3] -> [... x2 x3 x1]
func opcodeRot(op *ParsedOp, vm *Engine) error {
	return vm.dstack.RotN(1)
}

// opcodeSwap swaps the top two items on the stack.
//
// Stack transformation: [... x1 x2] -> [... x2 x1]
func opcodeSwap(op *ParsedOp, vm *Engine) error {
	return vm.dstack.SwapN(1)
}

// opcodeTuck inserts a duplicate of the top item of the data stack before the
// second-to-top item.
//
// Stack transformation: [... x1 x2] -> [... x2 x1 x2]
func opcodeTuck(op *ParsedOp, vm *Engine) error {
	return vm.dstack.Tuck()
}

// opcodeCat concatenates two byte sequences. The result must
// not be larger than MaxScriptElementSize.
//
// Stack transformation: {Ox11} {0x22, 0x33} bscript.OpCAT -> 0x112233
func opcodeCat(op *ParsedOp, vm *Engine) error {
	b, err := vm.dstack.PopByteArray()
	if err != nil {
		return err
	}

	a, err := vm.dstack.PopByteArray()
	if err != nil {
		return err
	}

	c := append(a, b...)
	if len(c) > bscript.MaxScriptElementSize {
		return scriptError(ErrElementTooBig,
			"concatenated size %d exceeds max allowed size %d", len(c), bscript.MaxScriptElementSize)
	}

	vm.dstack.PushByteArray(c)
	return nil
}

// opcodeSplit splits the operand at the given position.
// This operation is the exact inverse of bscript.OpCAT
//
// Stack transformation: x n bscript.OpSPLIT -> x1 x2
func opcodeSplit(op *ParsedOp, vm *Engine) error {
	n, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	c, err := vm.dstack.PopByteArray()
	if err != nil {
		return err
	}

	if n.Int32() > int32(len(c)) {
		return scriptError(ErrNumberTooBig, "n is larger than length of array")
	}
	if n < 0 {
		return scriptError(ErrNumberTooSmall, "n is negative")
	}

	a := c[:n]
	b := c[n:]
	vm.dstack.PushByteArray(a)
	vm.dstack.PushByteArray(b)

	return nil
}

// opcodeNum2Bin converts the numeric value into a byte sequence of a
// certain size, taking account of the sign bit. The byte sequence
// produced uses the little-endian encoding.
//
// Stack transformation: a b bscript.OpNUM2BIN -> x
func opcodeNum2bin(op *ParsedOp, vm *Engine) error {
	n, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	a, err := vm.dstack.PopByteArray()
	if err != nil {
		return err
	}

	size := int(n.Int32())
	if size > bscript.MaxScriptElementSize {
		return scriptError(ErrNumberTooBig, "n is larger than the max of %d", defaultScriptNumLen)
	}

	// encode a as a script num so that we we take the bytes it
	// will be minimally encoded.
	sn, err := makeScriptNum(a, false, len(a))
	if err != nil {
		return err
	}

	b := sn.Bytes()
	if len(b) > size {
		return scriptError(ErrNumberTooSmall, "cannot fit it into n sized array")
	}
	if len(b) == size {
		vm.dstack.PushByteArray(b)
		return nil
	}

	signbit := byte(0x00)
	if len(b) > 0 {
		signbit = b[0] & 0x80
		b[len(b)-1] &= 0x7f
	}

	for len(b) < size-1 {
		b = append(b, 0x00)
	}

	b = append(b, signbit)

	vm.dstack.PushByteArray(b)
	return nil
}

// opcodeBin2num converts the byte sequence into a numeric value,
// including minimal encoding. The byte sequence must encode the
// value in little-endian encoding.
//
// Stack transformation: a bscript.OpBIN2NUM -> x
func opcodeBin2num(op *ParsedOp, vm *Engine) error {
	a, err := vm.dstack.PopByteArray()
	if err != nil {
		return err
	}

	n, err := makeScriptNum(a, false, len(a))
	if err != nil {
		return err
	}
	if len(n.Bytes()) > defaultScriptNumLen {
		return scriptError(ErrNumberTooBig,
			fmt.Sprintf("script numbers are limited to %d bytes", defaultScriptNumLen))
	}

	vm.dstack.PushInt(n)
	return nil
}

// opcodeSize pushes the size of the top item of the data stack onto the data
// stack.
//
// Stack transformation: [... x1] -> [... x1 len(x1)]
func opcodeSize(op *ParsedOp, vm *Engine) error {
	so, err := vm.dstack.PeekByteArray(0)
	if err != nil {
		return err
	}

	vm.dstack.PushInt(scriptNum(len(so)))
	return nil
}

// opcodeInvert flips all of the top stack item's bits
//
// Stack transformation: a -> ~a
func opcodeInvert(op *ParsedOp, vm *Engine) error {
	ba, err := vm.dstack.PeekByteArray(0)
	if err != nil {
		return err
	}

	for i := range ba {
		ba[i] = ba[i] ^ 0xFF
	}

	return nil
}

// opcodeAnd executes a boolean and between each bit in the operands
//
// Stack transformation: x1 x2 bscript.OpAND -> out
func opcodeAnd(op *ParsedOp, vm *Engine) error {
	a, err := vm.dstack.PopByteArray()
	if err != nil {
		return err
	}

	b, err := vm.dstack.PopByteArray()
	if err != nil {
		return err
	}

	if len(a) != len(b) {
		return scriptError(ErrInvalidInputLength, "byte arrays are not the same length")
	}

	c := make([]byte, len(a))
	for i := range a {
		c[i] = a[i] & b[i]
	}

	vm.dstack.PushByteArray(c)
	return nil
}

// opcodeOr executes a boolean or between each bit in the operands
//
// Stack transformation: x1 x2 bscript.OpOR -> out
func opcodeOr(op *ParsedOp, vm *Engine) error {
	a, err := vm.dstack.PopByteArray()
	if err != nil {
		return err
	}

	b, err := vm.dstack.PopByteArray()
	if err != nil {
		return err
	}

	if len(a) != len(b) {
		return scriptError(ErrInvalidInputLength, "byte arrays are not the same length")
	}

	c := make([]byte, len(a))
	for i := range a {
		c[i] = a[i] | b[i]
	}

	vm.dstack.PushByteArray(c)
	return nil
}

// opcodeXor executes a boolean xor between each bit in the operands
//
// Stack transformation: x1 x2 bscript.OpXOR -> out
func opcodeXor(op *ParsedOp, vm *Engine) error {
	a, err := vm.dstack.PopByteArray()
	if err != nil {
		return err
	}

	b, err := vm.dstack.PopByteArray()
	if err != nil {
		return err
	}

	if len(a) != len(b) {
		return scriptError(ErrInvalidInputLength, "byte arrays are not the same length")
	}

	c := make([]byte, len(a))
	for i := range a {
		c[i] = a[i] ^ b[i]
	}

	vm.dstack.PushByteArray(c)
	return nil
}

// opcodeEqual removes the top 2 items of the data stack, compares them as raw
// bytes, and pushes the result, encoded as a boolean, back to the stack.
//
// Stack transformation: [... x1 x2] -> [... bool]
func opcodeEqual(op *ParsedOp, vm *Engine) error {
	a, err := vm.dstack.PopByteArray()
	if err != nil {
		return err
	}

	b, err := vm.dstack.PopByteArray()
	if err != nil {
		return err
	}

	vm.dstack.PushBool(bytes.Equal(a, b))
	return nil
}

// opcodeEqualVerify is a combination of opcodeEqual and opcodeVerify.
// Specifically, it removes the top 2 items of the data stack, compares them,
// and pushes the result, encoded as a boolean, back to the stack.  Then, it
// examines the top item on the data stack as a boolean value and verifies it
// evaluates to true.  An error is returned if it does not.
//
// Stack transformation: [... x1 x2] -> [... bool] -> [...]
func opcodeEqualVerify(op *ParsedOp, vm *Engine) error {
	if err := opcodeEqual(op, vm); err != nil {
		return err
	}

	return abstractVerify(op, vm, ErrEqualVerify)
}

// opcode1Add treats the top item on the data stack as an integer and replaces
// it with its incremented value (plus 1).
//
// Stack transformation: [... x1 x2] -> [... x1 x2+1]
func opcode1Add(op *ParsedOp, vm *Engine) error {
	m, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	vm.dstack.PushInt(m + 1)
	return nil
}

// opcode1Sub treats the top item on the data stack as an integer and replaces
// it with its decremented value (minus 1).
//
// Stack transformation: [... x1 x2] -> [... x1 x2-1]
func opcode1Sub(op *ParsedOp, vm *Engine) error {
	m, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	vm.dstack.PushInt(m - 1)
	return nil
}

// opcodeNegate treats the top item on the data stack as an integer and replaces
// it with its negation.
//
// Stack transformation: [... x1 x2] -> [... x1 -x2]
func opcodeNegate(op *ParsedOp, vm *Engine) error {
	m, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	vm.dstack.PushInt(-m)
	return nil
}

// opcodeAbs treats the top item on the data stack as an integer and replaces it
// it with its absolute value.
//
// Stack transformation: [... x1 x2] -> [... x1 abs(x2)]
func opcodeAbs(op *ParsedOp, vm *Engine) error {
	m, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	if m < 0 {
		m = -m
	}

	vm.dstack.PushInt(m)
	return nil
}

// opcodeNot treats the top item on the data stack as an integer and replaces
// it with its "inverted" value (0 becomes 1, non-zero becomes 0).
//
// NOTE: While it would probably make more sense to treat the top item as a
// boolean, and push the opposite, which is really what the intention of this
// opcode is, it is extremely important that is not done because integers are
// interpreted differently than booleans and the consensus rules for this opcode
// dictate the item is interpreted as an integer.
//
// Stack transformation (x2==0): [... x1 0] -> [... x1 1]
// Stack transformation (x2!=0): [... x1 1] -> [... x1 0]
// Stack transformation (x2!=0): [... x1 17] -> [... x1 0]
func opcodeNot(op *ParsedOp, vm *Engine) error {
	m, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	var n scriptNum
	if m == 0 {
		n = 1
	}

	vm.dstack.PushInt(n)
	return nil
}

// opcode0NotEqual treats the top item on the data stack as an integer and
// replaces it with either a 0 if it is zero, or a 1 if it is not zero.
//
// Stack transformation (x2==0): [... x1 0] -> [... x1 0]
// Stack transformation (x2!=0): [... x1 1] -> [... x1 1]
// Stack transformation (x2!=0): [... x1 17] -> [... x1 1]
func opcode0NotEqual(op *ParsedOp, vm *Engine) error {
	m, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	if m != 0 {
		m = 1
	}

	vm.dstack.PushInt(m)
	return nil
}

// opcodeAdd treats the top two items on the data stack as integers and replaces
// them with their sum.
//
// Stack transformation: [... x1 x2] -> [... x1+x2]
func opcodeAdd(op *ParsedOp, vm *Engine) error {
	v0, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	v1, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	vm.dstack.PushInt(v0 + v1)
	return nil
}

// opcodeSub treats the top two items on the data stack as integers and replaces
// them with the result of subtracting the top entry from the second-to-top
// entry.
//
// Stack transformation: [... x1 x2] -> [... x1-x2]
func opcodeSub(op *ParsedOp, vm *Engine) error {
	v0, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	v1, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	vm.dstack.PushInt(v1 - v0)
	return nil
}

// opcodeMul treats the top two items on the data stack as integers and replaces
// them with the result of subtracting the top entry from the second-to-top
// entry.
func opcodeMul(op *ParsedOp, vm *Engine) error {
	n1, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	n2, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	n3 := n1.Int32() * n2.Int32()
	vm.dstack.PushInt(scriptNum(n3))

	return nil
}

// opcodeDiv return the integer quotient of a and b. If the result
// would be a non-integer it is rounded towards zero.
//
// Stack transformation: a b bscript.OpDIV -> out
func opcodeDiv(op *ParsedOp, vm *Engine) error {
	b, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	a, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	if b == 0 {
		return scriptError(ErrNumberTooSmall, "divide by zero")
	}

	vm.dstack.PushInt(a / b)
	return nil
}

// opcodeMod returns the remainder after dividing a by b. The output will
// be represented using the least number of bytes required.
//
// Stack transformation: a b bscript.OpMOD -> out
func opcodeMod(op *ParsedOp, vm *Engine) error {
	b, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	a, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	if b == 0 {
		return scriptError(ErrNumberTooSmall, "mod by zero")
	}

	vm.dstack.PushInt(a % b)
	return nil
}

func opcodeLShift(op *ParsedOp, vm *Engine) error {
	n, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	if n.Int32() < 0 {
		return scriptError(ErrNumberTooSmall, "n less than 0")
	}

	x, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	d := x.Int32() << uint32(n.Int32())
	vm.dstack.PushInt(scriptNum(d))

	return nil
}

func opcodeRShift(op *ParsedOp, vm *Engine) error {
	n, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	if n.Int32() < 0 {
		return scriptError(ErrNumberTooSmall, "n less than 0")
	}

	x, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	d := x.Int32() >> uint32(n.Int32())
	vm.dstack.PushInt(scriptNum(d))

	return nil
}

// opcodeBoolAnd treats the top two items on the data stack as integers.  When
// both of them are not zero, they are replaced with a 1, otherwise a 0.
//
// Stack transformation (x1==0, x2==0): [... 0 0] -> [... 0]
// Stack transformation (x1!=0, x2==0): [... 5 0] -> [... 0]
// Stack transformation (x1==0, x2!=0): [... 0 7] -> [... 0]
// Stack transformation (x1!=0, x2!=0): [... 4 8] -> [... 1]
func opcodeBoolAnd(op *ParsedOp, vm *Engine) error {
	v0, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	v1, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	var n scriptNum
	if v0 != 0 && v1 != 0 {
		n = 1
	}

	vm.dstack.PushInt(n)
	return nil
}

// opcodeBoolOr treats the top two items on the data stack as integers.  When
// either of them are not zero, they are replaced with a 1, otherwise a 0.
//
// Stack transformation (x1==0, x2==0): [... 0 0] -> [... 0]
// Stack transformation (x1!=0, x2==0): [... 5 0] -> [... 1]
// Stack transformation (x1==0, x2!=0): [... 0 7] -> [... 1]
// Stack transformation (x1!=0, x2!=0): [... 4 8] -> [... 1]
func opcodeBoolOr(op *ParsedOp, vm *Engine) error {
	v0, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	v1, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	var n scriptNum
	if v0 != 0 || v1 != 0 {
		n = 1
	}

	vm.dstack.PushInt(n)
	return nil
}

// opcodeNumEqual treats the top two items on the data stack as integers.  When
// they are equal, they are replaced with a 1, otherwise a 0.
//
// Stack transformation (x1==x2): [... 5 5] -> [... 1]
// Stack transformation (x1!=x2): [... 5 7] -> [... 0]
func opcodeNumEqual(op *ParsedOp, vm *Engine) error {
	v0, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	v1, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	var n scriptNum
	if v0 == v1 {
		n = 1
	}

	vm.dstack.PushInt(n)
	return nil
}

// opcodeNumEqualVerify is a combination of opcodeNumEqual and opcodeVerify.
//
// Specifically, treats the top two items on the data stack as integers.  When
// they are equal, they are replaced with a 1, otherwise a 0.  Then, it examines
// the top item on the data stack as a boolean value and verifies it evaluates
// to true.  An error is returned if it does not.
//
// Stack transformation: [... x1 x2] -> [... bool] -> [...]
func opcodeNumEqualVerify(op *ParsedOp, vm *Engine) error {
	if err := opcodeNumEqual(op, vm); err != nil {
		return err
	}

	return abstractVerify(op, vm, ErrNumEqualVerify)
}

// opcodeNumNotEqual treats the top two items on the data stack as integers.
// When they are NOT equal, they are replaced with a 1, otherwise a 0.
//
// Stack transformation (x1==x2): [... 5 5] -> [... 0]
// Stack transformation (x1!=x2): [... 5 7] -> [... 1]
func opcodeNumNotEqual(op *ParsedOp, vm *Engine) error {
	v0, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	v1, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	var n scriptNum
	if v0 != v1 {
		n = 1
	}

	vm.dstack.PushInt(n)
	return nil
}

// opcodeLessThan treats the top two items on the data stack as integers.  When
// the second-to-top item is less than the top item, they are replaced with a 1,
// otherwise a 0.
//
// Stack transformation: [... x1 x2] -> [... bool]
func opcodeLessThan(op *ParsedOp, vm *Engine) error {
	v0, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	v1, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	var n scriptNum
	if v1 < v0 {
		n = 1
	}

	vm.dstack.PushInt(n)
	return nil
}

// opcodeGreaterThan treats the top two items on the data stack as integers.
// When the second-to-top item is greater than the top item, they are replaced
// with a 1, otherwise a 0.
//
// Stack transformation: [... x1 x2] -> [... bool]
func opcodeGreaterThan(op *ParsedOp, vm *Engine) error {
	v0, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	v1, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	var n scriptNum
	if v1 > v0 {
		n = 1
	}

	vm.dstack.PushInt(n)
	return nil
}

// opcodeLessThanOrEqual treats the top two items on the data stack as integers.
// When the second-to-top item is less than or equal to the top item, they are
// replaced with a 1, otherwise a 0.
//
// Stack transformation: [... x1 x2] -> [... bool]
func opcodeLessThanOrEqual(op *ParsedOp, vm *Engine) error {
	v0, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	v1, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	var n scriptNum
	if v1 <= v0 {
		n = 1
	}

	vm.dstack.PushInt(n)
	return nil
}

// opcodeGreaterThanOrEqual treats the top two items on the data stack as
// integers.  When the second-to-top item is greater than or equal to the top
// item, they are replaced with a 1, otherwise a 0.
//
// Stack transformation: [... x1 x2] -> [... bool]
func opcodeGreaterThanOrEqual(op *ParsedOp, vm *Engine) error {
	v0, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	v1, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	var n scriptNum
	if v1 >= v0 {
		n = 1
	}

	vm.dstack.PushInt(n)
	return nil
}

// opcodeMin treats the top two items on the data stack as integers and replaces
// them with the minimum of the two.
//
// Stack transformation: [... x1 x2] -> [... min(x1, x2)]
func opcodeMin(op *ParsedOp, vm *Engine) error {
	v0, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	v1, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	n := v0
	if v1 < v0 {
		n = v1
	}

	vm.dstack.PushInt(n)
	return nil
}

// opcodeMax treats the top two items on the data stack as integers and replaces
// them with the maximum of the two.
//
// Stack transformation: [... x1 x2] -> [... max(x1, x2)]
func opcodeMax(op *ParsedOp, vm *Engine) error {
	v0, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	v1, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	n := v0
	if v1 > v0 {
		n = v1
	}

	vm.dstack.PushInt(n)
	return nil
}

// opcodeWithin treats the top 3 items on the data stack as integers.  When the
// value to test is within the specified range (left inclusive), they are
// replaced with a 1, otherwise a 0.
//
// The top item is the max value, the second-top-item is the minimum value, and
// the third-to-top item is the value to test.
//
// Stack transformation: [... x1 min max] -> [... bool]
func opcodeWithin(op *ParsedOp, vm *Engine) error {
	maxVal, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	minVal, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	x, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	var n int
	if minVal <= x && x < maxVal {
		n = 1
	}

	vm.dstack.PushInt(scriptNum(n))
	return nil
}

// calcHash calculates the hash of hasher over buf.
func calcHash(buf []byte, hasher hash.Hash) []byte {
	hasher.Write(buf) // nolint:gosec // guaranteed to be hashable
	return hasher.Sum(nil)
}

// opcodeRipemd160 treats the top item of the data stack as raw bytes and
// replaces it with ripemd160(data).
//
// Stack transformation: [... x1] -> [... ripemd160(x1)]
func opcodeRipemd160(op *ParsedOp, vm *Engine) error {
	buf, err := vm.dstack.PopByteArray()
	if err != nil {
		return err
	}

	vm.dstack.PushByteArray(calcHash(buf, ripemd160.New()))
	return nil
}

// opcodeSha1 treats the top item of the data stack as raw bytes and replaces it
// with sha1(data).
//
// Stack transformation: [... x1] -> [... sha1(x1)]
func opcodeSha1(op *ParsedOp, vm *Engine) error {
	buf, err := vm.dstack.PopByteArray()
	if err != nil {
		return err
	}

	hash := sha1.Sum(buf) // nolint:gosec // operation is for sha1
	vm.dstack.PushByteArray(hash[:])
	return nil
}

// opcodeSha256 treats the top item of the data stack as raw bytes and replaces
// it with sha256(data).
//
// Stack transformation: [... x1] -> [... sha256(x1)]
func opcodeSha256(op *ParsedOp, vm *Engine) error {
	buf, err := vm.dstack.PopByteArray()
	if err != nil {
		return err
	}

	hash := sha256.Sum256(buf)
	vm.dstack.PushByteArray(hash[:])
	return nil
}

// opcodeHash160 treats the top item of the data stack as raw bytes and replaces
// it with ripemd160(sha256(data)).
//
// Stack transformation: [... x1] -> [... ripemd160(sha256(x1))]
func opcodeHash160(op *ParsedOp, vm *Engine) error {
	buf, err := vm.dstack.PopByteArray()
	if err != nil {
		return err
	}

	hash := sha256.Sum256(buf)
	vm.dstack.PushByteArray(calcHash(hash[:], ripemd160.New()))
	return nil
}

// opcodeHash256 treats the top item of the data stack as raw bytes and replaces
// it with sha256(sha256(data)).
//
// Stack transformation: [... x1] -> [... sha256(sha256(x1))]
func opcodeHash256(op *ParsedOp, vm *Engine) error {
	buf, err := vm.dstack.PopByteArray()
	if err != nil {
		return err
	}

	vm.dstack.PushByteArray(crypto.Sha256d(buf))
	return nil
}

// opcodeCodeSeparator stores the current script offset as the most recently
// seen bscript.OpCODESEPARATOR which is used during signature checking.
//
// This opcode does not change the contents of the data stack.
func opcodeCodeSeparator(op *ParsedOp, vm *Engine) error {
	vm.lastCodeSep = vm.scriptOff
	return nil
}

// opcodeCheckSig treats the top 2 items on the stack as a public key and a
// signature and replaces them with a bool which indicates if the signature was
// successfully verified.
//
// The process of verifying a signature requires calculating a signature hash in
// the same way the transaction signer did.  It involves hashing portions of the
// transaction based on the hash type byte (which is the final byte of the
// signature) and the portion of the script starting from the most recent
// bscript.OpCODESEPARATOR (or the beginning of the script if there are none) to the
// end of the script (with any other bscript.OpCODESEPARATORs removed).  Once this
// "script hash" is calculated, the signature is checked using standard
// cryptographic methods against the provided public key.
//
// Stack transformation: [... signature pubkey] -> [... bool]
func opcodeCheckSig(op *ParsedOp, vm *Engine) error {
	pkBytes, err := vm.dstack.PopByteArray()
	if err != nil {
		return err
	}

	fullSigBytes, err := vm.dstack.PopByteArray()
	if err != nil {
		return err
	}

	// The signature actually needs needs to be longer than this, but at
	// least 1 byte is needed for the hash type below.  The full length is
	// checked depending on the script flags and upon parsing the signature.
	if len(fullSigBytes) < 1 {
		vm.dstack.PushBool(false)
		return nil
	}

	// Trim off hashtype from the signature string and check if the
	// signature and pubkey conform to the strict encoding requirements
	// depending on the flags.
	//
	// NOTE: When the strict encoding flags are set, any errors in the
	// signature or public encoding here result in an immediate script error
	// (and thus no result bool is pushed to the data stack).  This differs
	// from the logic below where any errors in parsing the signature is
	// treated as the signature failure resulting in false being pushed to
	// the data stack.  This is required because the more general script
	// validation consensus rules do not have the new strict encoding
	// requirements enabled by the flags.
	shf := sighash.Flag(fullSigBytes[len(fullSigBytes)-1])
	sigBytes := fullSigBytes[:len(fullSigBytes)-1]
	if err = vm.checkHashTypeEncoding(shf); err != nil {
		return err
	}
	if err = vm.checkSignatureEncoding(sigBytes); err != nil {
		return err
	}
	if err = vm.checkPubKeyEncoding(pkBytes); err != nil {
		return err
	}

	// Get script starting from the most recent bscript.OpCODESEPARATOR.
	subScript := vm.subScript()

	// Generate the signature hash based on the signature hash type.
	var hash []byte

	// Remove the signature since there is no way for a signature
	// to sign itself.
	subScript = subScript.removeOpcodeByData(fullSigBytes)

	up, err := vm.scriptParser.Unparse(subScript)
	if err != nil {
		return err
	}

	//sigHashes := NewTxSigHashes(vm.tx)

	txCopy := vm.tx.Clone()
	txCopy.Inputs[vm.inputIdx].PreviousTxScript = up

	hash, err = txCopy.CalcInputSignatureHash(uint32(vm.inputIdx), shf)
	if err != nil {
		vm.dstack.PushBool(false)
		return err
	}

	pubKey, err := bec.ParsePubKey(pkBytes, bec.S256())
	if err != nil {
		vm.dstack.PushBool(false)
		return nil //nolint:nilerr // unexpected behaviour
	}

	var signature *bec.Signature
	if vm.hasFlag(ScriptVerifyStrictEncoding) || vm.hasFlag(ScriptVerifyDERSignatures) {
		signature, err = bec.ParseDERSignature(sigBytes, bec.S256())
	} else {
		signature, err = bec.ParseSignature(sigBytes, bec.S256())
	}
	if err != nil {
		vm.dstack.PushBool(false)
		return nil //nolint:nilerr // unexpected behaviour
	}

	var valid bool
	if vm.sigCache != nil {
		valid = vm.sigCache.Exists(hash, signature, pubKey)
		if !valid && signature.Verify(hash, pubKey) {
			vm.sigCache.Add(hash, signature, pubKey)
			valid = true
		}
	} else {
		valid = signature.Verify(hash, pubKey)
	}

	if !valid && vm.hasFlag(ScriptVerifyNullFail) && len(sigBytes) > 0 {
		return scriptError(ErrNullFail, "signature not empty on failed checksig")
	}

	vm.dstack.PushBool(valid)
	return nil
}

// opcodeCheckSigVerify is a combination of opcodeCheckSig and opcodeVerify.
// The opcodeCheckSig function is invoked followed by opcodeVerify.  See the
// documentation for each of those opcodes for more details.
//
// Stack transformation: signature pubkey] -> [... bool] -> [...]
func opcodeCheckSigVerify(op *ParsedOp, vm *Engine) error {
	if err := opcodeCheckSig(op, vm); err != nil {
		return err
	}

	return abstractVerify(op, vm, ErrCheckSigVerify)
}

// parsedSigInfo houses a raw signature along with its parsed form and a flag
// for whether or not it has already been parsed.  It is used to prevent parsing
// the same signature multiple times when verifying a multisig.
type parsedSigInfo struct {
	signature       []byte
	parsedSignature *bec.Signature
	parsed          bool
}

// opcodeCheckMultiSig treats the top item on the stack as an integer number of
// public keys, followed by that many entries as raw data representing the public
// keys, followed by the integer number of signatures, followed by that many
// entries as raw data representing the signatures.
//
// Due to a bug in the original Satoshi client implementation, an additional
// dummy argument is also required by the consensus rules, although it is not
// used.  The dummy value SHOULD be an bscript.Op0, although that is not required by
// the consensus rules.  When the ScriptStrictMultiSig flag is set, it must be
// bscript.Op0.
//
// All of the aforementioned stack items are replaced with a bool which
// indicates if the requisite number of signatures were successfully verified.
//
// See the opcodeCheckSigVerify documentation for more details about the process
// for verifying each signature.
//
// Stack transformation:
// [... dummy [sig ...] numsigs [pubkey ...] numpubkeys] -> [... bool]
func opcodeCheckMultiSig(op *ParsedOp, vm *Engine) error {
	numKeys, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	numPubKeys := int(numKeys.Int32())
	if numPubKeys < 0 {
		return scriptError(ErrInvalidPubKeyCount, "number of pubkeys %d is negative", numPubKeys)
	}
	if numPubKeys > bscript.MaxPubKeysPerMultiSig {
		return scriptError(ErrInvalidPubKeyCount, "too many pubkeys: %d > %d", numPubKeys, bscript.MaxPubKeysPerMultiSig)
	}
	vm.numOps += numPubKeys
	if vm.numOps > bscript.MaxOps {
		return scriptError(ErrTooManyOperations, "exceeded max operation limit of %d", bscript.MaxOps)
	}

	pubKeys := make([][]byte, 0, numPubKeys)
	for i := 0; i < numPubKeys; i++ {
		pubKey, err := vm.dstack.PopByteArray() //nolint:govet // ignore shadowed error
		if err != nil {
			return err
		}
		pubKeys = append(pubKeys, pubKey)
	}

	numSigs, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	numSignatures := int(numSigs.Int32())
	if numSignatures < 0 {
		return scriptError(ErrInvalidSignatureCount, "number of signatures %d is negative", numSignatures)
	}
	if numSignatures > numPubKeys {
		return scriptError(ErrInvalidSignatureCount, "more signatures than pubkeys: %d > %d", numSignatures, numPubKeys)
	}

	signatures := make([]*parsedSigInfo, 0, numSignatures)
	for i := 0; i < numSignatures; i++ {
		signature, err := vm.dstack.PopByteArray() //nolint:govet // ignore shadowed error
		if err != nil {
			return err
		}
		sigInfo := &parsedSigInfo{signature: signature}
		signatures = append(signatures, sigInfo)
	}

	// A bug in the original Satoshi client implementation means one more
	// stack value than should be used must be popped.  Unfortunately, this
	// buggy behaviour is now part of the consensus and a hard fork would be
	// required to fix it.
	dummy, err := vm.dstack.PopByteArray()
	if err != nil {
		return err
	}

	// Since the dummy argument is otherwise not checked, it could be any
	// value which unfortunately provides a source of malleability.  Thus,
	// there is a script flag to force an error when the value is NOT 0.
	if vm.hasFlag(ScriptStrictMultiSig) && len(dummy) != 0 {
		return scriptError(ErrSigNullDummy, "multisig dummy argument has length %d instead of 0", len(dummy))
	}

	// Get script starting from the most recent bscript.OpCODESEPARATOR.
	script := vm.subScript()

	for _, sigInfo := range signatures {
		script = script.removeOpcodeByData(sigInfo.signature)
	}

	success := true
	numPubKeys++
	pubKeyIdx := -1
	signatureIdx := 0
	//var sigHashes *TxSigHashes
	//if vm.hasFlag(ScriptVerifyBip143SigHash) {
	//	//sigHashes = NewTxSigHashes(vm.tx)
	//}
	for numSignatures > 0 {
		// When there are more signatures than public keys remaining,
		// there is no way to succeed since too many signatures are
		// invalid, so exit early.
		pubKeyIdx++
		numPubKeys--
		if numSignatures > numPubKeys {
			success = false
			break
		}

		sigInfo := signatures[signatureIdx]
		pubKey := pubKeys[pubKeyIdx]

		// The order of the signature and public key evaluation is
		// important here since it can be distinguished by an
		// bscript.OpCHECKMULTISIG NOT when the strict encoding flag is set.

		rawSig := sigInfo.signature
		if len(rawSig) == 0 {
			// Skip to the next pubkey if signature is empty.
			continue
		}

		// Split the signature into hash type and signature components.
		shf := sighash.Flag(rawSig[len(rawSig)-1])
		signature := rawSig[:len(rawSig)-1]

		// Only parse and check the signature encoding once.
		var parsedSig *bec.Signature
		if !sigInfo.parsed {
			if err := vm.checkHashTypeEncoding(shf); err != nil {
				return err
			}
			if err := vm.checkSignatureEncoding(signature); err != nil {
				return err
			}

			// Parse the signature.
			var err error
			if vm.hasFlag(ScriptVerifyStrictEncoding) ||
				vm.hasFlag(ScriptVerifyDERSignatures) {

				parsedSig, err = bec.ParseDERSignature(signature,
					bec.S256())
			} else {
				parsedSig, err = bec.ParseSignature(signature,
					bec.S256())
			}
			sigInfo.parsed = true
			if err != nil {
				continue
			}
			sigInfo.parsedSignature = parsedSig
		} else {
			// Skip to the next pubkey if the signature is invalid.
			if sigInfo.parsedSignature == nil {
				continue
			}

			// Use the already parsed signature.
			parsedSig = sigInfo.parsedSignature
		}

		if err := vm.checkPubKeyEncoding(pubKey); err != nil {
			return err
		}

		// Parse the pubkey.
		parsedPubKey, err := bec.ParsePubKey(pubKey, bec.S256())
		if err != nil {
			continue
		}

		up, err := vm.scriptParser.Unparse(script)
		if err != nil {
			vm.dstack.PushBool(false)
			return nil //nolint:nilerr // unexpected behaviour
		}

		// Generate the signature hash based on the signature hash type.
		txCopy := vm.tx.Clone()
		txCopy.Inputs[vm.inputIdx].PreviousTxScript = up

		signatureHash, err := txCopy.CalcInputSignatureHash(uint32(vm.inputIdx), shf)
		if err != nil {
			vm.dstack.PushBool(false)
			return nil //nolint:nilerr // unexpected behaviour
		}

		var valid bool
		if vm.sigCache != nil {
			valid = vm.sigCache.Exists(signatureHash, parsedSig, parsedPubKey)
			if !valid && parsedSig.Verify(signatureHash, parsedPubKey) {
				vm.sigCache.Add(signatureHash, parsedSig, parsedPubKey)
				valid = true
			}
		} else {
			valid = parsedSig.Verify(signatureHash, parsedPubKey)
		}

		if valid {
			// PubKey verified, move on to the next signature.
			signatureIdx++
			numSignatures--
		}
	}

	if !success && vm.hasFlag(ScriptVerifyNullFail) {
		for _, sig := range signatures {
			if len(sig.signature) > 0 {
				return scriptError(ErrNullFail, "not all signatures empty on failed checkmultisig")
			}
		}
	}

	vm.dstack.PushBool(success)
	return nil
}

// opcodeCheckMultiSigVerify is a combination of opcodeCheckMultiSig and
// opcodeVerify.  The opcodeCheckMultiSig is invoked followed by opcodeVerify.
// See the documentation for each of those opcodes for more details.
//
// Stack transformation:
// [... dummy [sig ...] numsigs [pubkey ...] numpubkeys] -> [... bool] -> [...]
func opcodeCheckMultiSigVerify(op *ParsedOp, vm *Engine) error {
	if err := opcodeCheckMultiSig(op, vm); err != nil {
		return err
	}

	return abstractVerify(op, vm, ErrCheckMultiSigVerify)
}
