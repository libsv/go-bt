package transaction

import (
	"encoding/hex"
	"testing"
)

const output = "8a08ac4a000000001976a9148bf10d323ac757268eb715e613cb8e8e1d1793aa88ac00000000"

func TestNewOutput(t *testing.T) {
	bytes, _ := hex.DecodeString(output)
	o, s := NewOutputFromBytes(bytes)

	// t.Errorf("\n%s\n", o)
	if s != 34 {
		t.Errorf("Expected 25, got %d", s)
	}

	if o.Value != 1252788362 {
		t.Errorf("Expected 1252788362, got %d", o.Value)
	}

	if len(o.Script) != 25 {
		t.Errorf("Expected 25, got %d", len(o.Script))
	}

	script := hex.EncodeToString(o.Script)
	if script != "76a9148bf10d323ac757268eb715e613cb8e8e1d1793aa88ac" {
		t.Errorf("Expected 76a9148bf10d323ac757268eb715e613cb8e8e1d1793aa88ac, got %x", script)
	}
}

func TestNewOutputForPublicKeyHash(t *testing.T) {
	publicKeyhash := "8fe80c75c9560e8b56ed64ea3c26e18d2c52211b" // This is the PKH for address mtdruWYVEV1wz5yL7GvpBj4MgifCB7yhPd
	value := uint64(5000)
	output, err := NewOutputForPublicKeyHash(publicKeyhash, value)
	if err != nil {
		t.Error("Error")
	}
	expected := "76a9148fe80c75c9560e8b56ed64ea3c26e18d2c52211b88ac"
	if hex.EncodeToString(output.Script) != expected {
		t.Errorf("Error script not correct\nExpected: %s\n     Got: %s\n", hex.EncodeToString(output.Script), expected)
	}
}

func TestNewOutputOpReturn(t *testing.T) {
	data := "This is some test data"
	dataBytes := []byte(data)
	output, err := NewOutputOpReturn(dataBytes)
	if err != nil {
		t.Error(err)
		return
	}
	dataHexStr := hex.EncodeToString(dataBytes)
	script := hex.EncodeToString(output.Script)
	expectedScript := "006a16" + dataHexStr

	if script != expectedScript {
		t.Errorf("Error op return hex expected %s, got %s", expectedScript, script)
	}
}
