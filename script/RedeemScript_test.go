package script_test

import (
	"bytes"
	"encoding/hex"

	"github.com/bitcoinsv/bsvd/bsvec"
	"github.com/bitcoinsv/bsvd/chaincfg/chainhash"
	"github.com/libsv/libsv/script"

	"testing"

	"github.com/libsv/libsv/crypto"
)

func TestGetRedeemScript(t *testing.T) {
	expected := "33284G9cdLGrgDExvFvA2HcFRiupUQjzuR"
	rs, err := script.NewRedeemScript(2)
	if err != nil {
		t.Error(err)
	}
	err = rs.AddPublicKey("xpub661MyMwAqRbcFvmkwJ82wpjkNmjMWb8n4Pp9Gz3dJjPZMh4uW7z9CpSsTNjEcH3KW5Tibn77qDM9X7gJyjxySEgzBmoQ9LGxSrgHMXTMqx6", []uint32{0, 0})
	if err != nil {
		t.Error(err)
	}
	err = rs.AddPublicKey("xpub661MyMwAqRbcF5ivRisXcZTEoy7d9DfLF6fLqpu5GWMfeUyGHuWJHVp5uexDqXTWoySh8pNx3ELW7qymwPNg3UEYHjwh1tpdm3P9J2j4g32", []uint32{0, 0})
	if err != nil {
		t.Error(err)
	}

	if rs.GetAddress() != expected {
		t.Errorf("Expected %q, got %q", expected, rs.GetAddress())
	}
}

func TestHash160(t *testing.T) {
	input, _ := hex.DecodeString("522103d10369cb9603521e3b2b2f13b71d356e9465867c7c79233e58d85f82dec241942103f2538c34a0991dcbcae32c56c1158822c88a4149e0549363dd9c541a6455114552ae")
	expected := "0e95261082d65c384a6106f114474bc0784ba67e"
	result := crypto.Hash160(input)

	if hex.EncodeToString(result) != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestGetRedeemScript2(t *testing.T) {
	expected := "33284G9cdLGrgDExvFvA2HcFRiupUQjzuR"
	rs, err := script.NewRedeemScript(2)
	if err != nil {
		t.Error(err)
	}
	err = rs.AddPublicKey("xpub661MyMwAqRbcF5ivRisXcZTEoy7d9DfLF6fLqpu5GWMfeUyGHuWJHVp5uexDqXTWoySh8pNx3ELW7qymwPNg3UEYHjwh1tpdm3P9J2j4g32", []uint32{0, 0})
	if err != nil {
		t.Error(err)
	}
	err = rs.AddPublicKey("xpub661MyMwAqRbcFvmkwJ82wpjkNmjMWb8n4Pp9Gz3dJjPZMh4uW7z9CpSsTNjEcH3KW5Tibn77qDM9X7gJyjxySEgzBmoQ9LGxSrgHMXTMqx6", []uint32{0, 0})
	if err != nil {
		t.Error(err)
	}

	if rs.GetAddress() != expected {
		t.Errorf("Expected Address %q, got %q", expected, rs.GetAddress())
	}
}

func TestGetRedeemScriptFromElectrumRedeemScript(t *testing.T) {
	s := "524c53ff0488b21e000000000000000000362f7a9030543db8751401c387d6a71e870f1895b3a62569d455e8ee5f5f5e5f03036624c6df96984db6b4e625b6707c017eb0e0d137cd13a0c989bfa77a4473fd000000004c53ff0488b21e0000000000000000008b20425398995f3c866ea6ce5c1828a516b007379cf97b136bffbdc86f75df14036454bad23b019eae34f10aff8b8d6d8deb18cb31354e5a169ee09d8a4560e8250000000052ae"
	expected := "33284G9cdLGrgDExvFvA2HcFRiupUQjzuR"

	rs, err := script.NewRedeemScriptFromElectrum(s)
	if err != nil {
		t.Error(err)
	}

	if rs.GetAddress() != expected {
		t.Errorf("Expected Address %q, got %q", expected, rs.GetAddress())
	}
}

func TestSignRedeemScript(t *testing.T) {
	var redeemScript, _ = hex.DecodeString("524c53ff0488b21e000000000000000000362f7a9030543db8751401c387d6a71e870f1895b3a62569d455e8ee5f5f5e5f03036624c6df96984db6b4e625b6707c017eb0e0d137cd13a0c989bfa77a4473fd000000004c53ff0488b21e0000000000000000008b20425398995f3c866ea6ce5c1828a516b007379cf97b136bffbdc86f75df14036454bad23b019eae34f10aff8b8d6d8deb18cb31354e5a169ee09d8a4560e8250000000052ae")
	var expectedSignature, _ = hex.DecodeString("304402206d4db58f03ba3875a0f442b3f27b9035d281f851abb24d99ecd25ca6b4c528f30220465169db20a1f52345af3a7dda0a7aefa9415d5dc1403435bf08d4d180b7bc01")

	const privHex = "5f86e4023a4e94f00463f81b70ff951f83f896a0a3e6ed89cf163c152f954f8b"

	pkBytes, err := hex.DecodeString(privHex)
	if err != nil {
		t.Error(err)
		return
	}
	privKey, pubKey := bsvec.PrivKeyFromBytes(bsvec.S256(), pkBytes)

	// Sign a message using the private key.
	messageHash := chainhash.DoubleHashB(redeemScript)
	signature, err := privKey.Sign(messageHash)
	if err != nil {
		t.Error(err)
		return
	}

	// Serialize and display the signature.
	serializedSignature := signature.Serialize()

	res := bytes.Compare(serializedSignature, expectedSignature)

	if res != 0 {
		t.Errorf("expected err to be %v, but got %v", expectedSignature, serializedSignature)
	}

	// Verify the signature for the message using the public key.
	verified := signature.Verify(messageHash, pubKey)
	if !verified {
		t.Error("Signature is not verified")
	}

}
