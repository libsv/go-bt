package address

import (
	"encoding/hex"
	"testing"

	"github.com/btcsuite/btcd/btcec"
	"github.com/libsv/libsv/address"
)

func TestNewAddressFromStringMainnet(t *testing.T) {
	addressMain := "1E7ucTTWRTahCyViPhxSMor2pj4VGQdFMr"
	expectedPublicKeyhash := "8fe80c75c9560e8b56ed64ea3c26e18d2c52211b"

	addr, err := address.NewAddressFromString(addressMain)
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

func TestNewAddressFromStringTestnet(t *testing.T) {
	addressTestnet := "mtdruWYVEV1wz5yL7GvpBj4MgifCB7yhPd"
	expectedPublicKeyhash := "8fe80c75c9560e8b56ed64ea3c26e18d2c52211b"

	addr, err := address.NewAddressFromString(addressTestnet)
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

func TestNewAddressFromPublicKeyStringMainnet(t *testing.T) {
	pubKey := "026cf33373a9f3f6c676b75b543180703df225f7f8edbffedc417718a8ad4e89ce"
	expectedPublicKeyhash := "00ac6144c4db7b5790f343cf0477a65fb8a02eb7"
	expectedAddress := "114ZWApV4EEU8frr7zygqQcB1V2BodGZuS"

	addr, err := address.NewAddressFromPublicKeyString(pubKey, true)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	if addr.PublicKeyHash != expectedPublicKeyhash {
		t.Errorf("PKH is incorrect,\ngot: %s\nexpected: %s", addr.PublicKeyHash, expectedPublicKeyhash)
	}

	if addr.AddressString != expectedAddress {
		t.Errorf("Address is incorrect,\ngot: %s\nexpected: %s", addr.AddressString, expectedAddress)
	}
}

func TestNewAddressFromPublicKeyStringTestnet(t *testing.T) {
	pubKey := "026cf33373a9f3f6c676b75b543180703df225f7f8edbffedc417718a8ad4e89ce"
	expectedPublicKeyhash := "00ac6144c4db7b5790f343cf0477a65fb8a02eb7"
	expectedAddress := "mfaWoDuTsFfiunLTqZx4fKpVsUctiDV9jk"

	addr, err := address.NewAddressFromPublicKeyString(pubKey, false)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	if addr.PublicKeyHash != expectedPublicKeyhash {
		t.Errorf("PKH is incorrect,\ngot: %s\nexpected: %s", addr.PublicKeyHash, expectedPublicKeyhash)
	}

	if addr.AddressString != expectedAddress {
		t.Errorf("Address is incorrect,\ngot: %s\nexpected: %s", addr.AddressString, expectedAddress)
	}
}

func TestNewAddressFromPublicKey(t *testing.T) {
	pubKeyBytes, err := hex.DecodeString("026cf33373a9f3f6c676b75b543180703df225f7f8edbffedc417718a8ad4e89ce")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	pubKey, err := btcec.ParsePubKey(pubKeyBytes, btcec.S256())
	if err != nil {
		t.Error(err)
		return
	}
	expectedPublicKeyhash := "00ac6144c4db7b5790f343cf0477a65fb8a02eb7"
	expectedAddress := "114ZWApV4EEU8frr7zygqQcB1V2BodGZuS"

	addr, err := address.NewAddressFromPublicKey(pubKey, true)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	if addr.PublicKeyHash != expectedPublicKeyhash {
		t.Errorf("PKH is incorrect,\ngot: %s\nexpected: %s", addr.PublicKeyHash, expectedPublicKeyhash)
	}

	if addr.AddressString != expectedAddress {
		t.Errorf("Address is incorrect,\ngot: %s\nexpected: %s", addr.AddressString, expectedAddress)
	}
}
