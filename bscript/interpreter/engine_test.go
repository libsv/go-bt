// Copyright (c) 2013-2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package interpreter

import (
	"testing"

	"github.com/libsv/go-bt/v2"
	"github.com/libsv/go-bt/v2/bscript"
	"github.com/libsv/go-bt/v2/sighash"
)

// TestBadPC sets the pc to a deliberately bad result then confirms that Step()
// and Disasm fail correctly.
func TestBadPC(t *testing.T) {
	t.Parallel()

	tests := []struct {
		script, off int
	}{
		{script: 2, off: 0},
		{script: 0, off: 2},
	}

	uls, err := bscript.NewFromASM("OP_NOP")
	if err != nil {
		t.Errorf("failed to create unlocking script %e", err)
	}

	tx := &bt.Tx{
		Version: 1,
		Inputs: []*bt.Input{{
			PreviousTxOutIndex: 0,
			UnlockingScript:    uls,
			SequenceNumber:     4294967295,
		}},
		Outputs: []*bt.Output{{
			Satoshis: 1000000000,
		}},
		LockTime: 0,
	}

	ls, err := bscript.NewFromASM("OP_NOP")
	if err != nil {
		t.Errorf("failed to created locking script %e", err)
	}
	txOut := &bt.Output{
		LockingScript: ls,
	}

	for _, test := range tests {
		vm, err := NewEngine(EngineOpts{
			PreviousTxOut: txOut,
			Tx:            tx,
			InputIdx:      0,
		})
		//vm, err := NewEngine(pkScript, tx, 0, 0, nil, nil, -1)
		if err != nil {
			t.Errorf("Failed to create script: %v", err)
		}

		// set to after all scripts
		vm.scriptIdx = test.script
		vm.scriptOff = test.off

		_, err = vm.Step()
		if err == nil {
			t.Errorf("Step with invalid pc (%v) succeeds!", test)
			continue
		}

		_, err = vm.DisasmPC()
		if err == nil {
			t.Errorf("DisasmPC with invalid pc (%v) succeeds!",
				test)
		}
	}
}

// TestCheckErrorCondition tests the execute early test in CheckErrorCondition()
// since most code paths are tested elsewhere.
func TestCheckErrorCondition(t *testing.T) {
	t.Parallel()

	tx := &bt.Tx{
		Version: 1,
		Inputs: []*bt.Input{{
			PreviousTxOutIndex: 0,
			UnlockingScript:    &bscript.Script{},
			SequenceNumber:     4294967295,
		}},
		Outputs: []*bt.Output{{
			Satoshis: 1000000000,
		}},
		LockTime: 0,
	}

	ls, err := bscript.NewFromASM("OP_NOP OP_NOP OP_NOP OP_NOP OP_NOP OP_NOP OP_NOP OP_NOP OP_NOP OP_NOP OP_TRUE")
	if err != nil {
		t.Errorf("failed to created locking script %e", err)
	}
	txOut := &bt.Output{
		LockingScript: ls,
	}

	vm, err := NewEngine(EngineOpts{
		Tx:            tx,
		PreviousTxOut: txOut,
		InputIdx:      0,
	})
	if err != nil {
		t.Errorf("failed to create script: %v", err)
	}

	for i := 0; i < len(*ls)-1; i++ {
		done, err := vm.Step()
		if err != nil {
			t.Fatalf("failed to step %dth time: %v", i, err)
		}
		if done {
			t.Fatalf("finshed early on %dth time", i)
		}

		err = vm.CheckErrorCondition(false)
		if !IsErrorCode(err, ErrScriptUnfinished) {
			t.Fatalf("got unexepected error %v on %dth iteration",
				err, i)
		}
	}
	done, err := vm.Step()
	if err != nil {
		t.Fatalf("final step failed %v", err)
	}
	if !done {
		t.Fatalf("final step isn't done!")
	}

	err = vm.CheckErrorCondition(false)
	if err != nil {
		t.Errorf("unexpected error %v on final check", err)
	}
}

