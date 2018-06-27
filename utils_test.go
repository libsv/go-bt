package cryptolib

import (
	"testing"
)

func TestReverseHexString(t *testing.T) {

	expectedEven := "3910"
	expectedOdd := "391004"
	expectedLong := "4512710951122431"

	rhEven := ReverseHexString("1039")

	if rhEven != expectedEven {
		t.Errorf("Expected reversed string to be '%+v', got %+v", expectedEven, rhEven)
	}

	rhOdd := ReverseHexString("41039")
	if rhOdd != expectedOdd {
		t.Errorf("Expected reversed string to be '%+v', got %+v", expectedOdd, rhOdd)
	}

	rhLong := ReverseHexString("3124125109711245")

	if rhLong != expectedLong {
		t.Errorf("Expected reversed string to be '%+v', got %+v", expectedLong, rhLong)
	}
}

func TestDifficultyFromBits(t *testing.T) {
	// genisis block should be difficulty 1
	testDifficulty("1d00ffff", float64(1), t)
	testDifficulty("1745fb53", float64(4.022059196164954e+12), t)
	testDifficulty("207fffff", float64(4.6565423739069247e-10), t)
}

func testDifficulty(bits string, expected float64, t *testing.T) {
	d, _ := DifficultyFromBits(bits)

	if d != expected {
		t.Errorf("Expected difficulty of '%s' to be '%v', got %v", bits, expected, d)
	}
}

// func TestTransactionVersion(t *testing.T) {
// 	// transaction version 1, toInternalByteOrder 01000000
// 	//  transaction version 2, toInternalByteOrder 02000000
// 	// 17dec =  11, toInternalByteOrder 11000000

// 	b := getLittleEndianBytes(1, 4)
// 	expected := []byte{1, 0, 0, 0}
// 	if !reflect.DeepEqual(b, expected) {
// 		t.Errorf("Expected transaction verison of '%s' to be '%+v', got %+v", "1", expected, b)
// 	}

// 	if len(b) != 4 {
// 		t.Errorf("Expected transaction version length to be 4, got %+v", len(b))
// 	}
// }

// func TestCreateJobError(t *testing.T) {
// 	bt := BlockTemplate{}
// 	_, err := bt.CreateJob("BCH", "mxo3ES3wLp8UcejGNTtPH1BCwn8F8JZi37", "coingeek", 4)

// 	if err == nil {
// 		t.Errorf("Expected CreateJob to return an error")
// 	}
// }
