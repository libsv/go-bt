package bt_test

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/libsv/go-bt/v2"
	"github.com/stretchr/testify/assert"
)

func convertIntToBytes(int uint64) []byte {
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.LittleEndian, int); err != nil {
		return nil
	}
	return buf.Bytes()
}

func TestDecodeVarInt(t *testing.T) {
	t.Parallel()

	var tests = []struct {
		testName       string
		input          []byte
		expectedResult uint64
		expectedSize   int
	}{
		{"0xff", convertIntToBytes(0xff), 0, 9},
		{"0xfe", convertIntToBytes(0xfe), 0, 5},
		{"0xfd", convertIntToBytes(0xfd), 0, 3},
		{"1", convertIntToBytes(1), 1, 1},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			r, s := bt.DecodeVarInt(test.input)
			assert.Equal(t, test.expectedResult, r)
			assert.Equal(t, test.expectedSize, s)
		})
	}
}

func TestVarIntUpperLimitInc(t *testing.T) {
	t.Parallel()

	var tests = []struct {
		testName       string
		input          uint64
		expectedResult int
	}{
		{"0", 0, 0},
		{"10", 10, 0},
		{"100", 100, 0},
		{"252", 252, 2},
		{"65535", 65535, 2},
		{"4294967295", 4294967295, 4},
		{"18446744073709551615", 18446744073709551615, -1},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			r := bt.VarIntUpperLimitInc(test.input)
			assert.Equal(t, test.expectedResult, r)
		})
	}
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
