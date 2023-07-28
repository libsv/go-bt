// Copyright (c) 2013-2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package interpreter

import (
	"errors"
	"testing"

	"github.com/libsv/go-bt/v2"
	"github.com/libsv/go-bt/v2/bscript"
	"github.com/libsv/go-bt/v2/bscript/interpreter/errs"
	"github.com/libsv/go-bt/v2/bscript/interpreter/scriptflag"
	"github.com/libsv/go-bt/v2/sighash"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	uscript, err := bscript.NewFromASM("OP_NOP")
	if err != nil {
		t.Errorf("failed to create unlocking script %e", err)
	}

	tx := &bt.Tx{
		Version: 1,
		Inputs: []*bt.Input{{
			PreviousTxOutIndex: 0,
			UnlockingScript:    uscript,
			SequenceNumber:     4294967295,
		}},
		Outputs: []*bt.Output{{
			Satoshis: 1000000000,
		}},
		LockTime: 0,
	}

	lscript, err := bscript.NewFromASM("OP_NOP")
	if err != nil {
		t.Errorf("failed to created locking script %e", err)
	}
	txOut := &bt.Output{
		LockingScript: lscript,
	}

	for _, test := range tests {
		vm := &thread{
			scriptParser: &DefaultOpcodeParser{},
			cfg:          &beforeGenesisConfig{},
		}
		err := vm.apply(&execOpts{
			previousTxOut: txOut,
			tx:            tx,
			inputIdx:      0,
		})
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

		if err == nil {
			t.Errorf("DisasmPC with invalid pc (%v) succeeds!",
				test)
		}
	}
}

// TestCheckErrorCondition tests to execute early test in CheckErrorCondition()
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

	lscript, err := bscript.NewFromASM("OP_NOP OP_NOP OP_NOP OP_NOP OP_NOP OP_NOP OP_NOP OP_NOP OP_NOP OP_NOP OP_TRUE")
	if err != nil {
		t.Errorf("failed to created locking script %e", err)
	}
	txOut := &bt.Output{
		LockingScript: lscript,
	}

	vm := &thread{
		scriptParser: &DefaultOpcodeParser{},
		cfg:          &beforeGenesisConfig{},
	}

	err = vm.apply(&execOpts{
		previousTxOut: txOut,
		inputIdx:      0,
		tx:            tx,
	})
	if err != nil {
		t.Errorf("failed to configure thread %v", err)
	}

	var done bool
	for i := 0; i < len(*lscript); i++ {
		done, err = vm.Step()
		if err != nil {
			t.Fatalf("failed to step %dth time: %v", i, err)
		}
		if done && i != len(*lscript)-1 {
			t.Fatalf("finished early on %dth time", i)
		}
	}
	err = vm.CheckErrorCondition(false)
	if err != nil {
		t.Errorf("unexpected error %v on final check", err)
	}
}

