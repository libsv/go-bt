// Copyright (c) 2013-2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package interpreter

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"testing"

	"github.com/libsv/go-bt/v2"
	"github.com/libsv/go-bt/v2/bscript"
	"github.com/libsv/go-bt/v2/bscript/interpreter/errs"
	"github.com/libsv/go-bt/v2/bscript/interpreter/scriptflag"
)

var opcodeByName = make(map[string]byte)

func init() {
	// Initialise the opcode name to value map using the contents of the
	// opcode array.  Also add entries for "OP_FALSE", "OP_TRUE", and
	// "OP_NOP2" since they are aliases for "OP_0", "OP_1",
	// and "OP_CHECKLOCKTIMEVERIFY" respectively.
	for _, op := range &opcodeArray {
		opcodeByName[op.Name()] = op.val
	}
	opcodeByName["OP_0"] = bscript.Op0
	opcodeByName["OP_1"] = bscript.Op1
	opcodeByName["OP_CHECKLOCKTIMEVERIFY"] = bscript.OpCHECKLOCKTIMEVERIFY
	opcodeByName["OP_CHECKSEQUENCEVERIFY"] = bscript.OpCHECKSEQUENCEVERIFY
	opcodeByName["OP_RESERVED"] = bscript.OpRESERVED

}

// parseShortForm parses a string as as used in the Bitcoin Core reference tests
// into the script it came from.
//
// The format used for these tests is pretty simple if ad-hoc:
//   - Opcodes other than the push opcodes and unknown are present as
//     either OP_NAME or just NAME
//   - Plain numbers are made into push operations
//   - Numbers beginning with 0x are inserted into the []byte as-is (so
//     0x14 is OP_DATA_20)
//   - Single quoted strings are pushed as data
//   - Anything else is an error
func parseShortForm(script string) (*bscript.Script, error) {
	// Only create the short form opcode map once.
	if shortFormOps == nil {
		ops := make(map[string]byte)
		for opcodeName, opcodeValue := range opcodeByName {
			if strings.Contains(opcodeName, "OP_UNKNOWN") {
				continue
			}
			ops[opcodeName] = opcodeValue

			// The opcodes named OP_# can't have the OP_ prefix
			// stripped or they would conflict with the plain
			// numbers.  Also, since OP_FALSE and OP_TRUE are
			// aliases for the OP_0, and OP_1, respectively, they
			// have the same value, so detect those by name and
			// allow them.
			if (opcodeName == "OP_FALSE" || opcodeName == "OP_TRUE") ||
				(opcodeValue != bscript.Op0 && (opcodeValue < bscript.Op1 ||
					opcodeValue > bscript.Op16)) {

				ops[strings.TrimPrefix(opcodeName, "OP_")] = opcodeValue
			}
		}
		shortFormOps = ops
	}

	// Split only does one separator so convert all \n and tab into  space.
	script = strings.Replace(script, "\n", " ", -1)
	script = strings.Replace(script, "\t", " ", -1)
	tokens := strings.Split(script, " ")

	var scr bscript.Script
	for _, tok := range tokens {
		if len(tok) == 0 {
			continue
		}
		// if parses as a plain number
		if num, err := strconv.ParseInt(tok, 10, 64); err == nil {
			if num == 0 {
				scr.AppendOpCode(bscript.Op0)
			} else if num == -1 || (1 <= num && num <= 16) {
				scr.AppendOpCode((bscript.Op1 - 1) + byte(num))
			} else {
				scr.AppendPushData(scriptNum(num).Bytes())
			}
			continue
		} else if bts, err := parseHex(tok); err == nil {
			// Concatenate the bytes manually since the test code
			// intentionally creates scripts that are too large and
			// would cause the builder to error otherwise.
			scr = append(scr, bts...)
		} else if len(tok) >= 2 &&
			tok[0] == '\'' && tok[len(tok)-1] == '\'' {
			scr.AppendPushData([]byte(tok[1 : len(tok)-1]))
		} else if opcode, ok := shortFormOps[tok]; ok {
			scr.AppendOpCode(opcode)
		} else {
			return nil, fmt.Errorf("bad token %q", tok)
		}

	}

	return &scr, nil
}

