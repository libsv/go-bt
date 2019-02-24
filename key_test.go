package cryptolib

import (
	"crypto/ecdsa"
	"encoding/hex"
	"math/big"
	"testing"
)

func TestPrivateKeyToWif1(t *testing.T) {
	seed := "0C28FCA386C7A227600B2FE50B7CAE11EC86D3BF1FBE471BE89827E19D72AA1D"
	hex, err := hex.DecodeString(seed)
	if err != nil {
		t.Error(err)
	}

	wif := PrivateKeyToWIF(hex)
	if wif != "5HueCGU8rMjxEXxiPuD5BDku4MkFqeZyd4dZ1jvhTVqvbTLvyTJ" {
		t.Errorf("PrivateKeyToWIF failed")
	}
}

func TestPrivateKeyToWif2(t *testing.T) {
	seed := "f19c523315891e6e15ae0608a35eec2e00ebd6d1984cf167f46336dabd9b2de4"
	hex, err := hex.DecodeString(seed)
	if err != nil {
		t.Error(err)
	}

	wif := PrivateKeyToWIF(hex)
	if wif != "5KehCbbxxMsPomgbYqJf2VXKtiD8UKVuaHStjaUyRsZ1X2KjmFZ" {
		t.Errorf("PrivateKeyToWIF failed")
	}

	var priv *ecdsa.PrivateKey

	k := new(big.Int)
	k.SetString(seed, 16)

	priv = new(ecdsa.PrivateKey)
	curve := new(KoblitzCurve)
	priv.PublicKey.Curve = curve
	priv.D = k

	priv.PublicKey.X, priv.PublicKey.Y = curve.ScalarBaseMult(k.Bytes())
	t.Logf("%x", priv.Public())

	// pubkeyBytes, _ := x509.MarshalPKIXPublicKey(&priv.PublicKey)

	// t.Logf("%+v", len(priv.PublicKey.X)+len(priv.PublicKey.Y))
	t.Fail()
}
