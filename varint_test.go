package bt_test

import (
	"testing"

	"github.com/libsv/go-bt"
	"github.com/stretchr/testify/assert"
)

func TestDecodeVarInt(t *testing.T) {
	// todo: create test(s)
}

func TestVarIntUpperLimitInc(t *testing.T) {
	// todo: create test(s)
}

func TestVarInt(t *testing.T) {
	t.Parallel()

	var varIntTests = []struct {
		testName    string
		input       uint64
		expectedLen int
	}{
		{"VarInt 1 byte Lower", 0, 1},
		{"VarInt 1 byte Upper", 252, 1},
		{"VarInt 3 byte Lower", 253, 3},
		{"VarInt 3 byte Upper", 65535, 3},
		{"VarInt 5 byte Lower", 65536, 5},
		{"VarInt 5 byte Upper", 4294967295, 5},
		{"VarInt 9 byte Lower", 4294967296, 9},
		{"VarInt 9 byte Upper", 18446744073709551615, 9},
	}

	for _, varIntTest := range varIntTests {
		t.Run(varIntTest.testName, func(t *testing.T) {
			assert.Equal(t, varIntTest.expectedLen, len(bt.VarInt(varIntTest.input)))
		})
	}
}
