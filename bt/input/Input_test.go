package input_test

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/libsv/libsv/bt/input"
	"github.com/libsv/libsv/script"
)

const inputHexStr = "4c6ec863cf3e0284b407a1a1b8138c76f98280812cb9653231f385a0305fc76f010000006b483045022100f01c1a1679c9437398d691c8497f278fa2d615efc05115688bf2c3335b45c88602201b54437e54fb53bc50545de44ea8c64e9e583952771fcc663c8687dc2638f7854121037e87bbd3b680748a74372640628a8f32d3a841ceeef6f75626ab030c1a04824fffffffff"

func TestNew(t *testing.T) {
	b, _ := hex.DecodeString(inputHexStr)
	i, s, err := input.NewFromBytes(b)
	if err != nil {
		t.Errorf("Invalid inputHexStr")
	}
	// t.Errorf("\n%s\n", i)

	if s != 148 {
		t.Errorf("Expected 148, got %d", s)
	}

	if i.PreviousTxOutIndex != 1 {
		t.Errorf("Expected 1, got %d", i.PreviousTxOutIndex)
	}

	if len(*i.UnlockingScript) != 107 {
		t.Errorf("Expected 107, got %d", len(*i.UnlockingScript))
	}

	if i.SequenceNumber != 0xFFFFFFFF {
		t.Errorf("Expected 0xFFFFFFFF, got %x", i.SequenceNumber)
	}
}

func TestNewFromUTXO(t *testing.T) {
	i, err := input.NewFromUTXO("a61021694ee0fd7c3d441aab7b387e356f5552957d5a01705a66766fe86ec9e5", 4, 5064, "a9149cbe9f5e72fa286ac8a38052d1d5337aa363ea7f88ac", 0xffffffff)
	if err != nil {
		t.Error(err)
	}

	if i.PreviousTxID != "a61021694ee0fd7c3d441aab7b387e356f5552957d5a01705a66766fe86ec9e5" {
		t.Errorf("Expected 'a61021694ee0fd7c3d441aab7b387e356f5552957d5a01705a66766fe86ec9e5', got %s", i.PreviousTxID)
	}

	if i.PreviousTxOutIndex != 4 {
		t.Errorf("Expected 4, got %d", i.PreviousTxOutIndex)
	}

	if i.PreviousTxSatoshis != 5064 {
		t.Errorf("Expected 5064, got %d", i.PreviousTxSatoshis)
	}

	es, err := script.NewFromHexString("a9149cbe9f5e72fa286ac8a38052d1d5337aa363ea7f88ac")
	if err != nil {
		t.Error(err)
	}

	if !bytes.Equal(*i.PreviousTxScript, *es) {
		t.Errorf("Expected a9149cbe9f5e72fa286ac8a38052d1d5337aa363ea7f88ac, got %x", *i.PreviousTxScript)
	}

	if i.SequenceNumber != 0xFFFFFFFF {
		t.Errorf("Expected 0xFFFFFFFF, got %x", i.SequenceNumber)
	}
}
