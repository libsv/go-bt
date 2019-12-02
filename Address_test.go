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
		t.Errorf("PKH from testnet address %s incorrect,\ngot: %s\nexpected: %s", addressTestnet, publicKeyhash, expectedPublicKeyhashTestnet)
	}

	if publicKeyhash != expectedPublicKeyhashLivenet {
		t.Errorf("PKH from Live address %s incorrect,\ngot: %s\nexpected: %s", addressLive, publicKeyhash, expectedPublicKeyhashTestnet)
	}
}

func TestPublicKeyHashFromPublicKeyStr(t *testing.T) {
	pubKey := "03630019a270db9f09ba635bccee980a0b96e19d89533c6a9be26e5f6282ccc47a"
	expectedPublicKeyhash := "05a23cf9b42a0ccb5cf2b2bcb70bd3ac0d2c9852"

	publicKeyHash, err := PublicKeyHashFromPublicKeyStr(pubKey)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	if expectedPublicKeyhash != publicKeyHash {
		t.Logf("Expected %q, got %q", expectedPublicKeyhash, publicKeyHash)
		t.FailNow()
	}
}
