package transaction

import "testing"

func TestXPubToPkey(t *testing.T) {
	rs, err := NewRedeemScript(2)
	if err != nil {
		t.Error(err)
	}

	rs.AddPublicKey("xpub661MyMwAqRbcF5ivRisXcZTEoy7d9DfLF6fLqpu5GWMfeUyGHuWJHVp5uexDqXTWoySh8pNx3ELW7qymwPNg3UEYHjwh1tpdm3P9J2j4g32")

	t.Log(rs.getPublicKeys())
}

func TestGetRedeemScript(t *testing.T) {
	rs, err := NewRedeemScript(2)
	if err != nil {
		t.Error(err)
	}

	rs.AddPublicKey("xpub661MyMwAqRbcFvmkwJ82wpjkNmjMWb8n4Pp9Gz3dJjPZMh4uW7z9CpSsTNjEcH3KW5Tibn77qDM9X7gJyjxySEgzBmoQ9LGxSrgHMXTMqx6")
	//xpub68yCP8WLutn3nn2UQ4UbFK4hBMU7vXGsUR29knoJhJj3bfuDR9Gn3BMwDMQWwgaH9aa4yKjPHEnSwea5mzYvEYsJu3AbAKuC7zdBbhcSsjf'
	//03f2538c34a0991dcbcae32c56c1158822c88a4149e0549363dd9c541a64551145'

	rs.AddPublicKey("xpub661MyMwAqRbcF5ivRisXcZTEoy7d9DfLF6fLqpu5GWMfeUyGHuWJHVp5uexDqXTWoySh8pNx3ELW7qymwPNg3UEYHjwh1tpdm3P9J2j4g32")
	//xpub68Sp48e9T9Yh7iSQSkajRkEoDNdweNDkhUY5DtKHqd3yRF47qqdMiLM57AobrtnARaydYXMAJkWxNMKTaBZwGsPcuyTCgArtrigFDMaNqV9'
	//03d10369cb9603521e3b2b2f13b71d356e9465867c7c79233e58d85f82dec24194'

	// P2SH: 0e95261082d65c384a6106f114474bc0784ba67e
	// Addr: 33284G9cdLGrgDExvFvA2HcFRiupUQjzuR

	t.Logf("PKS: %+x", rs.getPublicKeys())

	t.Logf("ADDRESS: %s", rs.getAddress())

	t.Logf("REDEEM SCRIPT: %x", rs.getRedeemScript())

	t.Logf("REDEEM SCRIPT HASH: %x", rs.getRedeemScriptHash())

	t.Logf("SCRIPT PUB KEY: %x", rs.getScriptPubKey())
}
