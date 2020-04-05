package libsv

import (
	"encoding/hex"
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
// 	_, err := bt.CreateJob("BSV", "mxo3ES3wLp8UcejGNTtPH1BCwn8F8JZi37", "coingeek", 4)

// 	if err == nil {
// 		t.Errorf("Expected CreateJob to return an error")
// 	}
// }

func TestDecodePartsSimple(t *testing.T) {
	s := "05000102030401FF02ABCD"
	parts, err := DecodeStringParts(s)
	if err != nil {
		t.Error(err)
	}

	if len(parts) != 3 {
		t.Errorf("Expected 3 parts, got %d", len(parts))
	}
	// t.Logf("%+v", parts)
}

func TestDecodePartsSimpleAndEncode(t *testing.T) {
	s := "05000102030401FF02ABCD"
	parts, err := DecodeStringParts(s)
	if err != nil {
		t.Error(err)
	}

	if len(parts) != 3 {
		t.Errorf("Expected 3 parts, got %d", len(parts))
	}

	p, err := EncodeParts(parts)
	if err != nil {
		t.Error(err)
	}

	h := hex.EncodeToString(p)

	expected := "05000102030401ff02abcd"
	if h != expected {
		t.Errorf("Expected %q, got %q", expected, h)
	}
	// t.Logf("%x", p)
}

func TestDecodePartsEmpty(t *testing.T) {
	s := ""
	parts, err := DecodeStringParts(s)
	if err != nil {
		t.Error(err)
	}

	if len(parts) != 0 {
		t.Errorf("Expected [], got %+v", parts)
	}
}

func TestDecodePartsComplex(t *testing.T) {
	s := "524c53ff0488b21e000000000000000000362f7a9030543db8751401c387d6a71e870f1895b3a62569d455e8ee5f5f5e5f03036624c6df96984db6b4e625b6707c017eb0e0d137cd13a0c989bfa77a4473fd000000004c53ff0488b21e0000000000000000008b20425398995f3c866ea6ce5c1828a516b007379cf97b136bffbdc86f75df14036454bad23b019eae34f10aff8b8d6d8deb18cb31354e5a169ee09d8a4560e8250000000052ae"
	parts, err := DecodeStringParts(s)
	if err != nil {
		t.Error(err)
	}

	if len(parts) != 5 {
		t.Errorf("Expected 5 parts, got %d", len(parts))
	}

	t.Logf("%+v", parts)
}
