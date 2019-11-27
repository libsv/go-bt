package cryptolib

import (
	"testing"
)

func TestAddressToPubKeyHash(t *testing.T) {
	publicKeyhash := "8fe80c75c9560e8b56ed64ea3c26e18d2c52211b"

	addressTestnet := "mtdruWYVEV1wz5yL7GvpBj4MgifCB7yhPd"
	addr, err := NewAddressFromString(addressTestnet)
	expectedPublicKeyhashTestnet := addr.PublicKeyHash

	addressLive := "1E7ucTTWRTahCyViPhxSMor2pj4VGQdFMr"
	addr2, err := NewAddressFromString(addressLive)
	expectedPublicKeyhashLivenet := addr2.PublicKeyHash

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

func TestPublicKeyToAddress(t *testing.T) {
	publicKey := "0285e9737a74c30a873f74df05124f2aa6f53042c2fc0a130d6cbd7d16b944b004"
	// pkh := "9cf8b938ce2b14f68c59d7ed166b2ae242198037"

	expectedAddressTestnet := "mupwfbLpEposb7h4E8WyxCWbt5UYMHcV27"
	addr, err := NewAddressFromPublicKey(publicKey, false)
	addressTestnet := addr.AddressString

	expectedAddressLive := "1FJzNYFqRoNcp1DSWZYc8HJH25sqPgmyw3"
	addr2, err := NewAddressFromPublicKey(publicKey, true)
	addressLive := addr2.AddressString

	if err != nil {
		t.Error(err)
	}

	if addressTestnet != expectedAddressTestnet {
		t.Errorf("Address Testnet from public key address %s incorrect,\ngot: %s\nexpected: %s", publicKey, addressTestnet, expectedAddressTestnet)
	}

	if addressLive != expectedAddressLive {
		t.Errorf("Address Live from public key  %s incorrect,\ngot: %s\nexpected: %s", publicKey, addressLive, expectedAddressLive)
	}
}
