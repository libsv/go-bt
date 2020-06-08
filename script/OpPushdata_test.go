package script_test

import (
	"encoding/hex"
	"testing"

	"github.com/libsv/libsv/script"
)

func TestDecodePartsSimple(t *testing.T) {
	s := "05000102030401FF02ABCD"
	parts, err := script.DecodeStringParts(s)
	if err != nil {
		t.Error(err)
	}

	if len(parts) != 3 {
		t.Errorf("Expected 3 parts, got %d", len(parts))
	}
}

func TestDecodePartsSimpleAndEncode(t *testing.T) {
	s := "05000102030401FF02ABCD"
	parts, err := script.DecodeStringParts(s)
	if err != nil {
		t.Error(err)
	}

	if len(parts) != 3 {
		t.Errorf("Expected 3 parts, got %d", len(parts))
	}

	p, err := script.EncodeParts(parts)
	if err != nil {
		t.Error(err)
	}

	h := hex.EncodeToString(p)

	expected := "05000102030401ff02abcd"
	if h != expected {
		t.Errorf("Expected %q, got %q", expected, h)
	}
}

func TestDecodePartsEmpty(t *testing.T) {
	s := ""
	parts, err := script.DecodeStringParts(s)
	if err != nil {
		t.Error(err)
	}

	if len(parts) != 0 {
		t.Errorf("Expected [], got %+v", parts)
	}
}

func TestDecodePartsComplex(t *testing.T) {
	s := "524c53ff0488b21e000000000000000000362f7a9030543db8751401c387d6a71e870f1895b3a62569d455e8ee5f5f5e5f03036624c6df96984db6b4e625b6707c017eb0e0d137cd13a0c989bfa77a4473fd000000004c53ff0488b21e0000000000000000008b20425398995f3c866ea6ce5c1828a516b007379cf97b136bffbdc86f75df14036454bad23b019eae34f10aff8b8d6d8deb18cb31354e5a169ee09d8a4560e8250000000052ae"
	parts, err := script.DecodeStringParts(s)
	if err != nil {
		t.Error(err)
	}

	if len(parts) != 5 {
		t.Errorf("Expected 5 parts, got %d", len(parts))
	}
}
