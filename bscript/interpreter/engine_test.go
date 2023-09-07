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

	prevTxHex1 = `01000000014cc6b457cc6a235b966cec69bc4e4ea1813b71bddb2adf800848e4430e622b3d000000006a47304402201c1b7c535ff8bbee0960e0dad34e0a07857eaae5abc5a556427f4cc95e36cea50220676e3fd4eb69e98d8f9659c3bfceb0cdb34a6926ff644a6d79666e2c8266cc78c321035941a219bcd9688318028afeef55183634f010a933de9d8469ff6e702d96c238ffffffff011671000000000000220687623971234575ab76a914fbcf31b659334eeb086693fc3b4005ce29e1c21788ac00000000`

	txHex2 = `01000000034b8fcb7a23da7016355f50c5d1c8c7136f014ee9ace434350cdbd8c301881e4400000000fd5b044db7027b0a20202273657373696f6e4964223a202233636630346432322d636137342d343730392d383637322d323233313764316430646261222c0a2020226275796572223a207b0a20202020227075626c69634b6579223a2022303261376238633535363632656538646331623533346363333861626332383738643162383261643865396238363063656238353461386465383339336261663933222c0a2020202022637573746f6469616e223a2022544f4b454e4f56415445220a20207d2c0a20202273656c6c6572223a207b0a20202020227075626c69634b6579223a2022303231666438326631366431623636393639636237616131666435616362363964333963326635623933336266353464613836316531313637623735303062643534222c0a2020202022637573746f6469616e223a2022544f4b454e4f56415445220a20207d2c0a202022657865637574696f6e4964223a2022313438393030303030303030303030313538222c0a20202273796d626f6c223a20224b5341412d53504f54222c0a2020227175616e74697479223a2022312e30222c0a2020227072696365223a202238302e30222c0a2020227472616e73616374696f6e54696d65223a2022323032332d30352d32335431393a30303a30342e3833385a222c0a2020226d73674f726967696e223a2022474d4558222c0a2020226d7367554944223a202234616430383338372d323037332d346136342d383735342d613132383464623237393330222c0a2020226d736754797065223a20225452414445222c0a2020226d736754696d657374616d70223a20313638343836383430363433302c0a20202263617074757265644174223a20313638343836383430373636362c0a2020226576656e744964223a202263346263613235612d396263632d343364352d396539352d323538666666363534623137220a7d2412242242150a2742150a1912100912030f041918130410241105150d120d1024000000004ca001000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000004b8fcb7a23da7016355f50c5d1c8c7136f014ee9ace434350cdbd8c301881e440000000003adadac0100000000000000000000006c960ca3c7d91e65a2272a7b4059200a09d0b3cc4eef615ed9948f0e6b59aaa500000000c3000000483045022100fa0cea230d94a5f61a2e8b0b83d25bd28d33dca80c4b0c592ed2bfd135cf571302203e124f85b393cd7a09e2ba92189e3f0d5f3eee177f6a5b1c8b2cce841f06e1dec3473044022068209ef4ab5b548585217f12b5796042e19d8c9cabd9221d481a7c68be28bacf022074c0f2a2cf8a75e8fc4673e99fc7fe0abba05300aaa818f74328dc82e5698003c3483045022100d8f1ac46c4b0fb3bd6de7eb71ff6aed0fa42d9a9ac9372618dc2127301a6a1ec0220318f49dd87b54bdfc5b632ab9279a809a50c8087f54536e8e1f5adb8849de31dc3000000004e76c1ab14c5e89c3539587e126249658a5f5972b18790621bf33bfe910c51eb0a0000001514378e9efb1f8330a321df32bc0dbfdee93cf5bcc8ffffffff4e76c1ab14c5e89c3539587e126249658a5f5972b18790621bf33bfe910c51eb0b0000001514d1e0da74dfcc481916df4a0d0e92008a0924326bffffffff010000000000000000fd3c02006a4d3702546f6b656e6f76617465206861766520637265617465642061206e6577205452414445206576656e74206265747765656e3a0a42757965723a203032613762386335353636326565386463316235333463633338616263323837386431623832616438653962383630636562383534613864653833393362616639330a427579657220637573746f6469616e3a20544f4b454e4f564154450a616e640a53656c6c65723a203032316664383266313664316236363936396362376161316664356163623639643339633266356239333362663534646138363165313136376237353030626435340a53656c6c657220637573746f6469616e3a20544f4b454e4f564154450a0a666f7220312e3020756e697473204b5341412d53504f542061742061207072696365206f66202438302e302055534420656163682e0a0a5468652074726164652076656e756520697320474d45580a0a546869732074726164652077617320636f6d706c6574656420617420323032332d30352d32335431393a30303a30342e3833385a0a0a466f72206d6f726520696e666f726d6174696f6e207365653a2068747470733a2f2f726563656970747669657765722e746f6b656e6f766174652e636f6d2f74726164652f3134383930303030303030303030303135380a4b5341412d53706f742074726164652056657273696f6e3a20302e320a54686973207472616465206973207265636f72646564206f6e2074686520546f6b656e6f76617465207472616465206c65646765722e00000000`

	prevTxHex2 = `01000000034e76c1ab14c5e89c3539587e126249658a5f5972b18790621bf33bfe910c51eb000000001514c20e5ee14158447c755d84770ed78c15a28ea835ffffffff4e76c1ab14c5e89c3539587e126249658a5f5972b18790621bf33bfe910c51eb010000001514a6ee8da20f9e18f09f41f090fd4c39a425fe4fe9ffffffff4e76c1ab14c5e89c3539587e126249658a5f5972b18790621bf33bfe910c51eb020000001514299f25978f6e8c46e7bd251cad128ffd48e90f67ffffffff010100000000000000fdfc16280a0a5543522d3136372076302e3161736d0a7777772e746f6b656e6f766174652e636f6d0a0a0a0a7576a91439900cd4a915bccd7efb137348fd3b0e5c1e3c978763ac6a6821026d25662a8c4d6d7822753f7ef95f964cd26f52d3d4be923ee858cb1e9830acedad547a547a517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f75816b816b816b816b816b816b816b816b816b816b816b816b816b816b816b816b816b816b816b816b816b816b816b816b816b816b816b816b816b816b816b816b816b816b816b817f6c7f6c7f6c7f6c7f6c7f6c7f6c7f6c7f6c7f6c7f6c7f6c7f6c7f6c7f6c7f6c7f6c7f6c7f6c7f6c7f6c7f6c7f6c7f6c7f6c7f6c7f6c7f6c7f6c7f6c7f6c7f6c7f6c7f6c7f6c7f7c6b7e7c6b7e7c6b7e7c6b7e7c6b7e7c6b7e7c6b7e7c6b7e7c6b7e7c6b7e7c6b7e7c6b7e7c6b7e7c6b7e7c6b7e7c6b7e7c6b7e7c6b7ea820d5c10a6bcb6f1122b49474822d75b6e27a704fd8525a96aec73b0a9c17adfb2a886c756c766c6c7b7b52796c6c6c76094b5341412d53504f548778094b5341412d53504f54879b696c6c6c6c7604474d4558886c756c76055452414445886c756c756c755f7901687f77517f7c7f7c03adadac88587f7c8151881e546f6b656e6f76617465206861766520637265617465642061206e6577207b7e17206576656e74206265747765656e3a0a42757965723a207e5b7a7e120a427579657220637573746f6469616e3a207e5a7a7e0d0a616e640a53656c6c65723a207e597a7e130a53656c6c657220637573746f6469616e3a207e587a7e060a0a666f72207e557a7e0720756e697473207e557a7e102061742061207072696365206f6620247e547a7e1f2055534420656163682e0a0a5468652074726164652076656e7565206973207e7b7e1e0a0a546869732074726164652077617320636f6d706c65746564206174207e7b7e470a0a466f72206d6f726520696e666f726d6174696f6e207365653a2068747470733a2f2f726563656970747669657765722e746f6b656e6f766174652e636f6d2f74726164652f7e7b7e4c540a4b5341412d53706f742074726164652056657273696f6e3a20302e320a54686973207472616465206973207265636f72646564206f6e2074686520546f6b656e6f76617465207472616465206c65646765722e7e827c7e03006a4d7c7e82090000000000000000fd7c7e7c7eaa7c547f7c836901207f7588517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f00012180607e007b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b93770800000000000000007e01217f757c517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f00012180607e007b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b937b760130013aa5630130677601610167a56901576894817b5498817d957b93770800000000000000007e01217f757b7c547aaa517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e01007e011f7f817e00011f80517e9321414136d08c5ed2bf3ba048afe6dcaebafeffffffffffffffffffffffffffffff007d5296789f637897785296789f639467776867776876927f76927f76927f76927f76927f76927f76927f76927f76927f76927f76927f76927f76927f76927f76927f76927f76927f76927f76927f76927f76927f76927f76927f76927f76927f76927f76927f76927f76927f76927f76927f7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e827c7e23022079be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798027c7e827c7e01307c7e01c37e2102b405d7f0322a89d0f9f3a98e6f938fdc1c969a8d1382a2bf66a71ae74a1e83b020b204c9123d063d0863b3422c5345cc7f73f564e66040c74c0f687b90366ed81520260e588c8eed91df4813e5dd18388c91878b888df5a7a908bf7ce0aa76243c442463f72ad3f50dc949c895babb7d704b86df66210395ec3440cf2a7fdb1e89df7302000000756dabadadac00000000`

	prevTxHex3 = `020000000466480a3bea1f7ff2d1bd9b97f58b4ff8c7071630c82587466e74394d3d72534a010000008c483045022100ea1f5ca3c76565f9b5df4ef7aaa44dd12061377f778d98052263e689092f562a02202948c67add07a199631955db0a38b88d998bcd543f70bd15aef6da3caf63b4edc32103901b72cffbf358b2f8dd3887d6645372bd041d4e47d18155ec0ef3e184086b5620c8925e7b008668089d3ae1fc1cf450f7f45f0b4af43cd7d30b84446ecb374d6dffffffff8b79ec45e4d35784691b30dff3b4ab3932894cc1d121200e18c6620709f5fc81040000006b483045022100e387c4b8de5d6def9563ad8165fd16628521d62e9170a2f2e6c25d0fa0b0457a0220486ab708f24dda1061cdde92907ab5bda5558bb0eca19c3899970327008a4a31412102673faf7ee3a8c9953e4a596dedb2aa620451259975cb759d2fb1cc76e98f695dffffffff8e46ed77a8fd9fef7d619f34b979220f73b740b907ecabe26260d13f40806679020000006a4730440220737dbdc0b328e33b45d43db1da46aa4d75ec12fb49c147baefd7fa994510868402205d8cfa4d0a3fc6d25bfbd52e41e15e68a6c6e51831eb2d0ed34c006bf9cdccf94121036b38d33ab9e9579325762f4d0e595537a6379d73c1b977f70a046195f663b18affffffffaa70fdf6a7b1957fc9c7644250f4005914c219c84c95bf04a21e9fee06542c62010000006b48304502210093310204d3b2e01b385d32a76278e31b6b226fbd569d18a1a2efbc52f8b36fc702203949236bab80e2c669309da070434d08d2e50b0332ae295eca422fd49bbe88e1412102207d0891b88c096f1f8503a684c387b4f9440c80707118ec14227adadd15db78ffffffff0538538202000000001976a9140eab67faf587701349e4e99f0914c2e064a8e93288ac8408000000000000fd6103a914179b4c7a45646a509473df5a444b6e18b723bd148876a9142e0fa8744508c13de3fe065d7ed2016370cc433f88ac6a4d2d037b227469746c65223a2246726f672043617274656c202331373935222c226465736372697074696f6e223a2246726f6773206d75737420756e69746520746f2064657374726f7920746865206c697a617264732e20446f20796f75206861766520776861742069742074616b65733f222c22696d616765223a22623a2f2f61353333663036313134353665333438326536306136666433346337663165366265393365663134303261396639363139313539306334303534326230306335222c226e756d626572223a313739352c22736572696573223a333639302c2273636f7265223a2235392e3131222c2272616e6b223a333033382c22726172697479223a22436f6d6d6f6e222c2261747472696275746573223a5b7b2274726169745f74797065223a224261636b67726f756e64222c2276616c7565223a225465616c204a756d626c65222c22636f756e74223a3131352c22726172697479223a22556e636f6d6d6f6e227d2c7b2274726169745f74797065223a2246726f67222c2276616c7565223a22526574726f20426c7565222c22636f756e74223a3433322c22726172697479223a22436f6d6d6f6e227d2c7b2274726169745f74797065223a22426f6479222c2276616c7565223a22507572706c6520466c616e6e656c222c22636f756e74223a36342c22726172697479223a22436f6d6d6f6e227d2c7b2274726169745f74797065223a224d6f757468222c2276616c7565223a224e6f204d6f757468204974656d222c22636f756e74223a313335382c22726172697479223a22436f6d6d6f6e227d2c7b2274726169745f74797065223a2245796573222c2276616c7565223a224f72616e676520457965205061746368222c22636f756e74223a3130332c22726172697479223a2252617265227d2c7b2274726169745f74797065223a2248656164222c2276616c7565223a2250657420436869636b222c22636f756e74223a36392c22726172697479223a22436f6d6d6f6e227d2c7b2274726169745f74797065223a2248616e64222c2276616c7565223a224e6f2048616e64204974656d222c22636f756e74223a3939322c22726172697479223a22436f6d6d6f6e227d5d7d68b11900000000001976a91497df51a1dea118bd689099125b42d75e48d2f5ec88aca8e91300000000001976a914b002bc36771ade5ec8af85d921e25ea20e278fb188aca41f5603000000001976a914ebd2539da5cea9507decf76f6a28140a86624fe188ac00000000`

	txHex3 = `02000000021c2bff8cb2e37f9018ee6e47512492ee65fa2012ce6c5998b6a2e9583974dabc010000008b473044022029d0a05f2601ee89d63e7a61a8f5877f20e7a48214d3aa6e8421bb938feec8a80220785478ad3019ec91c5545199539ccfd5704021f1c962becd48e0264f7e16de86c32102207d0891b88c096f1f8503a684c387b4f9440c80707118ec14227adadd15db7820c8925e7b008668089d3ae1fc1cf450f7f45f0b4af43cd7d30b84446ecb374d6dffffffff215b80a60dc756a488066fa95b90cceec4fd731ef489d51047b41e7aa5a95bf0040000006a47304402203951e4ebccaa652e360d8b2fab2ea9936a1eec19f27d6a1d9791c32b4e46540e02202529a8af4795bcf7dfe9dbb4826bb9f1467cc255de947e8c07a5961287aa713e41210253fe24fd82a07d02010d9ca82f99870c0e5e7402a9b26c9d25ae753e40754c4dffffffff0544ca0203000000001976a9142e0fa8744508c13de3fe065d7ed2016370cc433f88ac8408000000000000fd6103a914179b4c7a45646a509473df5a444b6e18b723bd148876a91497e5faf26e48d9015269c2592c6e4886ac2d161288ac6a4d2d037b227469746c65223a2246726f672043617274656c202331373935222c226465736372697074696f6e223a2246726f6773206d75737420756e69746520746f2064657374726f7920746865206c697a617264732e20446f20796f75206861766520776861742069742074616b65733f222c22696d616765223a22623a2f2f61353333663036313134353665333438326536306136666433346337663165366265393365663134303261396639363139313539306334303534326230306335222c226e756d626572223a313739352c22736572696573223a333639302c2273636f7265223a2235392e3131222c2272616e6b223a333033382c22726172697479223a22436f6d6d6f6e222c2261747472696275746573223a5b7b2274726169745f74797065223a224261636b67726f756e64222c2276616c7565223a225465616c204a756d626c65222c22636f756e74223a3131352c22726172697479223a22556e636f6d6d6f6e227d2c7b2274726169745f74797065223a2246726f67222c2276616c7565223a22526574726f20426c7565222c22636f756e74223a3433322c22726172697479223a22436f6d6d6f6e227d2c7b2274726169745f74797065223a22426f6479222c2276616c7565223a22507572706c6520466c616e6e656c222c22636f756e74223a36342c22726172697479223a22436f6d6d6f6e227d2c7b2274726169745f74797065223a224d6f757468222c2276616c7565223a224e6f204d6f757468204974656d222c22636f756e74223a313335382c22726172697479223a22436f6d6d6f6e227d2c7b2274726169745f74797065223a2245796573222c2276616c7565223a224f72616e676520457965205061746368222c22636f756e74223a3130332c22726172697479223a2252617265227d2c7b2274726169745f74797065223a2248656164222c2276616c7565223a2250657420436869636b222c22636f756e74223a36392c22726172697479223a22436f6d6d6f6e227d2c7b2274726169745f74797065223a2248616e64222c2276616c7565223a224e6f2048616e64204974656d222c22636f756e74223a3939322c22726172697479223a22436f6d6d6f6e227d5d7de4d41e00000000001976a91497df51a1dea118bd689099125b42d75e48d2f5ec88ac30e51700000000001976a91484c9b30c0e3529a6d260b361f70902f962d4b77088acec93e340000000001976a914863f485dae59224cc5993b26bf50da2e7c368c8a88ac00000000`
)