// TestInvalidFlagCombinations ensures the script engine returns the expected
// error when disallowed flag combinations are specified.
func TestInvalidFlagCombinations(t *testing.T) {
	t.Parallel()

	tests := []ScriptFlags{
		ScriptVerifyCleanStack,
	}

	uls, err := bscript.NewFromASM("OP_NOP")
	if err != nil {
		t.Errorf("failed to create unlocking script %e", err)
	}

	tx := &bt.Tx{
		Version: 1,
		Inputs: []*bt.Input{{
			PreviousTxOutIndex: 0,
			UnlockingScript:    uls,
			SequenceNumber:     4294967295,
		}},
		Outputs: []*bt.Output{{
			Satoshis: 1000000000,
		}},
		LockTime: 0,
	}

	ls, err := bscript.NewFromASM("OP_NOP")
	if err != nil {
		t.Errorf("failed to created locking script %e", err)
	}
	txOut := &bt.Output{
		LockingScript: ls,
	}

	for i, test := range tests {
		_, err := NewEngine(EngineOpts{
			Tx:            tx,
			Flags:         test,
			InputIdx:      0,
			PreviousTxOut: txOut,
		})
		if !IsErrorCode(err, ErrInvalidFlags) {
			t.Fatalf("TestInvalidFlagCombinations #%d unexpected "+
				"error: %v", i, err)
		}
	}
}

// TestCheckPubKeyEncoding ensures the internal checkPubKeyEncoding function
// works as expected.
func TestCheckPubKeyEncoding(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		key     []byte
		isValid bool
	}{
		{
			name: "uncompressed ok",
			key: hexToBytes("0411db93e1dcdb8a016b49840f8c53bc1eb68" +
				"a382e97b1482ecad7b148a6909a5cb2e0eaddfb84ccf" +
				"9744464f82e160bfa9b8b64f9d4c03f999b8643f656b" +
				"412a3"),
			isValid: true,
		},
		{
			name: "compressed ok",
			key: hexToBytes("02ce0b14fb842b1ba549fdd675c98075f12e9" +
				"c510f8ef52bd021a9a1f4809d3b4d"),
			isValid: true,
		},
		{
			name: "compressed ok",
			key: hexToBytes("032689c7c2dab13309fb143e0e8fe39634252" +
				"1887e976690b6b47f5b2a4b7d448e"),
			isValid: true,
		},
		{
			name: "hybrid",
			key: hexToBytes("0679be667ef9dcbbac55a06295ce870b07029" +
				"bfcdb2dce28d959f2815b16f81798483ada7726a3c46" +
				"55da4fbfc0e1108a8fd17b448a68554199c47d08ffb1" +
				"0d4b8"),
			isValid: false,
		},
		{
			name:    "empty",
			key:     nil,
			isValid: false,
		},
	}

	vm := Engine{flags: ScriptVerifyStrictEncoding}
	for _, test := range tests {
		err := vm.checkPubKeyEncoding(test.key)
		if err != nil && test.isValid {
			t.Errorf("checkSignatureEncoding test '%s' failed "+
				"when it should have succeeded: %v", test.name,
				err)
		} else if err == nil && !test.isValid {
			t.Errorf("checkSignatureEncooding test '%s' succeeded "+
				"when it should have failed", test.name)
		}
	}

}

