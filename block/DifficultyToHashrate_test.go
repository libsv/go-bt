package block_test

import (
	"testing"

	"github.com/libsv/libsv/block"
	"github.com/libsv/libsv/utils"
)

func TestDifficultyToHashratBSV(t *testing.T) {
	a := block.DifficultyToHashrate("BSV", 22000, 7)
	b := utils.HumanHash(a)
	expected := "13.50 TH/s"
	if b != expected {
		t.Errorf("Failed to calculate hashrate, expected %s got %s", expected, b)
	}
}

func TestDifficultyToHashrateRSV(t *testing.T) {
	a := block.DifficultyToHashrate("RSV", 22000, 7)
	b := utils.HumanHash(a)
	expected := "6.29 kH/s"
	if b != expected {
		t.Errorf("Failed to calculate hashrate, expected %s got %s", expected, b)
	}
}
