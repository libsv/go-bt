package transaction

import (
	"testing"
)

func TestGetRedeemScript(t *testing.T) {
	expected := "33284G9cdLGrgDExvFvA2HcFRiupUQjzuR"
	rs, err := NewRedeemScript(2)
	if err != nil {
		t.Error(err)
	}
	rs.AddPublicKey("xpub661MyMwAqRbcFvmkwJ82wpjkNmjMWb8n4Pp9Gz3dJjPZMh4uW7z9CpSsTNjEcH3KW5Tibn77qDM9X7gJyjxySEgzBmoQ9LGxSrgHMXTMqx6", []uint32{0, 0})
	rs.AddPublicKey("xpub661MyMwAqRbcF5ivRisXcZTEoy7d9DfLF6fLqpu5GWMfeUyGHuWJHVp5uexDqXTWoySh8pNx3ELW7qymwPNg3UEYHjwh1tpdm3P9J2j4g32", []uint32{0, 0})

	if rs.getAddress() != expected {
		t.Errorf("Expected %q, got %q", expected, rs.getAddress())
	}
}

func TestGetRedeemScript2(t *testing.T) {
	expected := "33284G9cdLGrgDExvFvA2HcFRiupUQjzuR"
	rs, err := NewRedeemScript(2)
	if err != nil {
		t.Error(err)
	}
	rs.AddPublicKey("xpub661MyMwAqRbcF5ivRisXcZTEoy7d9DfLF6fLqpu5GWMfeUyGHuWJHVp5uexDqXTWoySh8pNx3ELW7qymwPNg3UEYHjwh1tpdm3P9J2j4g32", []uint32{0, 0})
	rs.AddPublicKey("xpub661MyMwAqRbcFvmkwJ82wpjkNmjMWb8n4Pp9Gz3dJjPZMh4uW7z9CpSsTNjEcH3KW5Tibn77qDM9X7gJyjxySEgzBmoQ9LGxSrgHMXTMqx6", []uint32{0, 0})

	if rs.getAddress() != expected {
		t.Errorf("Expected Address %q, got %q", expected, rs.getAddress())
	}

	// t.Logf("%x", rs.getRedeemScript())
	// t.Logf("%+v", rs)
}

func TestGetRedeemScriptFromElectrumRedeemScript(t *testing.T) {
	s := "524c53ff0488b21e000000000000000000362f7a9030543db8751401c387d6a71e870f1895b3a62569d455e8ee5f5f5e5f03036624c6df96984db6b4e625b6707c017eb0e0d137cd13a0c989bfa77a4473fd000000004c53ff0488b21e0000000000000000008b20425398995f3c866ea6ce5c1828a516b007379cf97b136bffbdc86f75df14036454bad23b019eae34f10aff8b8d6d8deb18cb31354e5a169ee09d8a4560e8250000000052ae"
	expected := "33284G9cdLGrgDExvFvA2HcFRiupUQjzuR"

	rs, err := NewRedeemScriptFromElectrum(s)
	if err != nil {
		t.Error(err)
	}

	if rs.getAddress() != expected {
		t.Errorf("Expected Address %q, got %q", expected, rs.getAddress())
	}

	// t.Log(rs.getAddress())

}
