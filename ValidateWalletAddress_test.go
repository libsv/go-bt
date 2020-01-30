package cryptolib

import "testing"

func TestValidateLegacyAddress(t *testing.T) {
	ok, err := ValidateWalletAddress("BSV", "1BvBMSEYstWetqTFn5Au4m4GFg7xJaNVN2")
	if err != nil {
		t.Error(err)
	}

	if !ok {
		t.Error("Should have returned true")
	}
}

func TestValidateLegacyAddressWithWrongCoin(t *testing.T) {
	ok, err := ValidateWalletAddress("TSV", "1BvBMSEYstWetqTFn5Au4m4GFg7xJaNVN2")
	if err == nil {
		t.Error(err)
	}

	if ok {
		t.Error("Should have returned false")
	}
}

func TestValidateLegacyTestAddress(t *testing.T) {
	ok, err := ValidateWalletAddress("TSV", "mipcBbFg9gMiCh81Kj8tqqdgoZub1ZJRfn")
	if err != nil {
		t.Error(err)
	}

	if !ok {
		t.Error("Should have returned true")
	}
}
func TestValidateLegacyP2SHAddress(t *testing.T) {
	ok, err := ValidateWalletAddress("BSV", "3EktnHQD7RiAE6uzMj2ZifT9YgRrkSgzQX")
	if err == nil {
		t.Error(err)
	}

	if ok {
		t.Error("Should have returned false")
	}
}

func TestValidateBareMultisigAddress(t *testing.T) {
	ok, err := ValidateWalletAddress("BSV", "bitcoin-script:0101522102e5b3f2970648b5592b7303367ab7d7d49e6e27dd80c7b5da18a22dac67a51a322103da6bf6a0c1a06ae7c4091542e0eaa29f2678e7957b78ba09cbe5a36241a4ad0452aeb245ccc7")
	if err != nil {
		t.Error(err)
	}

	if !ok {
		t.Error("Should have returned true")
	}
}

func TestValidateBareMultisigAddressBTC(t *testing.T) {
	ok, err := ValidateWalletAddress("BTC", "bitcoin-script:0101522102e5b3f2970648b5592b7303367ab7d7d49e6e27dd80c7b5da18a22dac67a51a322103da6bf6a0c1a06ae7c4091542e0eaa29f2678e7957b78ba09cbe5a36241a4ad0452aeb245ccc7")
	if err == nil {
		t.Error(err)
	}

	if ok {
		t.Error("Should have returned false")
	}
}

func TestValidBCHAddress(t *testing.T) {
	ok, err := ValidateWalletAddress("BCH", "bitcoincash:qr6m7j9njldwwzlg9v7v53unlr4jkmx6eylep8ekg2")

	if err != nil {
		t.Error(err)
	}

	if !ok {
		t.Error("Should have returned true")
	}
}

func TestWrongCoin(t *testing.T) {
	ok, _ := ValidateWalletAddress("RSV", "bitcoincash:qr6m7j9njldwwzlg9v7v53unlr4jkmx6eylep8ekg2")

	if ok {
		t.Error("Should have returned false")
	}
}

func TestValidBTCAddress(t *testing.T) {
	ok, err := ValidateWalletAddress("BTC", "19di19ddE1g1wQUhW25MjHybqLXmud8DZj")

	if err != nil {
		t.Error(err)
	}

	if !ok {
		t.Error("Should have returned true")
	}
}