// scriptTestName returns a descriptive test name for the given reference script
// test data.
func scriptTestName(test []interface{}) (string, error) {

	// The test must consist of at least a signature script, public key script,
	// flags, and expected error.  Finally, it may optionally contain a comment.
	if len(test) < 4 || 6 < len(test) {
		fmt.Printf("%#v\n", test)
		return "", fmt.Errorf("invalid test length %d", len(test))
	}

	// Use the comment for the test name if one is specified, otherwise,
	// construct the name based on the signature script, public key script,
	// and flags.
	var name string
	if len(test) >= 5 {
		name = fmt.Sprintf("test (%s)", test[len(test)-1])
	} else {
		name = fmt.Sprintf("test ([%s, %s, %s])", test[0],
			test[1], test[2])
	}
	return name, nil
}

// parse hex string into a []byte.
func parseHex(tok string) ([]byte, error) {
	if !strings.HasPrefix(tok, "0x") {
		return nil, errors.New("not a hex number")
	}
	return hex.DecodeString(tok[2:])
}

// shortFormOps holds a map of opcode names to values for use in short form
// parsing.  It is declared here so it only needs to be created once.
var shortFormOps map[string]byte

// parseScriptFlags parses the provided flags string from the format used in the
// reference tests into ScriptFlags suitable for use in the script engine.
func parseScriptFlags(flagStr string) (scriptflag.Flag, error) {
	var flags scriptflag.Flag

	sFlags := strings.Split(flagStr, ",")
	for _, flag := range sFlags {
		switch flag {
		case "":
			// Nothing.
		case "CHECKLOCKTIMEVERIFY":
			flags |= scriptflag.VerifyCheckLockTimeVerify
		case "CHECKSEQUENCEVERIFY":
			flags |= scriptflag.VerifyCheckSequenceVerify
		case "CLEANSTACK":
			flags |= scriptflag.VerifyCleanStack
		case "DERSIG":
			flags |= scriptflag.VerifyDERSignatures
		case "DISCOURAGE_UPGRADABLE_NOPS":
			flags |= scriptflag.DiscourageUpgradableNops
		case "LOW_S":
			flags |= scriptflag.VerifyLowS
		case "MINIMALDATA":
			flags |= scriptflag.VerifyMinimalData
		case "NONE":
			// Nothing.
		case "NULLDUMMY":
			flags |= scriptflag.StrictMultiSig
		case "NULLFAIL":
			flags |= scriptflag.VerifyNullFail
		case "P2SH":
			flags |= scriptflag.Bip16
		case "SIGPUSHONLY":
			flags |= scriptflag.VerifySigPushOnly
		case "STRICTENC":
			flags |= scriptflag.VerifyStrictEncoding
		case "UTXO_AFTER_GENESIS":
			flags |= scriptflag.UTXOAfterGenesis
		case "MINIMALIF":
			flags |= scriptflag.VerifyMinimalIf
		case "SIGHASH_FORKID":
			flags |= scriptflag.EnableSighashForkID
		default:
			return flags, fmt.Errorf("invalid flag: %s", flag)
		}
	}
	return flags, nil
}

