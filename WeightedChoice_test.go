package cryptolib

import (
	"testing"
)

// var choices []Choice = {
// 	Choice{Weight: 1, Item: "one"},
// 	Choice{Weight: 2, Item: "two"},
// 	Choice{Weight: 3, Item: "three"},
// }

func TestOneChoices(t *testing.T) {
	choices := []Choice{
		Choice{Weight: 1, Item: "one"},
	}

	choice, err := WeightedChoice(choices)
	if err != nil {
		t.Error(err)
	}

	if choice.Weight != 1 {
		t.Errorf("Weight should be zero, got %d", choice.Weight)
	}

	if choice.Item != "one" {
		t.Errorf("Item should be 'one', got '%s'", choice.Item)
	}
}

func TestMultiWithZeroChoices(t *testing.T) {
	choices := []Choice{
		Choice{Weight: 1, Item: "one"},
		Choice{Weight: 0, Item: "zero"},
	}

	choice, err := WeightedChoice(choices)
	if err != nil {
		t.Error(err)
	}

	if choice.Weight != 1 {
		t.Errorf("Weight should be zero, got %d", choice.Weight)
	}

	if choice.Item != "one" {
		t.Errorf("Item should be 'one', got '%s'", choice.Item)
	}
}

func TestMultiChoices(t *testing.T) {
	choices := []Choice{
		Choice{Weight: 1, Item: "one"},
		Choice{Weight: 10000000, Item: "infinity"},
	}

	choice, err := WeightedChoice(choices)
	if err != nil {
		t.Error(err)
	}

	if choice.Weight != 10000000 {
		t.Errorf("Weight should be 10000000, got %d", choice.Weight)
	}

	if choice.Item != "infinity" {
		t.Errorf("Item should be 'infinity', got '%s'", choice.Item)
	}
}

func TestZeroChoices(t *testing.T) {
	choices := []Choice{
		Choice{Weight: 0, Item: "one"},
		Choice{Weight: 0, Item: "two"},
		Choice{Weight: 1, Item: "three"},
	}

	choice, err := WeightedChoice(choices)
	if err != nil {
		t.Error(err)
	}

	if choice.Weight != 0 {
		t.Errorf("Weight should be zero, got %d", choice.Weight)
	}
}