// TestCheckSignatureEncoding ensures the internal checkSignatureEncoding
// function works as expected.
func TestCheckSignatureEncoding(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		sig     []byte
		isValid bool
	}{
		{
			name: "valid signature",
			sig: hexToBytes("304402204e45e16932b8af514961a1d3a1a25" +
				"fdf3f4f7732e9d624c6c61548ab5fb8cd41022018152" +
				"2ec8eca07de4860a4acdd12909d831cc56cbbac46220" +
				"82221a8768d1d09"),
			isValid: true,
		},
		{
			name:    "empty.",
			sig:     nil,
			isValid: false,
		},
		{
			name: "bad magic",
			sig: hexToBytes("314402204e45e16932b8af514961a1d3a1a25" +
				"fdf3f4f7732e9d624c6c61548ab5fb8cd41022018152" +
				"2ec8eca07de4860a4acdd12909d831cc56cbbac46220" +
				"82221a8768d1d09"),
			isValid: false,
		},
		{
			name: "bad 1st int marker magic",
			sig: hexToBytes("304403204e45e16932b8af514961a1d3a1a25" +
				"fdf3f4f7732e9d624c6c61548ab5fb8cd41022018152" +
				"2ec8eca07de4860a4acdd12909d831cc56cbbac46220" +
				"82221a8768d1d09"),
			isValid: false,
		},
		{
			name: "bad 2nd int marker",
			sig: hexToBytes("304402204e45e16932b8af514961a1d3a1a25" +
				"fdf3f4f7732e9d624c6c61548ab5fb8cd41032018152" +
				"2ec8eca07de4860a4acdd12909d831cc56cbbac46220" +
				"82221a8768d1d09"),
			isValid: false,
		},
		{
			name: "short len",
			sig: hexToBytes("304302204e45e16932b8af514961a1d3a1a25" +
				"fdf3f4f7732e9d624c6c61548ab5fb8cd41022018152" +
				"2ec8eca07de4860a4acdd12909d831cc56cbbac46220" +
				"82221a8768d1d09"),
			isValid: false,
		},
		{
			name: "long len",
			sig: hexToBytes("304502204e45e16932b8af514961a1d3a1a25" +
				"fdf3f4f7732e9d624c6c61548ab5fb8cd41022018152" +
				"2ec8eca07de4860a4acdd12909d831cc56cbbac46220" +
				"82221a8768d1d09"),
			isValid: false,
		},
		{
			name: "long X",
			sig: hexToBytes("304402424e45e16932b8af514961a1d3a1a25" +
				"fdf3f4f7732e9d624c6c61548ab5fb8cd41022018152" +
				"2ec8eca07de4860a4acdd12909d831cc56cbbac46220" +
				"82221a8768d1d09"),
			isValid: false,
		},
		{
			name: "long Y",
			sig: hexToBytes("304402204e45e16932b8af514961a1d3a1a25" +
				"fdf3f4f7732e9d624c6c61548ab5fb8cd41022118152" +
				"2ec8eca07de4860a4acdd12909d831cc56cbbac46220" +
				"82221a8768d1d09"),
			isValid: false,
		},
		{
			name: "short Y",
			sig: hexToBytes("304402204e45e16932b8af514961a1d3a1a25" +
				"fdf3f4f7732e9d624c6c61548ab5fb8cd41021918152" +
				"2ec8eca07de4860a4acdd12909d831cc56cbbac46220" +
				"82221a8768d1d09"),
			isValid: false,
		},
		{
			name: "trailing crap",
			sig: hexToBytes("304402204e45e16932b8af514961a1d3a1a25" +
				"fdf3f4f7732e9d624c6c61548ab5fb8cd41022018152" +
				"2ec8eca07de4860a4acdd12909d831cc56cbbac46220" +
				"82221a8768d1d0901"),
			isValid: false,
		},
		{
			name: "X == N ",
			sig: hexToBytes("30440220fffffffffffffffffffffffffffff" +
				"ffebaaedce6af48a03bbfd25e8cd0364141022018152" +
				"2ec8eca07de4860a4acdd12909d831cc56cbbac46220" +
				"82221a8768d1d09"),
			isValid: false,
		},
		{
			name: "X == N ",
			sig: hexToBytes("30440220fffffffffffffffffffffffffffff" +
				"ffebaaedce6af48a03bbfd25e8cd0364142022018152" +
				"2ec8eca07de4860a4acdd12909d831cc56cbbac46220" +
				"82221a8768d1d09"),
			isValid: false,
		},
		{
			name: "Y == N",
			sig: hexToBytes("304402204e45e16932b8af514961a1d3a1a25" +
				"fdf3f4f7732e9d624c6c61548ab5fb8cd410220fffff" +
				"ffffffffffffffffffffffffffebaaedce6af48a03bb" +
				"fd25e8cd0364141"),
			isValid: false,
		},
		{
			name: "Y > N",
			sig: hexToBytes("304402204e45e16932b8af514961a1d3a1a25" +
				"fdf3f4f7732e9d624c6c61548ab5fb8cd410220fffff" +
				"ffffffffffffffffffffffffffebaaedce6af48a03bb" +
				"fd25e8cd0364142"),
			isValid: false,
		},
		{
			name: "0 len X",
			sig: hexToBytes("302402000220181522ec8eca07de4860a4acd" +
				"d12909d831cc56cbbac4622082221a8768d1d09"),
			isValid: false,
		},
		{
			name: "0 len Y",
			sig: hexToBytes("302402204e45e16932b8af514961a1d3a1a25" +
				"fdf3f4f7732e9d624c6c61548ab5fb8cd410200"),
			isValid: false,
		},
		{
			name: "extra R padding",
			sig: hexToBytes("30450221004e45e16932b8af514961a1d3a1a" +
				"25fdf3f4f7732e9d624c6c61548ab5fb8cd410220181" +
				"522ec8eca07de4860a4acdd12909d831cc56cbbac462" +
				"2082221a8768d1d09"),
			isValid: false,
		},
		{
			name: "extra S padding",
			sig: hexToBytes("304502204e45e16932b8af514961a1d3a1a25" +
				"fdf3f4f7732e9d624c6c61548ab5fb8cd41022100181" +
				"522ec8eca07de4860a4acdd12909d831cc56cbbac462" +
				"2082221a8768d1d09"),
			isValid: false,
		},
	}

	vm := Engine{flags: ScriptVerifyStrictEncoding}
	for _, test := range tests {
		err := vm.checkSignatureEncoding(test.sig)
		if err != nil && test.isValid {
			t.Errorf("checkSignatureEncoding test '%s' failed "+
				"when it should have succeeded: %v", test.name,
				err)
		} else if err == nil && !test.isValid {
			t.Errorf("checkSignatureEncooding test '%s' succeeded "+
				"when it should have failed", test.name)
		}
	}
}