// parseExpectedResult parses the provided expected result string into allowed
// script error codes.  An error is returned if the expected result string is
// not supported.
func parseExpectedResult(expected string) ([]errs.ErrorCode, error) {
	switch expected {
	case "OK":
		return nil, nil
	case "INVALID_NUMBER_RANGE", "SPLIT_RANGE":
		return []errs.ErrorCode{errs.ErrNumberTooBig, errs.ErrNumberTooSmall}, nil
	case "OPERAND_SIZE":
		return []errs.ErrorCode{errs.ErrInvalidInputLength}, nil
	case "PUBKEYTYPE":
		return []errs.ErrorCode{errs.ErrPubKeyType}, nil
	case "SIG_DER":
		return []errs.ErrorCode{errs.ErrSigTooShort, errs.ErrSigTooLong,
			errs.ErrSigInvalidSeqID, errs.ErrSigInvalidDataLen, errs.ErrSigMissingSTypeID,
			errs.ErrSigMissingSLen, errs.ErrSigInvalidSLen,
			errs.ErrSigInvalidRIntID, errs.ErrSigZeroRLen, errs.ErrSigNegativeR,
			errs.ErrSigTooMuchRPadding, errs.ErrSigInvalidSIntID,
			errs.ErrSigZeroSLen, errs.ErrSigNegativeS, errs.ErrSigTooMuchSPadding,
			errs.ErrInvalidSigHashType}, nil
	case "EVAL_FALSE":
		return []errs.ErrorCode{errs.ErrEvalFalse, errs.ErrEmptyStack}, nil
	case "EQUALVERIFY":
		return []errs.ErrorCode{errs.ErrEqualVerify}, nil
	case "NULLFAIL":
		return []errs.ErrorCode{errs.ErrNullFail}, nil
	case "SIG_HIGH_S":
		return []errs.ErrorCode{errs.ErrSigHighS}, nil
	case "SIG_HASHTYPE":
		return []errs.ErrorCode{errs.ErrInvalidSigHashType}, nil
	case "SIG_NULLDUMMY":
		return []errs.ErrorCode{errs.ErrSigNullDummy}, nil
	case "SIG_PUSHONLY":
		return []errs.ErrorCode{errs.ErrNotPushOnly}, nil
	case "CLEANSTACK":
		return []errs.ErrorCode{errs.ErrCleanStack}, nil
	case "BAD_OPCODE":
		return []errs.ErrorCode{errs.ErrReservedOpcode, errs.ErrMalformedPush}, nil
	case "UNBALANCED_CONDITIONAL":
		return []errs.ErrorCode{errs.ErrUnbalancedConditional,
			errs.ErrInvalidStackOperation}, nil
	case "OP_RETURN":
		return []errs.ErrorCode{errs.ErrEarlyReturn}, nil
	case "VERIFY":
		return []errs.ErrorCode{errs.ErrVerify}, nil
	case "INVALID_STACK_OPERATION", "INVALID_ALTSTACK_OPERATION":
		return []errs.ErrorCode{errs.ErrInvalidStackOperation}, nil
	case "DISABLED_OPCODE":
		return []errs.ErrorCode{errs.ErrDisabledOpcode}, nil
	case "DISCOURAGE_UPGRADABLE_NOPS":
		return []errs.ErrorCode{errs.ErrDiscourageUpgradableNOPs}, nil
	case "SCRIPTNUM_OVERFLOW":
		return []errs.ErrorCode{errs.ErrNumberTooBig}, nil
	case "NUMBER_SIZE":
		return []errs.ErrorCode{errs.ErrNumberTooBig, errs.ErrNumberTooSmall}, nil
	case "PUSH_SIZE":
		return []errs.ErrorCode{errs.ErrElementTooBig}, nil
	case "OP_COUNT":
		return []errs.ErrorCode{errs.ErrTooManyOperations}, nil
	case "STACK_SIZE":
		return []errs.ErrorCode{errs.ErrStackOverflow}, nil
	case "SCRIPT_SIZE":
		return []errs.ErrorCode{errs.ErrScriptTooBig}, nil
	case "ELEMENT_SIZE":
		return []errs.ErrorCode{errs.ErrElementTooBig}, nil
	case "PUBKEY_COUNT":
		return []errs.ErrorCode{errs.ErrInvalidPubKeyCount}, nil
	case "SIG_COUNT":
		return []errs.ErrorCode{errs.ErrInvalidSignatureCount}, nil
	case "MINIMALDATA":
		return []errs.ErrorCode{errs.ErrMinimalData}, nil
	case "MINIMALIF":
		return []errs.ErrorCode{errs.ErrMinimalIf}, nil
	case "NEGATIVE_LOCKTIME":
		return []errs.ErrorCode{errs.ErrNegativeLockTime}, nil
	case "UNSATISFIED_LOCKTIME":
		return []errs.ErrorCode{errs.ErrUnsatisfiedLockTime}, nil
	case "SCRIPTNUM_MINENCODE":
		return []errs.ErrorCode{errs.ErrMinimalData}, nil
	case "DIV_BY_ZERO", "MOD_BY_ZERO":
		return []errs.ErrorCode{errs.ErrDivideByZero}, nil
	case "CHECKSIGVERIFY":
		return []errs.ErrorCode{errs.ErrCheckSigVerify}, nil
	case "ILLEGAL_FORKID":
		return []errs.ErrorCode{errs.ErrIllegalForkID}, nil
	}

	return nil, fmt.Errorf("unrecognised expected result in test data: %v",
		expected)
}

