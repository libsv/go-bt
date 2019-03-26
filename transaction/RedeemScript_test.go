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
	rs.AddPublicKey("xpub661MyMwAqRbcFvmkwJ82wpjkNmjMWb8n4Pp9Gz3dJjPZMh4uW7z9CpSsTNjEcH3KW5Tibn77qDM9X7gJyjxySEgzBmoQ9LGxSrgHMXTMqx6")
	rs.AddPublicKey("xpub661MyMwAqRbcF5ivRisXcZTEoy7d9DfLF6fLqpu5GWMfeUyGHuWJHVp5uexDqXTWoySh8pNx3ELW7qymwPNg3UEYHjwh1tpdm3P9J2j4g32")

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
	rs.AddPublicKey("xpub661MyMwAqRbcF5ivRisXcZTEoy7d9DfLF6fLqpu5GWMfeUyGHuWJHVp5uexDqXTWoySh8pNx3ELW7qymwPNg3UEYHjwh1tpdm3P9J2j4g32")
	rs.AddPublicKey("xpub661MyMwAqRbcFvmkwJ82wpjkNmjMWb8n4Pp9Gz3dJjPZMh4uW7z9CpSsTNjEcH3KW5Tibn77qDM9X7gJyjxySEgzBmoQ9LGxSrgHMXTMqx6")

	if rs.getAddress() != expected {
		t.Errorf("Expected Address %q, got %q", expected, rs.getAddress())
	}
}
