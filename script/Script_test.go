package script_test

import (
	"encoding/hex"
	"testing"

	"github.com/libsv/libsv/script"
)

func TestNewP2PKHScriptFromPubKeyStr(t *testing.T) {
	scriptP2PKH, err := script.NewP2PKHFromPubKeyStr("023717efaec6761e457f55c8417815505b695209d0bbfed8c3265be425b373c2d6")
	if err != nil {
		t.Error(err)
	}

	res := hex.EncodeToString(*scriptP2PKH)
	expected := "76a9144d5d1920331b71735a97a606d9734aed83cb3dfa88ac"

	if res != expected {
		t.Errorf("Expected %q, got %q", expected, res)
	}
}

func TestNewFromHexString(t *testing.T) {
	s, err := script.NewFromHexString("76a914e2a623699e81b291c0327f408fea765d534baa2a88ac")
	if err != nil {
		t.Error(err)
	}

	res := hex.EncodeToString(*s)
	expected := "76a914e2a623699e81b291c0327f408fea765d534baa2a88ac"

	if res != expected {
		t.Errorf("Expected %q, got %q", expected, res)
	}
}

func TestNewFromASM(t *testing.T) {
	s, err := script.NewFromHexString("76a914e2a623699e81b291c0327f408fea765d534baa2a88ac")
	if err != nil {
		t.Error(err)
	}

	res, err := s.ToASM()
	if err != nil {
		t.Error(err)
	}
	expected := "OP_DUP OP_HASH160 e2a623699e81b291c0327f408fea765d534baa2a OP_EQUALVERIFY OP_CHECKSIG"

	if res != expected {
		t.Errorf("Expected %q, got %q", expected, res)
	}
}

func TestToASM(t *testing.T) {
	s, err := script.NewFromASM("OP_DUP OP_HASH160 e2a623699e81b291c0327f408fea765d534baa2a OP_EQUALVERIFY OP_CHECKSIG")
	if err != nil {
		t.Error(err)
	}

	res := hex.EncodeToString(*s)
	expected := "76a914e2a623699e81b291c0327f408fea765d534baa2a88ac"

	if res != expected {
		t.Errorf("Expected %q, got %q", expected, res)
	}
}

func TestIsPublicKeyHashOut(t *testing.T) {
	b, _ := hex.DecodeString("76a91403ececf2d12a7f614aef4c82ecf13c303bd9975d88ac")
	scriptPub := script.NewFromBytes(b)

	res := scriptPub.IsPublicKeyHashOut()

	if !res {
		t.Errorf("Expected %t, got %t", true, res)
	}
}

func TestIsPublicKeyOut(t *testing.T) {
	b, _ := hex.DecodeString("2102f0d97c290e79bf2a8660c406aa56b6f189ff79f2245cc5aff82808b58131b4d5ac")
	scriptPub := script.NewFromBytes(b)

	res := scriptPub.IsPublicKeyOut()

	if !res {
		t.Errorf("Expected %t, got %t", true, res)
	}
}

func TestIsScriptHashOut(t *testing.T) {
	b, _ := hex.DecodeString("a9149de5aeaff9c48431ba4dd6e8af73d51f38e451cb87")
	scriptPub := script.NewFromBytes(b)

	res := scriptPub.IsScriptHashOut()

	if !res {
		t.Errorf("Expected %t, got %t", true, res)
	}
}

func TestIsMultisigOut(t *testing.T) { // TODO: check this
	b, _ := hex.DecodeString("5201110122013353ae")
	scriptPub := script.NewFromBytes(b)

	res := scriptPub.IsMultisigOut()

	if !res {
		t.Errorf("Expected %t, got %t", true, res)
	}
}

func TestGetPublicKeyHash(t *testing.T) {
	b, _ := hex.DecodeString("76a91404d03f746652cfcb6cb55119ab473a045137d26588ac")
	s := script.NewFromBytes(b)

	pkh, err := s.GetPublicKeyHash()
	if err != nil {
		t.Error(err)
	}
	res := hex.EncodeToString(pkh)

	expected := "04d03f746652cfcb6cb55119ab473a045137d265"

	if res != expected {
		t.Errorf("Expected %q, got %q", res, expected)
	}
}

func TestGetPublicKeyHashAsString(t *testing.T) {

	s, err := script.NewFromHexString("76a91404d03f746652cfcb6cb55119ab473a045137d26588ac")
	if err != nil {
		t.Error(err)
	}

	pkh, err := s.GetPublicKeyHash()
	if err != nil {
		t.Error(err)
	}
	res := hex.EncodeToString(pkh)

	expected := "04d03f746652cfcb6cb55119ab473a045137d265"

	if res != expected {
		t.Errorf("Expected %q, got %q", res, expected)
	}
}

func TestGetPublicKeyHashEmptyScript(t *testing.T) {
	s := &script.Script{}

	_, err := s.GetPublicKeyHash()
	if err == nil {
		t.Error("Expected 'Script is empty'")
	}
}