// createSpendTx generates a basic spending transaction given the passed
// signature and locking scripts.
func createSpendingTx(sigScript, pkScript *bscript.Script, outputValue int64) *bt.Tx {

	coinbaseTx := &bt.Tx{
		Version:  1,
		LockTime: 0,
		Inputs: []*bt.Input{{
			PreviousTxOutIndex: ^uint32(0),
			UnlockingScript:    bscript.NewFromBytes([]byte{bscript.Op0, bscript.Op0}),
			SequenceNumber:     0xffffffff,
		}},
		Outputs: []*bt.Output{{
			Satoshis:      uint64(outputValue),
			LockingScript: pkScript,
		}},
	}
	coinbaseTx.Inputs[0].PreviousTxIDAdd(make([]byte, 32))

	spendingTx := &bt.Tx{
		Version:  1,
		LockTime: 0,
		Inputs: []*bt.Input{{
			PreviousTxOutIndex: 0,
			PreviousTxScript:   pkScript,
			UnlockingScript:    sigScript,
			SequenceNumber:     0xffffffff,
		}},
		Outputs: []*bt.Output{{
			Satoshis:      uint64(outputValue),
			LockingScript: bscript.NewFromBytes([]byte{}),
		}},
	}
	spendingTx.Inputs[0].PreviousTxIDAdd(coinbaseTx.TxIDBytes())

	return spendingTx
}

