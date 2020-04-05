package libsv

import (
	"encoding/hex"
	"fmt"
	"testing"
)

// Pay-to-PubKeyHash address
func TestP2PKHAddressToScript(t *testing.T) {
	script, err := AddressToScript("1DkmRkb5iQFkDu4NBysog5bugnsyx7kwtn")
	if err != nil {
		t.Error(err)
	} else {
		h := hex.EncodeToString(script)
		expected := "76a9148be87b3978d8ef936b30ddd4ed903f8da7abd27788ac"
		if h != expected {
			t.Errorf("Expected %s, got %s", expected, h)
		}
	}
}

// Pay-to-ScriptHash address
func TestP2SHAddressToScript(t *testing.T) {
	script, err := AddressToScript("37BvY7rFguYQvEL872Y7Fo77Y3EBApC2EK")
	if err != nil {
		t.Error(err)
	} else {
		h := hex.EncodeToString(script)
		expected := "a9143c5031fd7b3f8dfc4aef2d98b76e74b1bb7a51b887"
		if h != expected {
			t.Errorf("Expected %s, got %s", expected, h)
		}
	}
}

func TestShortAddressToScript(t *testing.T) {
	_, err := AddressToScript("ADD8E55")
	if err == nil {
		t.Errorf("Expected an error")
	} else {
		expected := "invalid address length for 'ADD8E55'"
		if fmt.Sprint(err) != expected {
			t.Errorf("Expected %s, got %s", expected, err)
		}
	}
}

func TestUnsupportedAddressToScript(t *testing.T) {
	_, err := AddressToScript("27BvY7rFguYQvEL872Y7Fo77Y3EBApC2EK")
	if err == nil {
		t.Errorf("Expected an error")
	} else {
		expected := "Address 27BvY7rFguYQvEL872Y7Fo77Y3EBApC2EK is not supported"
		if fmt.Sprint(err) != expected {
			t.Errorf("Expected %s, got %s", expected, err)
		}
	}
}
