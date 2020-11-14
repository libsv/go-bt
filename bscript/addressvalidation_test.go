package bscript_test

import (
	"testing"

	"github.com/libsv/go-bt/bscript"
	"github.com/stretchr/testify/assert"
)

func TestValidateAddressMainnetP2PKH(t *testing.T) {
	ok, err := bscript.ValidateAddress("114ZWApV4EEU8frr7zygqQcB1V2BodGZuS")
	assert.NoError(t, err)
	assert.Equal(t, true, ok)
}

func TestValidateAddressTestnetP2PKH(t *testing.T) {
	ok, err := bscript.ValidateAddress("mfaWoDuTsFfiunLTqZx4fKpVsUctiDV9jk")
	assert.NoError(t, err)
	assert.Equal(t, true, ok)
}

func TestValidateAddressBIP276(t *testing.T) {
	ok, err := bscript.ValidateAddress("bitcoin-script:0101522102e5b3f2970648b5592b7303367ab7d7d49e6e27dd80c7b5da18a22dac67a51a322103da6bf6a0c1a06ae7c4091542e0eaa29f2678e7957b78ba09cbe5a36241a4ad0452aeb245ccc7")
	assert.NoError(t, err)
	assert.Equal(t, true, ok)
}
