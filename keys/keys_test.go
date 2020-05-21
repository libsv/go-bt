package keys_test

import (
	"encoding/hex"
	"testing"

	"github.com/bitcoinsv/bsvd/bsvec"
	"github.com/bitcoinsv/bsvd/chaincfg"
	"github.com/bitcoinsv/bsvutil"
	"github.com/bitcoinsv/bsvutil/hdkeychain"
)

func TestPrivateKeyToWif1(t *testing.T) {
	seed := "0C28FCA386C7A227600B2FE50B7CAE11EC86D3BF1FBE471BE89827E19D72AA1D"
	h, err := hex.DecodeString(seed)
	if err != nil {
		t.Error(err)
	}

	priv, _ := bsvec.PrivKeyFromBytes(bsvec.S256(), h)

	wif, err := bsvutil.NewWIF(priv, &chaincfg.MainNetParams, false)
	if err != nil {
		t.Fatal(err)
	}

	if wif.String() != "5HueCGU8rMjxEXxiPuD5BDku4MkFqeZyd4dZ1jvhTVqvbTLvyTJ" {
		t.Errorf("PrivateKeyToWIF failed")
	}
}

func TestPrivateKeyToWif2(t *testing.T) {
	seed := "f19c523315891e6e15ae0608a35eec2e00ebd6d1984cf167f46336dabd9b2de4"
	h, err := hex.DecodeString(seed)
	if err != nil {
		t.Error(err)
	}
	priv, _ := bsvec.PrivKeyFromBytes(bsvec.S256(), h)

	wif, err := bsvutil.NewWIF(priv, &chaincfg.MainNetParams, false)
	if err != nil {
		t.Fatal(err)
	}

	if wif.String() != "5KehCbbxxMsPomgbYqJf2VXKtiD8UKVuaHStjaUyRsZ1X2KjmFZ" {
		t.Errorf("PrivateKeyToWIF failed")
	}
}

func TestConvertXPriv(t *testing.T) {
	const xprv = "xprv9s21ZrQH143K2beTKhLXFRWWFwH8jkwUssjk3SVTiApgmge7kNC3jhVc4NgHW8PhW2y7BCDErqnKpKuyQMjqSePPJooPJowAz5BVLThsv6c"
	const expected = "5f86e4023a4e94f00463f81b70ff951f83f896a0a3e6ed89cf163c152f954f8b"

	pk, err := hdkeychain.NewKeyFromString(xprv)
	if err != nil {
		t.Error(err)
	}

	priv, err := pk.ECPrivKey()
	if err != nil {
		t.Error(err)
	}

	res := hex.EncodeToString(priv.Serialize())

	if res != expected {
		t.Errorf("Expected %q, got %q", expected, res)
	}
}

func TestChild(t *testing.T) {
	pk, err := hdkeychain.NewKeyFromString("xpub661MyMwAqRbcFCY8gDwbuGXzaissPEB7QBGdktacSbuwPiXdftuKxCttUPAMxw7ezZ2DEcC9hPcoz5AK3S3corSXx5F1s3MNiqCBbQztkR4")
	if err != nil {
		t.Error(err)
		return
	}

	result, err := pk.Child(0)
	if err != nil {
		t.Error(err)
		return
	}

	result2, err := result.Child(0)
	if err != nil {
		t.Error(err)
		return
	}

	expected := "02a07294fdca4c963bf61351509519b270a209cfc11fc12e95c1992ddf0c6faa59"
	publicKey, err := result2.ECPubKey()
	publicKeyStr := hex.EncodeToString(publicKey.SerializeCompressed())

	if publicKeyStr != expected {
		t.Errorf("Expected %q, got %q", publicKey, expected)
	}
}