// TestScripts ensures all of the tests in script_tests.json execute with the
// expected results as defined in the test data.
func TestScripts(t *testing.T) {
	file, err := ioutil.ReadFile("data/script_tests.json")
	if err != nil {
		t.Fatalf("TestScripts: %v\n", err)
	}

	var tests [][]interface{}
	err = json.Unmarshal(file, &tests)
	if err != nil {
		t.Fatalf("TestScripts couldn't Unmarshal: %v", err)
	}

	// Create a signature cache to use only if requested.
	for i, test := range tests {
		// "Format is: [[wit..., amount]?, scriptSig, scriptPubKey,
		//    flags, expected_scripterror, ... comments]"

		// Skip single line comments.
		if len(test) == 1 {
			continue
		}

		// Construct a name for the test based on the comment and test
		// data.
		name, err := scriptTestName(test)
		if err != nil {
			t.Errorf("TestScripts: invalid test #%d: %v", i, err)
			continue
		}

		var inputAmt int64
		if v, ok := test[0].([]interface{}); ok {
			if f, ok := v[0].(float64); ok {
				inputAmt = int64(f * 100000000)
			}

			test = test[1:]
		}

		// Extract and parse the signature script from the test fields.
		scriptSigStr, ok := test[0].(string)
		if !ok {
			t.Errorf("%s: signature script is not a string", name)
			continue
		}
		scriptSig, err := parseShortForm(scriptSigStr)
		if err != nil {
			t.Errorf("%s: can't parse signature script: %v", name,
				err)
			continue
		}

		// Extract and parse the public key script from the test fields.
		scriptPubKeyStr, ok := test[1].(string)
		if !ok {
			t.Errorf("%s: public key script is not a string", name)
			continue
		}
		scriptPubKey, err := parseShortForm(scriptPubKeyStr)
		if err != nil {
			t.Errorf("%s: can't parse public key script: %v", name,
				err)
			continue
		}

		// Extract and parse the script flags from the test fields.
		flagsStr, ok := test[2].(string)
		if !ok {
			t.Errorf("%s: flags field is not a string", name)
			continue
		}
		flags, err := parseScriptFlags(flagsStr)
		if err != nil {
			t.Errorf("%s: %v", name, err)
			continue
		}

		// Extract and parse the expected result from the test fields.
		//
		// Convert the expected result string into the allowed script
		// error codes.  This is necessary because interpreter is more
		// fine grained with its errors than the reference test data, so
		// some of the reference test data errors map to more than one
		// possibility.
		resultStr, ok := test[3].(string)
		if !ok {
			t.Errorf("%s: result field is not a string", name)
			continue
		}
		allowedErrorCodes, err := parseExpectedResult(resultStr)
		if err != nil {
			t.Errorf("%s: %v", name, err)
			continue
		}

		// Generate a transaction pair such that one spends from the
		// other and the provided signature and public key scripts are
		// used, then create a new engine to execute the scripts.
		tx := createSpendingTx(scriptSig, scriptPubKey,
			inputAmt)

		err = NewEngine().Execute(ExecutionParams{
			PreviousTxOut: &bt.Output{LockingScript: scriptPubKey, Satoshis: uint64(inputAmt)},
			Tx:            tx,
			InputIdx:      0,
			Flags:         flags,
		})

		// Ensure there were no errors when the expected result is OK.
		if resultStr == "OK" {
			if err != nil {
				t.Errorf("%s failed to execute: %v", name, err)
			}
			continue
		}

		// At this point an error was expected so ensure the result of
		// the execution matches it.
		success := false
		for _, code := range allowedErrorCodes {
			if errs.IsErrorCode(err, code) {
				success = true
				break
			}
		}
		if !success {
			serr := &errs.Error{}
			if ok := errors.As(err, serr); ok {
				t.Errorf("%s: want error codes %v, got %v", name, allowedErrorCodes, serr.ErrorCode)
				continue
			}
			t.Errorf("%s: want error codes %v, got err: %v (%T)", name, allowedErrorCodes, err, err)
			continue
		}
	}
}

// testVecF64ToUint32 properly handles conversion of float64s read from the JSON
// test data to unsigned 32-bit integers.  This is necessary because some of the
// test data uses -1 as a shortcut to mean max uint32 and direct conversion of a
// negative float to an unsigned int is implementation dependent and therefore
// doesn't result in the expected value on all platforms.  This function woks
// around that limitation by converting to a 32-bit signed integer first and
// then to a 32-bit unsigned integer which results in the expected behaviour on
// all platforms.
func testVecF64ToUint32(f float64) uint32 {
	return uint32(int32(f))
}

type txIOKey struct {
	id  string
	idx uint32
}

