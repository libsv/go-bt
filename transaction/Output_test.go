package transaction

import (
	"encoding/hex"
	"testing"
)

const output = "8a08ac4a000000001976a9148bf10d323ac757268eb715e613cb8e8e1d1793aa88ac00000000"

func TestNewOutput(t *testing.T) {
	bytes, _ := hex.DecodeString(output)
	o, s := NewOutput(bytes)

	// t.Errorf("\n%s\n", o)
	if s != 34 {
		t.Errorf("Expected 25, got %d", s)
	}

	if o.value != 1252788362 {
		t.Errorf("Expected 1252788362, got %d", o.value)
	}

	if o.scriptLen != 25 {
		t.Errorf("Expected 25, got %d", o.scriptLen)
	}

	script := hex.EncodeToString(o.script)
	if script != "76a9148bf10d323ac757268eb715e613cb8e8e1d1793aa88ac" {
		t.Errorf("Expected 76a9148bf10d323ac757268eb715e613cb8e8e1d1793aa88ac, got %x", script)
	}
}
