package bscript_test

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/bitcoinsv/bsvd/bsvec"
	"github.com/libsv/go-bt/bscript"
)

const testPublicKeyHash = "00ac6144c4db7b5790f343cf0477a65fb8a02eb7"

func TestNewFromStringMainnet(t *testing.T) {
	addressMain := "1E7ucTTWRTahCyViPhxSMor2pj4VGQdFMr"
	expectedPublicKeyhash := "8fe80c75c9560e8b56ed64ea3c26e18d2c52211b"

	addr, err := bscript.NewAddressFromString(addressMain)
	if err != nil {
		t.Error(err)
	}

	if addr.PublicKeyHash != expectedPublicKeyhash {
		t.Errorf("PKH from Main address %s incorrect,\ngot: %s\nexpected: %s", addressMain, addr.PublicKeyHash, expectedPublicKeyhash)
	}

	if addr.AddressString != addressMain {
		t.Errorf("Address from Main address %s incorrect,\ngot: %s\nexpected: %s", addressMain, addr.AddressString, addressMain)
	}
}

func TestNewFromStringTestnet(t *testing.T) {
	addressTestnet := "mtdruWYVEV1wz5yL7GvpBj4MgifCB7yhPd"
	expectedPublicKeyhash := "8fe80c75c9560e8b56ed64ea3c26e18d2c52211b"

	addr, err := bscript.NewAddressFromString(addressTestnet)
	if err != nil {
		t.Error(err)
	}

	if addr.PublicKeyHash != expectedPublicKeyhash {
		t.Errorf("PKH from Main address %s incorrect,\ngot: %s\nexpected: %s", addressTestnet, addr.PublicKeyHash, expectedPublicKeyhash)
	}

	if addr.AddressString != addressTestnet {
		t.Errorf("Address from Main address %s incorrect,\ngot: %s\nexpected: %s", addressTestnet, addr.AddressString, addressTestnet)
	}
}

func TestNewFromStringShortAddress(t *testing.T) {
	_, err := bscript.NewAddressFromString("ADD8E55")
	if err == nil {
		t.Errorf("Expected an error")
	} else {
		expected := "invalid address length for 'ADD8E55'"
		if fmt.Sprint(err) != expected {
			t.Errorf("Expected %s, got %s", expected, err)
		}
	}
}

func TestNewFromStringUnsupportedAddress(t *testing.T) {
	_, err := bscript.NewAddressFromString("27BvY7rFguYQvEL872Y7Fo77Y3EBApC2EK")
	if err == nil {
		t.Errorf("Expected an error")
	} else {
		expected := "Address 27BvY7rFguYQvEL872Y7Fo77Y3EBApC2EK is not supported"
		if fmt.Sprint(err) != expected {
			t.Errorf("Expected %s, got %s", expected, err)
		}
	}
}

func TestNewFromPublicKeyStringMainnet(t *testing.T) {
	pubKey := "026cf33373a9f3f6c676b75b543180703df225f7f8edbffedc417718a8ad4e89ce"
	expectedAddress := "114ZWApV4EEU8frr7zygqQcB1V2BodGZuS"

	addr, err := bscript.NewAddressFromPublicKeyString(pubKey, true)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	if addr.PublicKeyHash != testPublicKeyHash {
		t.Errorf("PKH is incorrect,\ngot: %s\nexpected: %s", addr.PublicKeyHash, testPublicKeyHash)
	}

	if addr.AddressString != expectedAddress {
		t.Errorf("Address is incorrect,\ngot: %s\nexpected: %s", addr.AddressString, expectedAddress)
	}
}

func TestNewFromPublicKeyStringTestnet(t *testing.T) {
	pubKey := "026cf33373a9f3f6c676b75b543180703df225f7f8edbffedc417718a8ad4e89ce"
	expectedAddress := "mfaWoDuTsFfiunLTqZx4fKpVsUctiDV9jk"

	addr, err := bscript.NewAddressFromPublicKeyString(pubKey, false)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	if addr.PublicKeyHash != testPublicKeyHash {
		t.Errorf("PKH is incorrect,\ngot: %s\nexpected: %s", addr.PublicKeyHash, testPublicKeyHash)
	}

	if addr.AddressString != expectedAddress {
		t.Errorf("Address is incorrect,\ngot: %s\nexpected: %s", addr.AddressString, expectedAddress)
	}
}

func TestNewFromPublicKey(t *testing.T) {
	pubKeyBytes, err := hex.DecodeString("026cf33373a9f3f6c676b75b543180703df225f7f8edbffedc417718a8ad4e89ce")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	var pubKey *bsvec.PublicKey
	pubKey, err = bsvec.ParsePubKey(pubKeyBytes, bsvec.S256())
	if err != nil {
		t.Error(err)
		return
	}
	expectedAddress := "114ZWApV4EEU8frr7zygqQcB1V2BodGZuS"

	var addr *bscript.Address
	addr, err = bscript.NewAddressFromPublicKey(pubKey, true)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	if addr.PublicKeyHash != testPublicKeyHash {
		t.Errorf("PKH is incorrect,\ngot: %s\nexpected: %s", addr.PublicKeyHash, testPublicKeyHash)
	}

	if addr.AddressString != expectedAddress {
		t.Errorf("Address is incorrect,\ngot: %s\nexpected: %s", addr.AddressString, expectedAddress)
	}
}

func TestBase58EncodeMissingChecksum(t *testing.T) {
	input, err := hex.DecodeString("0488b21e000000000000000000362f7a9030543db8751401c387d6a71e870f1895b3a62569d455e8ee5f5f5e5f03036624c6df96984db6b4e625b6707c017eb0e0d137cd13a0c989bfa77a4473fd")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	res := bscript.Base58EncodeMissingChecksum(input)

	expected := "xpub661MyMwAqRbcF5ivRisXcZTEoy7d9DfLF6fLqpu5GWMfeUyGHuWJHVp5uexDqXTWoySh8pNx3ELW7qymwPNg3UEYHjwh1tpdm3P9J2j4g32"

	if res != expected {
		t.Errorf("Expected %q, got %q", expected, res)
	}
}
