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

	res := scriptPub.IsP2PKH()

	if !res {
		t.Errorf("Expected %t, got %t", true, res)
	}
}

func TestIsPublicKeyOut(t *testing.T) {
	b, _ := hex.DecodeString("2102f0d97c290e79bf2a8660c406aa56b6f189ff79f2245cc5aff82808b58131b4d5ac")
	scriptPub := script.NewFromBytes(b)

	res := scriptPub.IsP2PK()

	if !res {
		t.Errorf("Expected %t, got %t", true, res)
	}
}

func TestIsScriptHashOut(t *testing.T) {
	b, _ := hex.DecodeString("a9149de5aeaff9c48431ba4dd6e8af73d51f38e451cb87")
	scriptPub := script.NewFromBytes(b)

	res := scriptPub.IsP2SH()

	if !res {
		t.Errorf("Expected %t, got %t", true, res)
	}
}

func TestIsScriptData(t *testing.T) {
	b, _ := hex.DecodeString("006a04ac1eed884d53027b2276657273696f6e223a22302e31222c22686569676874223a3634323436302c22707265764d696e65724964223a22303365393264336535633366376264393435646662663438653761393933393362316266623366313166333830616533306432383665376666326165633561323730222c22707265764d696e65724964536967223a2233303435303232313030643736333630653464323133333163613836663031386330343665353763393338663139373735303734373333333533363062653337303438636165316166333032323030626536363034353430323162663934363465393966356139353831613938633963663439353430373539386335396234373334623266646234383262663937222c226d696e65724964223a22303365393264336535633366376264393435646662663438653761393933393362316266623366313166333830616533306432383665376666326165633561323730222c2276637478223a7b2274784964223a2235373962343335393235613930656533396133376265336230306239303631653734633330633832343133663664306132303938653162656137613235313566222c22766f7574223a307d2c226d696e6572436f6e74616374223a7b22656d61696c223a22696e666f407461616c2e636f6d222c226e616d65223a225441414c20446973747269627574656420496e666f726d6174696f6e20546563686e6f6c6f67696573222c226d65726368616e74415049456e64506f696e74223a2268747470733a2f2f6d65726368616e746170692e7461616c2e636f6d2f227d7d46304402206fd1c6d6dd32cc85ddd2f30bc068445dd901c6bd85e394e45bb254716d2bb228022041f0f8b1b33c2e3702aee4ad47155548045ed945738b43dc0faed2e86faa12e4")
	scriptPub := script.NewFromBytes(b)

	res := scriptPub.IsData()

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
