package cryptolib

import (
	"testing"
)

func TestDifficultyToHashratBCH(t *testing.T) {
	a := DifficultyToHashrate("BCH", 22000, 7)
	b := HumanHash(a)
	expected := "13.50 TH/s"
	if b != expected {
		t.Errorf("Failed to calculate hashrate, expected %s got %s", expected, b)
	}
}

func TestDifficultyToHashrateRCH(t *testing.T) {
	a := DifficultyToHashrate("RCH", 22000, 7)
	b := HumanHash(a)
	expected := "6.29 kH/s"
	if b != expected {
		t.Errorf("Failed to calculate hashrate, expected %s got %s", expected, b)
	}
}
