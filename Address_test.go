package cryptolib

import "testing"

func TestAddressToPubKeyHash(t *testing.T) {
	publicKeyhash := "8fe80c75c9560e8b56ed64ea3c26e18d2c52211b"

	addressTestnet := "mtdruWYVEV1wz5yL7GvpBj4MgifCB7yhPd"
	expectedPublicKeyhashTestnet, err := AddressToPubKeyHash(addressTestnet)

	addressLive := "1E7ucTTWRTahCyViPhxSMor2pj4VGQdFMr"
	expectedPublicKeyhashLivenet, err := AddressToPubKeyHash(addressLive)

	if err != nil {
		t.Error(err)
	}

	if publicKeyhash != expectedPublicKeyhashTestnet {
		t.Errorf("PKH from testnet address incorrect,\ngot: %s\nexpected: %s", publicKeyhash, expectedPublicKeyhashTestnet)
	}

	if publicKeyhash != expectedPublicKeyhashLivenet {
		t.Errorf("PKH from testnet address incorrect,\ngot: %s\nexpected: %s", publicKeyhash, expectedPublicKeyhashTestnet)
	}
}