// TestTxInvalidTests ensures all of the tests in tx_invalid.json fail as
// expected.
func TestTxInvalidTests(t *testing.T) {
	file, err := ioutil.ReadFile("data/tx_invalid.json")
	if err != nil {
		t.Fatalf("TestTxInvalidTests: %v\n", err)
	}

	var tests [][]interface{}
	err = json.Unmarshal(file, &tests)
	if err != nil {
		t.Fatalf("TestTxInvalidTests couldn't Unmarshal: %v\n", err)
	}

	// form is either:
	//   ["this is a comment "]
	// or:
	//   [[[previous hash, previous index, previous scriptPubKey]...,]
	//	serializedTransaction, verifyFlags]
testloop:
	for i, test := range tests {
		inputs, ok := test[0].([]interface{})
		if !ok {
			continue
		}

		if len(test) != 3 {
			t.Errorf("bad test (bad length) %d: %v", i, test)
			continue

		}
		serializedhex, ok := test[1].(string)
		if !ok {
			t.Errorf("bad test (arg 2 not string) %d: %v", i, test)
			continue
		}
		serializedTx, err := hex.DecodeString(serializedhex)
		if err != nil {
			t.Errorf("bad test (arg 2 not hex %v) %d: %v", err, i,
				test)
			continue
		}

		tx, err := bt.NewTxFromBytes(serializedTx)
		if err != nil {
			t.Errorf("bad test (arg 2 not msgtx %v) %d: %v", err,
				i, test)
			continue
		}

		verifyFlags, ok := test[2].(string)
		if !ok {
			t.Errorf("bad test (arg 3 not string) %d: %v", i, test)
			continue
		}

		flags, err := parseScriptFlags(verifyFlags)
		if err != nil {
			t.Errorf("bad test %d: %v", i, err)
			continue
		}

		prevOuts := make(map[txIOKey]*bt.Output)
		for j, iinput := range inputs {
			input, ok := iinput.([]interface{})
			if !ok {
				t.Errorf("bad test (%dth input not array)"+
					"%d: %v", j, i, test)
				continue testloop
			}

			if len(input) < 3 || len(input) > 4 {
				t.Errorf("bad test (%dth input wrong length)"+
					"%d: %v", j, i, test)
				continue testloop
			}

			previoustx, ok := input[0].(string)
			if !ok {
				t.Errorf("bad test (%dth input hash not string)"+
					"%d: %v", j, i, test)
				continue testloop
			}

			idxf, ok := input[1].(float64)
			if !ok {
				t.Errorf("bad test (%dth input idx not number)"+
					"%d: %v", j, i, test)
				continue testloop
			}
			idx := testVecF64ToUint32(idxf)

			oscript, ok := input[2].(string)
			if !ok {
				t.Errorf("bad test (%dth input script not "+
					"string) %d: %v", j, i, test)
				continue testloop
			}

			script, err := parseShortForm(oscript)
			if err != nil {
				t.Errorf("bad test (%dth input script doesn't "+
					"parse %v) %d: %v", j, err, i, test)
				continue testloop
			}

			var inputValue float64
			if len(input) == 4 {
				inputValue, ok = input[3].(float64)
				if !ok {
					t.Errorf("bad test (%dth input value not int) "+
						"%d: %v", j, i, test)
					continue
				}
			}

			v := &bt.Output{
				Satoshis:      uint64(inputValue),
				LockingScript: script,
			}
			prevOuts[txIOKey{id: previoustx, idx: idx}] = v
		}

		for k, txin := range tx.Inputs {
			prevOut, ok := prevOuts[txIOKey{id: txin.PreviousTxIDStr(), idx: txin.PreviousTxOutIndex}]
			if !ok {
				t.Errorf("bad test (missing %dth input) %d:%v",
					k, i, test)
				continue testloop
			}
			// These are meant to fail, so as soon as the first
			// input fails the transaction has failed. (some of the
			// test txns have good inputs, too..
			err := NewEngine().Execute(ExecutionParams{
				PreviousTxOut: prevOut,
				Tx:            tx,
				InputIdx:      k,
				Flags:         flags,
			})
			if err != nil {
				continue testloop
			}
		}
		t.Errorf("test (%d:%v) succeeded when should fail",
			i, test)
	}
}

