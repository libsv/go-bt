package utils_test

import (
	"testing"

	"github.com/libsv/libsv/utils"
)

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

func TestHashFunctions(t *testing.T) {

	for _, varIntTest := range varIntTests {
		t.Run(varIntTest.testName, func(t *testing.T) {

			b := utils.VarInt(varIntTest.input)

			if len(b) != varIntTest.expectedLen {
				t.Errorf("Expected length to be '%+v', got %+v", varIntTest.expectedLen, len(b))
			}

		})
	}
}
