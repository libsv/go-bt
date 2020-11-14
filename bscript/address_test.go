package bscript_test

import (
	"encoding/hex"
	"testing"

	"github.com/bitcoinsv/bsvd/bsvec"
	"github.com/libsv/go-bt/bscript"
	"github.com/stretchr/testify/assert"
)

const testPublicKeyHash = "00ac6144c4db7b5790f343cf0477a65fb8a02eb7"

func TestNewFromStringMainnet(t *testing.T) {
	addressMain := "1E7ucTTWRTahCyViPhxSMor2pj4VGQdFMr"
	expectedPublicKeyhash := "8fe80c75c9560e8b56ed64ea3c26e18d2c52211b"

	addr, err := bscript.NewAddressFromString(addressMain)
	assert.NoError(t, err)
	assert.NotNil(t, addr)

	assert.Equal(t, expectedPublicKeyhash, addr.PublicKeyHash, addressMain)
	assert.Equal(t, addressMain, addr.AddressString)
}

func TestNewFromStringTestnet(t *testing.T) {
	addressTestnet := "mtdruWYVEV1wz5yL7GvpBj4MgifCB7yhPd"
	expectedPublicKeyhash := "8fe80c75c9560e8b56ed64ea3c26e18d2c52211b"

	addr, err := bscript.NewAddressFromString(addressTestnet)
	assert.NoError(t, err)
	assert.NotNil(t, addr)

	assert.Equal(t, expectedPublicKeyhash, addr.PublicKeyHash, addressTestnet)
	assert.Equal(t, addressTestnet, addr.AddressString)
}

func TestNewFromStringShortAddress(t *testing.T) {
	addr, err := bscript.NewAddressFromString("ADD8E55")
	assert.Error(t, err)
	assert.Nil(t, addr)
	assert.EqualError(t, err, "invalid address length for 'ADD8E55'")
}

func TestNewFromStringUnsupportedAddress(t *testing.T) {
	addr, err := bscript.NewAddressFromString("27BvY7rFguYQvEL872Y7Fo77Y3EBApC2EK")
	assert.Error(t, err)
	assert.Nil(t, addr)
	assert.EqualError(t, err, "address 27BvY7rFguYQvEL872Y7Fo77Y3EBApC2EK is not supported")
}

func TestNewFromPublicKeyStringMainnet(t *testing.T) {
	addr, err := bscript.NewAddressFromPublicKeyString(
		"026cf33373a9f3f6c676b75b543180703df225f7f8edbffedc417718a8ad4e89ce",
		true,
	)
	assert.NoError(t, err)
	assert.NotNil(t, addr)

	assert.Equal(t, testPublicKeyHash, addr.PublicKeyHash)
	assert.Equal(t, "114ZWApV4EEU8frr7zygqQcB1V2BodGZuS", addr.AddressString)
}

func TestNewFromPublicKeyStringTestnet(t *testing.T) {
	addr, err := bscript.NewAddressFromPublicKeyString(
		"026cf33373a9f3f6c676b75b543180703df225f7f8edbffedc417718a8ad4e89ce",
		false,
	)
	assert.NoError(t, err)
	assert.NotNil(t, addr)

	assert.Equal(t, testPublicKeyHash, addr.PublicKeyHash)
	assert.Equal(t, "mfaWoDuTsFfiunLTqZx4fKpVsUctiDV9jk", addr.AddressString)
}

func TestNewFromPublicKey(t *testing.T) {
	pubKeyBytes, err := hex.DecodeString("026cf33373a9f3f6c676b75b543180703df225f7f8edbffedc417718a8ad4e89ce")
	assert.NoError(t, err)

	var pubKey *bsvec.PublicKey
	pubKey, err = bsvec.ParsePubKey(pubKeyBytes, bsvec.S256())
	assert.NoError(t, err)
	assert.NotNil(t, pubKey)

	var addr *bscript.Address
	addr, err = bscript.NewAddressFromPublicKey(pubKey, true)
	assert.NoError(t, err)
	assert.NotNil(t, addr)

	assert.Equal(t, testPublicKeyHash, addr.PublicKeyHash)
	assert.Equal(t, "114ZWApV4EEU8frr7zygqQcB1V2BodGZuS", addr.AddressString)
}

func TestBase58EncodeMissingChecksum(t *testing.T) {
	input, err := hex.DecodeString("0488b21e000000000000000000362f7a9030543db8751401c387d6a71e870f1895b3a62569d455e8ee5f5f5e5f03036624c6df96984db6b4e625b6707c017eb0e0d137cd13a0c989bfa77a4473fd")
	assert.NoError(t, err)

	assert.Equal(t,
		"xpub661MyMwAqRbcF5ivRisXcZTEoy7d9DfLF6fLqpu5GWMfeUyGHuWJHVp5uexDqXTWoySh8pNx3ELW7qymwPNg3UEYHjwh1tpdm3P9J2j4g32",
		bscript.Base58EncodeMissingChecksum(input),
	)
}
