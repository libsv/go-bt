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
