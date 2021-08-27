// Copyright (c) 2013-2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package interpreter

import (
	"testing"

	"github.com/libsv/go-bt/v2/bscript"
)

// TestOpcodeDisabled tests the opcodeDisabled function manually because all
// disabled opcodes result in a script execution failure when executed normally,
// so the function is not called under normal circumstances.
func TestOpcodeDisabled(t *testing.T) {
	t.Parallel()

	tests := []byte{bscript.Op2MUL, bscript.Op2DIV}
	for _, opcodeVal := range tests {
		pop := ParsedOp{Op: opcodeArray[opcodeVal], Data: nil}
		err := opcodeDisabled(&pop, nil)
		if !IsErrorCode(err, ErrDisabledOpcode) {
			t.Errorf("opcodeDisabled: unexpected error - got %v, "+
				"want %v", err, ErrDisabledOpcode)
			continue
		}
	}
}