// TestTxValidTests ensures all of the tests in tx_valid.json pass as expected.
func TestTxValidTests(t *testing.T) {
	file, err := ioutil.ReadFile("data/tx_valid.json")
	if err != nil {
		t.Fatalf("TestTxValidTests: %v\n", err)
	}

	var tests [][]interface{}
	err = json.Unmarshal(file, &tests)
	if err != nil {
		t.Fatalf("TestTxValidTests couldn't Unmarshal: %v\n", err)
	}

	// form is either:
	//   ["this is a comment "]
	// or:
	//   [[[previous hash, previous index, previous scriptPubKey, input value]...,]
	//	serializedTransaction, verifyFlags]
testloop:
	for i, test := range tests {
		inputs, ok := test[0].([]interface{})
		if !ok {
			continue
		}

		if len(test) != 3 {
			t.Errorf("bad test (bad length) %d: %v", i, test)
			continue
		}
		serializedhex, ok := test[1].(string)
		if !ok {
			t.Errorf("bad test (arg 2 not string) %d: %v", i, test)
			continue
		}
		serializedTx, err := hex.DecodeString(serializedhex)
		if err != nil {
			t.Errorf("bad test (arg 2 not hex %v) %d: %v", err, i,
				test)
			continue
		}

		tx, err := bt.NewTxFromBytes(serializedTx)
		if err != nil {
			t.Errorf("bad test (arg 2 not msgtx %v) %d: %v", err,
				i, test)
			continue
		}

		verifyFlags, ok := test[2].(string)
		if !ok {
			t.Errorf("bad test (arg 3 not string) %d: %v", i, test)
			continue
		}

		flags, err := parseScriptFlags(verifyFlags)
		if err != nil {
			t.Errorf("bad test %d: %v", i, err)
			continue
		}

		prevOuts := make(map[txIOKey]*bt.Output)
		for j, iinput := range inputs {
			input, ok := iinput.([]interface{})
			if !ok {
				t.Errorf("bad test (%dth input not array)"+
					"%d: %v", j, i, test)
				continue
			}

			if len(input) < 3 || len(input) > 4 {
				t.Errorf("bad test (%dth input wrong length)"+
					"%d: %v", j, i, test)
				continue
			}

			previoustx, ok := input[0].(string)
			if !ok {
				t.Errorf("bad test (%dth input hash not string)"+
					"%d: %v", j, i, test)
				continue
			}

			idxf, ok := input[1].(float64)
			if !ok {
				t.Errorf("bad test (%dth input idx not number)"+
					"%d: %v", j, i, test)
				continue
			}
			idx := testVecF64ToUint32(idxf)

			oscript, ok := input[2].(string)
			if !ok {
				t.Errorf("bad test (%dth input script not "+
					"string) %d: %v", j, i, test)
				continue
			}

			script, err := parseShortForm(oscript)
			if err != nil {
				t.Errorf("bad test (%dth input script doesn't "+
					"parse %v) %d: %v", j, err, i, test)
				continue
			}

			var inputValue float64
			if len(input) == 4 {
				inputValue, ok = input[3].(float64)
				if !ok {
					t.Errorf("bad test (%dth input value not int) "+
						"%d: %v", j, i, test)
					continue
				}
			}

			v := &bt.Output{
				Satoshis:      uint64(inputValue),
				LockingScript: script,
			}
			prevOuts[txIOKey{id: previoustx, idx: idx}] = v
		}

		for k, txin := range tx.Inputs {
			prevOut, ok := prevOuts[txIOKey{id: txin.PreviousTxIDStr(), idx: txin.PreviousTxOutIndex}]
			if !ok {
				t.Errorf("bad test (missing %dth input) %d:%v",
					k, i, test)
				continue testloop
			}

			if err = NewEngine().Execute(ExecutionParams{
				PreviousTxOut: prevOut,
				Tx:            tx,
				InputIdx:      k,
				Flags:         flags,
			}); err != nil {
				t.Errorf("test (%d:%v:%d) failed to execute: "+
					"%v", i, test, k, err)
				continue
			}
		}
	}
}