func TestCheckHashTypeEncoding(t *testing.T) {
	var SigHashBug sighash.Flag = 0x20
	encodingTests := []struct {
		SigHash     sighash.Flag
		EngineFlags ScriptFlags
		ShouldFail  bool
	}{
		{
			sighash.All,
			ScriptVerifyStrictEncoding,
			false,
		},
		{
			sighash.None,
			ScriptVerifyStrictEncoding,
			false,
		},
		{
			sighash.Single,
			ScriptVerifyStrictEncoding,
			false,
		},
		{
			sighash.All | sighash.AnyOneCanPay,
			ScriptVerifyStrictEncoding,
			false,
		},
		{
			sighash.None | sighash.AnyOneCanPay,
			ScriptVerifyStrictEncoding,
			false,
		},
		{
			sighash.Single | sighash.AnyOneCanPay,
			ScriptVerifyStrictEncoding,
			false,
		},
		{
			sighash.All | sighash.ForkID,
			ScriptVerifyStrictEncoding,
			true,
		},
		{
			sighash.None | sighash.ForkID,
			ScriptVerifyStrictEncoding,
			true,
		},
		{
			sighash.Single | sighash.ForkID,
			ScriptVerifyStrictEncoding,
			true,
		},
		{
			sighash.All | sighash.AnyOneCanPay | sighash.ForkID,
			ScriptVerifyStrictEncoding,
			true,
		},
		{
			sighash.None | sighash.AnyOneCanPay | sighash.ForkID,
			ScriptVerifyStrictEncoding,
			true,
		},
		{
			sighash.Single | sighash.AnyOneCanPay | sighash.ForkID,
			ScriptVerifyStrictEncoding,
			true,
		},

		{
			sighash.All | sighash.ForkID,
			ScriptVerifyStrictEncoding | ScriptVerifyBip143SigHash,
			false,
		},
		{
			sighash.None | sighash.ForkID,
			ScriptVerifyStrictEncoding | ScriptVerifyBip143SigHash,
			false,
		},
		{
			sighash.Single | sighash.ForkID,
			ScriptVerifyStrictEncoding | ScriptVerifyBip143SigHash,
			false,
		},
		{
			sighash.All | sighash.AnyOneCanPay | sighash.ForkID,
			ScriptVerifyStrictEncoding | ScriptVerifyBip143SigHash,
			false,
		},
		{
			sighash.None | sighash.AnyOneCanPay | sighash.ForkID,
			ScriptVerifyStrictEncoding | ScriptVerifyBip143SigHash,
			false,
		},
		{
			sighash.Single | sighash.AnyOneCanPay | sighash.ForkID,
			ScriptVerifyStrictEncoding | ScriptVerifyBip143SigHash,
			false,
		},

		{
			sighash.All,
			ScriptVerifyStrictEncoding | ScriptVerifyBip143SigHash,
			true,
		},
		{
			sighash.None,
			ScriptVerifyStrictEncoding | ScriptVerifyBip143SigHash,
			true,
		},
		{
			sighash.Single,
			ScriptVerifyStrictEncoding | ScriptVerifyBip143SigHash,
			true,
		},
		{
			sighash.All | sighash.AnyOneCanPay,
			ScriptVerifyStrictEncoding | ScriptVerifyBip143SigHash,
			true,
		},
		{
			sighash.None | sighash.AnyOneCanPay,
			ScriptVerifyStrictEncoding | ScriptVerifyBip143SigHash,
			true,
		},
		{
			sighash.Single | sighash.AnyOneCanPay,
			ScriptVerifyStrictEncoding | ScriptVerifyBip143SigHash,
			true,
		},
		{
			sighash.Single | sighash.AnyOneCanPay | sighash.ForkID | SigHashBug,
			ScriptVerifyStrictEncoding | ScriptVerifyBip143SigHash,
			true,
		},
	}

	for i, test := range encodingTests {
		e := Engine{flags: test.EngineFlags}
		err := e.checkHashTypeEncoding(test.SigHash)
		if test.ShouldFail && err == nil {
			t.Errorf("Expected test %d to fail", i)
		} else if !test.ShouldFail && err != nil {
			t.Errorf("Expected test %d not to fail", i)
		}
	}
}