func TestExecute(t *testing.T) {
	tt := []struct {
		name      string
		txHex     string
		prevTxHex string
	}{
		{
			name:      "OP_CODESEPARATOR parsing",
			txHex:     txHex1,
			prevTxHex: prevTxHex1,
		},
		{
			name:      "OP_INVERT shouldn't modify items other than the top value of the stack",
			txHex:     txHex2,
			prevTxHex: prevTxHex2,
		},
		{
			name:      "failing",
			txHex:     txHex3,
			prevTxHex: prevTxHex3,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			tx, err := bt.NewTxFromString(tc.txHex)
			require.NoError(t, err)

			beforeScript, _ := tx.Inputs[0].UnlockingScript.ToASM()

			prevTx, err := bt.NewTxFromString(tc.prevTxHex)
			require.NoError(t, err)

			inputIdx := 0
			input := tx.InputIdx(inputIdx)
			prevOutput := prevTx.OutputIdx(int(input.PreviousTxOutIndex))

			err = NewEngine().Execute(
				WithTx(tx, inputIdx, prevOutput),
				WithForkID(),
				WithAfterGenesis(),
			)

			afterScript, _ := tx.Inputs[0].UnlockingScript.ToASM()

			require.Equal(t, beforeScript, afterScript)

			require.NoError(t, err)
		})
	}
}
