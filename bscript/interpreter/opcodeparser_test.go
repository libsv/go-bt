// Copyright (c) 2013-2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package interpreter

import (
	"testing"

	"github.com/libsv/go-bt/v2/bscript"
	"github.com/libsv/go-bt/v2/bscript/interpreter/errs"
	"github.com/stretchr/testify/require"
)

// TestOpcodeDisabled tests the opcodeDisabled function manually because all
// disabled opcodes result in a script execution failure when executed normally,
// so the function is not called under normal circumstances.
func TestOpcodeDisabled(t *testing.T) {
	t.Parallel()

	tests := []byte{bscript.Op2MUL, bscript.Op2DIV}
	for _, opcodeVal := range tests {
		pop := ParsedOpcode{op: opcodeArray[opcodeVal], Data: nil}
		err := opcodeDisabled(&pop, nil)
		if !errs.IsErrorCode(err, errs.ErrDisabledOpcode) {
			t.Errorf("opcodeDisabled: unexpected error - got %v, "+
				"want %v", err, errs.ErrDisabledOpcode)
			continue
		}
	}
}

func TestParse(t *testing.T) {
	tt := []struct {
		name            string
		scriptHexString string

		expectedParsedScript ParsedScript
	}{
		{
			name:            "op return",
			scriptHexString: "0168776a0024dc",

			expectedParsedScript: ParsedScript{
				ParsedOpcode{
					op: opcode{
						val:    bscript.OpDATA1,
						name:   "OP_DATA_1",
						length: 2,
						exec:   opcodePushData,
					},
					Data: []byte{bscript.OpENDIF},
				},
				ParsedOpcode{
					op: opcode{
						val:    bscript.OpNIP,
						name:   "OP_NIP",
						length: 1,
						exec:   opcodeNip,
					},
					Data: nil,
				},
				ParsedOpcode{
					op: opcode{
						val:    bscript.OpRETURN,
						name:   "OP_RETURN",
						length: 1,
						exec:   opcodeReturn,
					},
					Data: []byte{bscript.OpRETURN, bscript.Op0, bscript.OpDATA36, bscript.OpUNKNOWN220},
				},
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			s, err := bscript.NewFromHexString(tc.scriptHexString)
			require.NoError(t, err)

			codeParser := DefaultOpcodeParser{}
			p, err := codeParser.Parse(s)
			require.NoError(t, err)

			for i := range p {
				require.Equal(t, tc.expectedParsedScript[i].Data, p[i].Data)
				require.Equal(t, tc.expectedParsedScript[i].op.length, p[i].op.length)
				require.Equal(t, tc.expectedParsedScript[i].op.name, p[i].op.name)
				require.Equal(t, tc.expectedParsedScript[i].op.val, p[i].op.val)
			}
		})
	}
}
