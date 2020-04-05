package crypto

import (
	"github.com/jadwahab/libsv/crypto"
	"testing"
)

func TestNewPublicKey(t *testing.T) {
	p, err := crypto.NewPublicKey("xpub661MyMwAqRbcFCY8gDwbuGXzaissPEB7QBGdktacSbuwPiXdftuKxCttUPAMxw7ezZ2DEcC9hPcoz5AK3S3corSXx5F1s3MNiqCBbQztkR4")
	if err != nil {
		t.Error(err)
		return
	}

	expected := "xpub661MyMwAqRbcFCY8gDwbuGXzaissPEB7QBGdktacSbuwPiXdftuKxCttUPAMxw7ezZ2DEcC9hPcoz5AK3S3corSXx5F1s3MNiqCBbQztkR4"

	if p.XPubKey != expected {
		t.Errorf("Expected %q, got %q", expected, p.XPubKey)
	}

	// t.Log(p.String())

	expected = `{"network":"livenet","depth":0,"fingerPrint":"aa29b7c4","parentFingerPrint":"00000000","childIndex":"AAAAAA==","chainCode":"41fc504936a63056da1a0f9dd44cad3651b64a17b53e523e18a8d228a489c16a","privateKey":"","publicKey":"0362e448fdb4c7c307a80cc3c8ede19cd2599a5ea5c05b188fc56a25c59bfcf125","xprvkey":"","xpubkey":"xpub661MyMwAqRbcFCY8gDwbuGXzaissPEB7QBGdktacSbuwPiXdftuKxCttUPAMxw7ezZ2DEcC9hPcoz5AK3S3corSXx5F1s3MNiqCBbQztkR4"}`
	if p.String() != expected {
		t.Errorf("Expected %q, got %q", expected, p.String())
	}
}

func TestChild(t *testing.T) {
	p, err := crypto.NewPublicKey("xpub661MyMwAqRbcFCY8gDwbuGXzaissPEB7QBGdktacSbuwPiXdftuKxCttUPAMxw7ezZ2DEcC9hPcoz5AK3S3corSXx5F1s3MNiqCBbQztkR4")
	if err != nil {
		t.Error(err)
		return
	}

	result, err := p.Child(0)
	if err != nil {
		t.Error(err)
		return
	}

	result2, err := result.Child(0)
	if err != nil {
		t.Error(err)
		return
	}

	fingerPrint := "ea48e9f4"

	if result2.FingerPrintStr != fingerPrint {
		t.Errorf("Expected %q, got %q", fingerPrint, result2.FingerPrintStr)
	}

	chainCode := "55c16c7330e15bb3fcb55f66faa29242bb9847d14768a92b180e7c8b2536a429"
	if result2.ChainCodeStr != chainCode {
		t.Errorf("Expected %q, got %q", chainCode, result2.ChainCodeStr)
	}

	publicKey := "02a07294fdca4c963bf61351509519b270a209cfc11fc12e95c1992ddf0c6faa59"
	if result2.PublicKeyStr != publicKey {
		t.Errorf("Expected %q, got %q", publicKey, result2.PublicKeyStr)
	}

	t.Log(result2.XPubKey)
}

func TestAddress(t *testing.T) {
	p, err := crypto.NewPublicKey("xpub661MyMwAqRbcFCY8gDwbuGXzaissPEB7QBGdktacSbuwPiXdftuKxCttUPAMxw7ezZ2DEcC9hPcoz5AK3S3corSXx5F1s3MNiqCBbQztkR4")
	if err != nil {
		t.Error(err)
		return
	}

	result, _ := p.Address()
	expected := "1GWjrDd1SyJsUJfjAjUYnWeNuEdcd6UU9P"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestGetXPub(t *testing.T) {
	expected := "xpub661MyMwAqRbcFCY8gDwbuGXzaissPEB7QBGdktacSbuwPiXdftuKxCttUPAMxw7ezZ2DEcC9hPcoz5AK3S3corSXx5F1s3MNiqCBbQztkR4"
	p, err := crypto.NewPublicKey(expected)
	if err != nil {
		t.Error(err)
		return
	}

	result := p.GetXPub()
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}
