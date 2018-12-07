package transaction

import (
	"encoding/hex"
	"testing"
)

const input = "4c6ec863cf3e0284b407a1a1b8138c76f98280812cb9653231f385a0305fc76f010000006b483045022100f01c1a1679c9437398d691c8497f278fa2d615efc05115688bf2c3335b45c88602201b54437e54fb53bc50545de44ea8c64e9e583952771fcc663c8687dc2638f7854121037e87bbd3b680748a74372640628a8f32d3a841ceeef6f75626ab030c1a04824fffffffff"

func TestNew(t *testing.T) {
	bytes, _ := hex.DecodeString(input)
	i, s := NewInput(bytes)

	if s != 148 {
		t.Errorf("Expected 148, got %d", s)
	}

	if i.previousTxOutIndex != 1 {
		t.Errorf("Expected 1, got %d", i.previousTxOutIndex)
	}

	if i.txInScriptLength != 107 {
		t.Errorf("Expected 107, got %d", i.txInScriptLength)
	}

	if i.sequenceNumber != 0xFFFFFFFF {
		t.Errorf("Expected 0xFFFFFFFF, got %x", i.sequenceNumber)
	}
	// t.Error(i)
}
