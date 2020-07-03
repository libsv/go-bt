package script_test

import (
	"encoding/hex"
	"testing"

	"github.com/libsv/libsv/script"
)

func TestDecodePartsSimple(t *testing.T) {
	s := "05000102030401FF02ABCD"
	parts, err := script.DecodeStringParts(s)
	if err != nil {
		t.Error(err)
	}

	if len(parts) != 3 {
		t.Errorf("Expected 3 parts, got %d", len(parts))
	}
}

func TestDecodePartsSimpleAndEncode(t *testing.T) {
	s := "05000102030401FF02ABCD"
	parts, err := script.DecodeStringParts(s)
	if err != nil {
		t.Error(err)
	}

	if len(parts) != 3 {
		t.Errorf("Expected 3 parts, got %d", len(parts))
	}

	p, err := script.EncodeParts(parts)
	if err != nil {
		t.Error(err)
	}

	h := hex.EncodeToString(p)

	expected := "05000102030401ff02abcd"
	if h != expected {
		t.Errorf("Expected %q, got %q", expected, h)
	}
}

func TestDecodePartsEmpty(t *testing.T) {
	s := ""
	parts, err := script.DecodeStringParts(s)
	if err != nil {
		t.Error(err)
	}

	if len(parts) != 0 {
		t.Errorf("Expected [], got %+v", parts)
	}
}

func TestDecodePartsComplex(t *testing.T) {
	s := "524c53ff0488b21e000000000000000000362f7a9030543db8751401c387d6a71e870f1895b3a62569d455e8ee5f5f5e5f03036624c6df96984db6b4e625b6707c017eb0e0d137cd13a0c989bfa77a4473fd000000004c53ff0488b21e0000000000000000008b20425398995f3c866ea6ce5c1828a516b007379cf97b136bffbdc86f75df14036454bad23b019eae34f10aff8b8d6d8deb18cb31354e5a169ee09d8a4560e8250000000052ae"
	parts, err := script.DecodeStringParts(s)
	if err != nil {
		t.Error(err)
	}

	if len(parts) != 5 {
		t.Errorf("Expected 5 parts, got %d", len(parts))
	}
}

func TestDecodePartsBad(t *testing.T) {
	s := "05000000"
	_, err := script.DecodeStringParts(s)
	if err == nil {
		t.Errorf("Expected an error")
	}

	expected := "Not enough data"
	if err.Error() != expected {
		t.Errorf("Expected %q, got %q", expected, err.Error())
	}
}

func TestDecodePartsBad2(t *testing.T) {
	s := "4c05000000"
	_, err := script.DecodeStringParts(s)
	if err == nil {
		t.Errorf("Expected an error")
	}

	expected := "Not enough data"
	if err.Error() != expected {
		t.Errorf("Expected %q, got %q", expected, err.Error())
	}
}

func TestDecodePartsPanic(t *testing.T) {
	s := "006a046d657461226e3465394d57576a416f576b727646344674724e783252507533584d53344d786570201ed64f8e4ddb6843121dc11e1db6d07c62e59c621f047e1be0a9dd910ca606d04cfe080000000b00045479706503070006706f7374616c000355736503070004686f6d650006526567696f6e030700057374617465000a506f7374616c436f64650307000432383238000b44617465437265617465640d070018323032302d30362d32325431323a32343a32362e3337315a00035f69640307002f302e34623836326165372d323533352d346136312d386461322d3962616231633336353038312e302e342e31332e30000443697479030700046369747900054c696e65300307000474657374000b436f756e747279436f646503070002414500054c696e653103070005746573743200084469737472696374030700086469737472696374"
	parts, err := script.DecodeStringParts(s)
	if err != nil {
		t.Error(err)
	}
	t.Logf("%+v", parts)
}
