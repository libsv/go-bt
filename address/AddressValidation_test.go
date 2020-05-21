package address_test

import (
	"testing"

	"github.com/libsv/libsv/address"
)

func TestValidateAddressMainnetP2PKH(t *testing.T) {
	ok, err := address.ValidateAddress("114ZWApV4EEU8frr7zygqQcB1V2BodGZuS")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	if !ok {
		t.Errorf("114ZWApV4EEU8frr7zygqQcB1V2BodGZuS marked as invalid when it should be valid")
	}
}

func TestValidateAddressTestnetP2PKH(t *testing.T) {
	ok, err := address.ValidateAddress("mfaWoDuTsFfiunLTqZx4fKpVsUctiDV9jk")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	if !ok {
		t.Errorf("114ZWApV4EEU8frr7zygqQcB1V2BodGZuS marked as invalid when it should be valid")
	}
}

func TestValidateAddressBIP276(t *testing.T) {
	ok, err := address.ValidateAddress("bitcoin-script:0101522102e5b3f2970648b5592b7303367ab7d7d49e6e27dd80c7b5da18a22dac67a51a322103da6bf6a0c1a06ae7c4091542e0eaa29f2678e7957b78ba09cbe5a36241a4ad0452aeb245ccc7")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	if !ok {
		t.Errorf("114ZWApV4EEU8frr7zygqQcB1V2BodGZuS marked as invalid when it should be valid")
	}
}