func TestValidateParams(t *testing.T) {
	tests := map[string]struct {
		params execOpts
		expErr error
	}{
		"valid tx/previous out checksig script": {
			params: execOpts{
				tx: func() *bt.Tx {
					tx := bt.NewTx()
					err := tx.From("ae81577c1a2434929a1224cf19aa63e167d88029965e2ca6de24defff014d031", 0, "76a91454807ccc44c0eec0b0e187b3ce0e137e9c6cd65d88ac", 0)
					assert.NoError(t, err)

					uscript, err := bscript.NewFromHexString("483045022100a4d9da733aeb29f9ba94dcaa578e71662cf29dd9742ce4b022c098211f4fdb06022041d24db4eda239fa15a12cf91229f6c352adab3c1c10091fc2aa517fe0f487c5412102454c535854802e5eaeaf5cbecd20e0aa508486063b71194dfde34744f19f1a5d")
					assert.NoError(t, err)

					tx.Inputs[0].UnlockingScript = uscript

					return tx
				}(),
				previousTxOut: func() *bt.Output {
					cbLockingScript, err := bscript.NewFromHexString("76a91454807ccc44c0eec0b0e187b3ce0e137e9c6cd65d88ac")
					assert.NoError(t, err)

					return &bt.Output{LockingScript: cbLockingScript, Satoshis: 0}
				}(),
			},
		},
		"valid tx/previous out non-checksig script": {
			params: execOpts{
				tx: func() *bt.Tx {
					tx := bt.NewTx()
					err := tx.From("ae81577c1a2434929a1224cf19aa63e167d88029965e2ca6de24defff014d031", 0, "52529387", 0)
					assert.NoError(t, err)

					txUnlockingScript, err := bscript.NewFromASM("OP_4")
					assert.NoError(t, err)

					tx.Inputs[0].UnlockingScript = txUnlockingScript

					return tx
				}(),
				previousTxOut: func() *bt.Output {
					cbLockingScript, err := bscript.NewFromASM("OP_2 OP_2 OP_ADD OP_EQUAL")
					assert.NoError(t, err)

					return &bt.Output{LockingScript: cbLockingScript, Satoshis: 0}
				}(),
			},
		},
		"valid locking/unlocking script non-checksig": {
			params: execOpts{
				lockingScript: func() *bscript.Script {
					script, err := bscript.NewFromHexString("52529387")
					assert.NoError(t, err)
					return script
				}(),
				unlockingScript: func() *bscript.Script {
					script, err := bscript.NewFromHexString("54")
					assert.NoError(t, err)
					return script
				}(),
			},
		},
		"valid locking/unlocking script with check-sig": {
			params: execOpts{
				lockingScript: func() *bscript.Script {
					script, err := bscript.NewFromHexString("76a91454807ccc44c0eec0b0e187b3ce0e137e9c6cd65d88ac")
					assert.NoError(t, err)
					return script
				}(),
				unlockingScript: func() *bscript.Script {
					script, err := bscript.NewFromHexString("483045022100a4d9da733aeb29f9ba94dcaa578e71662cf29dd9742ce4b022c098211f4fdb06022041d24db4eda239fa15a12cf91229f6c352adab3c1c10091fc2aa517fe0f487c5412102454c535854802e5eaeaf5cbecd20e0aa508486063b71194dfde34744f19f1a5d")
					assert.NoError(t, err)
					return script
				}(),
				tx: func() *bt.Tx {
					tx := bt.NewTx()
					err := tx.From("ae81577c1a2434929a1224cf19aa63e167d88029965e2ca6de24defff014d031", 0, "76a91454807ccc44c0eec0b0e187b3ce0e137e9c6cd65d88ac", 0)
					assert.NoError(t, err)

					uscript, err := bscript.NewFromHexString("483045022100a4d9da733aeb29f9ba94dcaa578e71662cf29dd9742ce4b022c098211f4fdb06022041d24db4eda239fa15a12cf91229f6c352adab3c1c10091fc2aa517fe0f487c5412102454c535854802e5eaeaf5cbecd20e0aa508486063b71194dfde34744f19f1a5d")
					assert.NoError(t, err)

					tx.Inputs[0].UnlockingScript = uscript

					return tx
				}(),
				previousTxOut: func() *bt.Output {
					script, err := bscript.NewFromHexString("76a91454807ccc44c0eec0b0e187b3ce0e137e9c6cd65d88ac")
					assert.NoError(t, err)

					return &bt.Output{LockingScript: script, Satoshis: 0}
				}(),
			},
		},
		"no locking script provided errors": {
			params: execOpts{
				unlockingScript: func() *bscript.Script {
					script, err := bscript.NewFromHexString("483045022100a4d9da733aeb29f9ba94dcaa578e71662cf29dd9742ce4b022c098211f4fdb06022041d24db4eda239fa15a12cf91229f6c352adab3c1c10091fc2aa517fe0f487c5412102454c535854802e5eaeaf5cbecd20e0aa508486063b71194dfde34744f19f1a5d")
					assert.NoError(t, err)
					return script
				}(),
				tx: func() *bt.Tx {
					tx := bt.NewTx()
					err := tx.From("ae81577c1a2434929a1224cf19aa63e167d88029965e2ca6de24defff014d031", 0, "76a91454807ccc44c0eec0b0e187b3ce0e137e9c6cd65d88ac", 0)
					assert.NoError(t, err)

					uscript, err := bscript.NewFromHexString("483045022100a4d9da733aeb29f9ba94dcaa578e71662cf29dd9742ce4b022c098211f4fdb06022041d24db4eda239fa15a12cf91229f6c352adab3c1c10091fc2aa517fe0f487c5412102454c535854802e5eaeaf5cbecd20e0aa508486063b71194dfde34744f19f1a5d")
					assert.NoError(t, err)

					tx.Inputs[0].UnlockingScript = uscript

					return tx
				}(),
			},
			expErr: errors.New("no locking script provided"),
		},
		"no unlocking script provided errors": {
			params: execOpts{
				lockingScript: func() *bscript.Script {
					script, err := bscript.NewFromHexString("76a91454807ccc44c0eec0b0e187b3ce0e137e9c6cd65d88ac")
					assert.NoError(t, err)
					return script
				}(),
				previousTxOut: func() *bt.Output {
					script, err := bscript.NewFromHexString("76a91454807ccc44c0eec0b0e187b3ce0e137e9c6cd65d88ac")
					assert.NoError(t, err)

					return &bt.Output{LockingScript: script, Satoshis: 0}
				}(),
			},
			expErr: errors.New("no unlocking script provided"),
		},
		"invalid locking/unlocking script with checksig": {
			params: execOpts{
				lockingScript: func() *bscript.Script {
					script, err := bscript.NewFromHexString("76a91454807ccc44c0eec0b0e187b3ce0e137e9c6cd65d88ac")
					assert.NoError(t, err)
					return script
				}(),
				unlockingScript: func() *bscript.Script {
					script, err := bscript.NewFromHexString("483045022100a4d9da733aeb29f9ba94dcaa578e71662cf29dd9742ce4b022c098211f4fdb06022041d24db4eda239fa15a12cf91229f6c352adab3c1c10091fc2aa517fe0f487c5412102454c535854802e5eaeaf5cbecd20e0aa508486063b71194dfde34744f19f1a5d")
					assert.NoError(t, err)
					return script
				}(),
			},
			expErr: errors.New("tx and previous output must be supplied for checksig"),
		},
		"provided locking script that differs from previous txout's errors": {
			params: execOpts{
				lockingScript: func() *bscript.Script {
					script, err := bscript.NewFromHexString("52529387")
					assert.NoError(t, err)
					return script
				}(),
				unlockingScript: func() *bscript.Script {
					script, err := bscript.NewFromHexString("483045022100a4d9da733aeb29f9ba94dcaa578e71662cf29dd9742ce4b022c098211f4fdb06022041d24db4eda239fa15a12cf91229f6c352adab3c1c10091fc2aa517fe0f487c5412102454c535854802e5eaeaf5cbecd20e0aa508486063b71194dfde34744f19f1a5d")
					assert.NoError(t, err)
					return script
				}(),
				tx: func() *bt.Tx {
					tx := bt.NewTx()
					err := tx.From("ae81577c1a2434929a1224cf19aa63e167d88029965e2ca6de24defff014d031", 0, "76a91454807ccc44c0eec0b0e187b3ce0e137e9c6cd65d88ac", 0)
					assert.NoError(t, err)

					uscript, err := bscript.NewFromHexString("483045022100a4d9da733aeb29f9ba94dcaa578e71662cf29dd9742ce4b022c098211f4fdb06022041d24db4eda239fa15a12cf91229f6c352adab3c1c10091fc2aa517fe0f487c5412102454c535854802e5eaeaf5cbecd20e0aa508486063b71194dfde34744f19f1a5d")
					assert.NoError(t, err)

					tx.Inputs[0].UnlockingScript = uscript

					return tx
				}(),
				previousTxOut: func() *bt.Output {
					script, err := bscript.NewFromHexString("76a91454807ccc44c0eec0b0e187b3ce0e137e9c6cd65d88ac")
					assert.NoError(t, err)

					return &bt.Output{LockingScript: script, Satoshis: 0}
				}(),
			},
			expErr: errors.New("locking script does not match the previous outputs locking script"),
		},
		"provided unlocking script that differs from tx input's errors": {
			params: execOpts{
				lockingScript: func() *bscript.Script {
					script, err := bscript.NewFromHexString("76a91454807ccc44c0eec0b0e187b3ce0e137e9c6cd65d88ac")
					assert.NoError(t, err)
					return script
				}(),
				unlockingScript: func() *bscript.Script {
					script, err := bscript.NewFromHexString("84")
					assert.NoError(t, err)
					return script
				}(),
				tx: func() *bt.Tx {
					tx := bt.NewTx()
					err := tx.From("ae81577c1a2434929a1224cf19aa63e167d88029965e2ca6de24defff014d031", 0, "76a91454807ccc44c0eec0b0e187b3ce0e137e9c6cd65d88ac", 0)
					assert.NoError(t, err)

					uscript, err := bscript.NewFromHexString("483045022100a4d9da733aeb29f9ba94dcaa578e71662cf29dd9742ce4b022c098211f4fdb06022041d24db4eda239fa15a12cf91229f6c352adab3c1c10091fc2aa517fe0f487c5412102454c535854802e5eaeaf5cbecd20e0aa508486063b71194dfde34744f19f1a5d")
					assert.NoError(t, err)

					tx.Inputs[0].UnlockingScript = uscript

					return tx
				}(),
				previousTxOut: func() *bt.Output {
					script, err := bscript.NewFromHexString("76a91454807ccc44c0eec0b0e187b3ce0e137e9c6cd65d88ac")
					assert.NoError(t, err)

					return &bt.Output{LockingScript: script, Satoshis: 0}
				}(),
			},
			expErr: errors.New("unlocking script does not match the unlocking script of the requested input"),
		},
		"invalid input index errors": {
			params: execOpts{
				tx: func() *bt.Tx {
					tx := bt.NewTx()
					err := tx.From("ae81577c1a2434929a1224cf19aa63e167d88029965e2ca6de24defff014d031", 0, "76a91454807ccc44c0eec0b0e187b3ce0e137e9c6cd65d88ac", 0)
					assert.NoError(t, err)

					uscript, err := bscript.NewFromHexString("483045022100a4d9da733aeb29f9ba94dcaa578e71662cf29dd9742ce4b022c098211f4fdb06022041d24db4eda239fa15a12cf91229f6c352adab3c1c10091fc2aa517fe0f487c5412102454c535854802e5eaeaf5cbecd20e0aa508486063b71194dfde34744f19f1a5d")
					assert.NoError(t, err)

					tx.Inputs[0].UnlockingScript = uscript

					return tx
				}(),
				previousTxOut: func() *bt.Output {
					cbLockingScript, err := bscript.NewFromHexString("76a91454807ccc44c0eec0b0e187b3ce0e137e9c6cd65d88ac")
					assert.NoError(t, err)

					return &bt.Output{LockingScript: cbLockingScript, Satoshis: 0}
				}(),
				inputIdx: 5,
			},
			expErr: errors.New("transaction input index 5 is negative or >= 1"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := createThread(&test.params)

			if test.expErr != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, test.expErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestInvalidFlagCombinations ensures the script engine returns the expected
// error when disallowed flag combinations are specified.
func TestInvalidFlagCombinations(t *testing.T) {
	t.Parallel()

	tests := []scriptflag.Flag{
		scriptflag.VerifyCleanStack,
	}

	uscript, err := bscript.NewFromASM("OP_NOP")
	if err != nil {
		t.Errorf("failed to create unlocking script %e", err)
	}

	tx := &bt.Tx{
		Version: 1,
		Inputs: []*bt.Input{{
			PreviousTxOutIndex: 0,
			UnlockingScript:    uscript,
			SequenceNumber:     4294967295,
		}},
		Outputs: []*bt.Output{{
			Satoshis: 1000000000,
		}},
		LockTime: 0,
	}

	lscript, err := bscript.NewFromASM("OP_NOP")
	if err != nil {
		t.Errorf("failed to created locking script %e", err)
	}
	txOut := &bt.Output{
		LockingScript: lscript,
	}

	for i, test := range tests {
		vm := &thread{
			scriptParser: &DefaultOpcodeParser{},
			cfg:          &beforeGenesisConfig{},
		}
		err := vm.apply(&execOpts{
			tx:            tx,
			inputIdx:      0,
			previousTxOut: txOut,
			flags:         test,
		})
		if !errs.IsErrorCode(err, errs.ErrInvalidFlags) {
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

	vm := thread{flags: scriptflag.VerifyStrictEncoding}
	for _, test := range tests {
		err := vm.checkPubKeyEncoding(test.key)
		if err != nil && test.isValid {
			t.Errorf("checkSignatureEncoding test '%s' failed "+
				"when it should have succeeded: %v", test.name,
				err)
		} else if err == nil && !test.isValid {
			t.Errorf("checkSignatureEncoding test '%s' succeeded "+
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

	vm := thread{flags: scriptflag.VerifyStrictEncoding}
	for _, test := range tests {
		err := vm.checkSignatureEncoding(test.sig)
		if err != nil && test.isValid {
			t.Errorf("checkSignatureEncoding test '%s' failed "+
				"when it should have succeeded: %v", test.name,
				err)
		} else if err == nil && !test.isValid {
			t.Errorf("checkSignatureEncoding test '%s' succeeded "+
				"when it should have failed", test.name)
		}
	}
}

func TestCheckHashTypeEncoding(t *testing.T) {
	var SigHashBug sighash.Flag = 0x20
	encodingTests := []struct {
		SigHash     sighash.Flag
		EngineFlags scriptflag.Flag
		ShouldFail  bool
	}{
		{
			sighash.All,
			scriptflag.VerifyStrictEncoding,
			false,
		},
		{
			sighash.None,
			scriptflag.VerifyStrictEncoding,
			false,
		},
		{
			sighash.Single,
			scriptflag.VerifyStrictEncoding,
			false,
		},
		{
			sighash.All | sighash.AnyOneCanPay,
			scriptflag.VerifyStrictEncoding,
			false,
		},
		{
			sighash.None | sighash.AnyOneCanPay,
			scriptflag.VerifyStrictEncoding,
			false,
		},
		{
			sighash.Single | sighash.AnyOneCanPay,
			scriptflag.VerifyStrictEncoding,
			false,
		},
		{
			sighash.All | sighash.ForkID,
			scriptflag.VerifyStrictEncoding,
			true,
		},
		{
			sighash.None | sighash.ForkID,
			scriptflag.VerifyStrictEncoding,
			true,
		},
		{
			sighash.Single | sighash.ForkID,
			scriptflag.VerifyStrictEncoding,
			true,
		},
		{
			sighash.All | sighash.AnyOneCanPay | sighash.ForkID,
			scriptflag.VerifyStrictEncoding,
			true,
		},
		{
			sighash.None | sighash.AnyOneCanPay | sighash.ForkID,
			scriptflag.VerifyStrictEncoding,
			true,
		},
		{
			sighash.Single | sighash.AnyOneCanPay | sighash.ForkID,
			scriptflag.VerifyStrictEncoding,
			true,
		},

		{
			sighash.All | sighash.ForkID,
			scriptflag.VerifyStrictEncoding | scriptflag.VerifyBip143SigHash,
			false,
		},
		{
			sighash.None | sighash.ForkID,
			scriptflag.VerifyStrictEncoding | scriptflag.VerifyBip143SigHash,
			false,
		},
		{
			sighash.Single | sighash.ForkID,
			scriptflag.VerifyStrictEncoding | scriptflag.VerifyBip143SigHash,
			false,
		},
		{
			sighash.All | sighash.AnyOneCanPay | sighash.ForkID,
			scriptflag.VerifyStrictEncoding | scriptflag.VerifyBip143SigHash,
			false,
		},
		{
			sighash.None | sighash.AnyOneCanPay | sighash.ForkID,
			scriptflag.VerifyStrictEncoding | scriptflag.VerifyBip143SigHash,
			false,
		},
		{
			sighash.Single | sighash.AnyOneCanPay | sighash.ForkID,
			scriptflag.VerifyStrictEncoding | scriptflag.VerifyBip143SigHash,
			false,
		},

		{
			sighash.All,
			scriptflag.VerifyStrictEncoding | scriptflag.VerifyBip143SigHash,
			true,
		},
		{
			sighash.None,
			scriptflag.VerifyStrictEncoding | scriptflag.VerifyBip143SigHash,
			true,
		},
		{
			sighash.Single,
			scriptflag.VerifyStrictEncoding | scriptflag.VerifyBip143SigHash,
			true,
		},
		{
			sighash.All | sighash.AnyOneCanPay,
			scriptflag.VerifyStrictEncoding | scriptflag.VerifyBip143SigHash,
			true,
		},
		{
			sighash.None | sighash.AnyOneCanPay,
			scriptflag.VerifyStrictEncoding | scriptflag.VerifyBip143SigHash,
			true,
		},
		{
			sighash.Single | sighash.AnyOneCanPay,
			scriptflag.VerifyStrictEncoding | scriptflag.VerifyBip143SigHash,
			true,
		},
		{
			sighash.Single | sighash.AnyOneCanPay | sighash.ForkID | SigHashBug,
			scriptflag.VerifyStrictEncoding | scriptflag.VerifyBip143SigHash,
			true,
		},
	}

	for i, test := range encodingTests {
		e := thread{flags: test.EngineFlags}
		err := e.checkHashTypeEncoding(test.SigHash)
		if test.ShouldFail && err == nil {
			t.Errorf("Expected test %d to fail", i)
		} else if !test.ShouldFail && err != nil {
			t.Errorf("Expected test %d not to fail", i)
		}
	}
}

func TestEngine_WithState(t *testing.T) {
	tests := map[string]struct {
		lscript string
		uscript string
		state   *State
	}{
		"start midway": {
			lscript: "5253958852529387",
			uscript: "5456",
			state: &State{
				ScriptIdx: 1,
				OpcodeIdx: 1,
				DataStack: func() [][]byte {
					return [][]byte{{4}, {6}, {2}}
				}(),
				AltStack:             [][]byte{},
				CondStack:            []int{},
				ElseStack:            [][]byte{},
				Flags:                scriptflag.UTXOAfterGenesis | scriptflag.EnableSighashForkID,
				LastCodeSeparatorIdx: 0,
				NumOps:               3,
				SavedFirstStack:      [][]byte{},
				Scripts: func() []ParsedScript {
					lscript, err := bscript.NewFromHexString("5253958852529387")
					assert.NoError(t, err)
					uscript, err := bscript.NewFromHexString("5456")
					assert.NoError(t, err)

					var parser DefaultOpcodeParser
					parsedLScript, err := parser.Parse(lscript)
					assert.NoError(t, err)

					parsedUScript, err := parser.Parse(uscript)
					assert.NoError(t, err)

					return []ParsedScript{parsedUScript, parsedLScript}
				}(),
				Genesis: struct {
					AfterGenesis bool
					EarlyReturn  bool
				}{
					AfterGenesis: true,
				},
			},
		},
		"start at operation": {
			lscript: "5253958852529387",
			uscript: "5456",
			state: &State{
				ScriptIdx: 1,
				OpcodeIdx: 6,
				DataStack: func() [][]byte {
					return [][]byte{{4}, {2}, {2}}
				}(),
				AltStack:             [][]byte{},
				CondStack:            []int{},
				ElseStack:            [][]byte{},
				Flags:                scriptflag.UTXOAfterGenesis | scriptflag.EnableSighashForkID,
				LastCodeSeparatorIdx: 0,
				NumOps:               8,
				SavedFirstStack:      [][]byte{},
				Scripts: func() []ParsedScript {
					lscript, err := bscript.NewFromHexString("5253958852529387")
					assert.NoError(t, err)
					uscript, err := bscript.NewFromHexString("5456")
					assert.NoError(t, err)

					var parser DefaultOpcodeParser
					parsedLScript, err := parser.Parse(lscript)
					assert.NoError(t, err)

					parsedUScript, err := parser.Parse(uscript)
					assert.NoError(t, err)

					return []ParsedScript{parsedUScript, parsedLScript}
				}(),
				Genesis: struct {
					AfterGenesis bool
					EarlyReturn  bool
				}{
					AfterGenesis: true,
				},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			lscript, err := bscript.NewFromHexString(test.lscript)
			assert.NoError(t, err)
			uscript, err := bscript.NewFromHexString(test.uscript)
			assert.NoError(t, err)

			assert.NoError(t, NewEngine().Execute(
				WithScripts(lscript, uscript),
				WithForkID(),
				WithAfterGenesis(),
				WithState(test.state),
			))
		})
	}
}

const (
	txHex1 = `0100000001abdbd5873fbda1b08c19d899993301fd44c0aa735064ebb2248260b7adadf795000000006b483045022100e7813394c7a55941c1acf3c7032046c2aa5bf3a506b4ee09e4cb5761c1850f960220154769af29eef81d56d69eba1d7a5ab37eed15beb9eadcd2cb608ff2e09b3147c321035941a219bcd9688318028afeef55183634f010a933de9d8469ff6e702d96c238ffffffff010271000000000000220687623971234575ab76a914fbcf31b659334eeb086693fc3b4005ce29e1c21788ac00000000`

	txHex2 = `01000000014cc6b457cc6a235b966cec69bc4e4ea1813b71bddb2adf800848e4430e622b3d000000006a47304402201c1b7c535ff8bbee0960e0dad34e0a07857eaae5abc5a556427f4cc95e36cea50220676e3fd4eb69e98d8f9659c3bfceb0cdb34a6926ff644a6d79666e2c8266cc78c321035941a219bcd9688318028afeef55183634f010a933de9d8469ff6e702d96c238ffffffff011671000000000000220687623971234575ab76a914fbcf31b659334eeb086693fc3b4005ce29e1c21788ac00000000`
)

func TestExecute(t *testing.T) {
	t.Run("OP_CODESEPARATOR parsing", func(t *testing.T) {

		tx, err := bt.NewTxFromString(txHex1)
		require.NoError(t, err)

		prevTx, err := bt.NewTxFromString(txHex2)
		require.NoError(t, err)

		inputIdx := 0
		input := tx.InputIdx(inputIdx)
		prevOutput := prevTx.OutputIdx(int(input.PreviousTxOutIndex))

		err = NewEngine().Execute(
			WithTx(tx, inputIdx, prevOutput),
			WithForkID(),
			WithAfterGenesis(),
		)
		require.NoError(t, err)
	})
}
