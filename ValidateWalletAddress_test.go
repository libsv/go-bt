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

func TestValidBSVAddress(t *testing.T) {
	ok, err := ValidateWalletAddress("BSV", "bitcoincash:qr6m7j9njldwwzlg9v7v53unlr4jkmx6eylep8ekg2")

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
