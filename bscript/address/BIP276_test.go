package address_test

import (
	"testing"

	"github.com/libsv/go-bt/script/address"
)

func TestEncode(t *testing.T) {
	fakeScript := "fake script"
	fakeScriptBytes := []byte(fakeScript)

	s := address.EncodeBIP276(address.PrefixScript, address.NetworkMainnet, address.CurrentVersion, fakeScriptBytes)

	expected := "bitcoin-script:010166616b65207363726970746f0cd86a"

	if s != expected {
		t.Errorf("Expected: %q, got: %q", expected, s)
	}

}

func TestDecode(t *testing.T) {
	testScript := "bitcoin-script:010166616b65207363726970746f0cd86a"

	prefix, network, version, data, err := address.DecodeBIP276(testScript)
	if err != nil {
		t.Errorf("Error decoding bip276 string: %+v", err)
	} else {
		t.Logf("prefix: %q, network: %d, version: %d, data,: %s", prefix, network, version, data)

	}

}
