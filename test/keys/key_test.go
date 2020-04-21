package keys

import (
	"encoding/hex"
	"testing"

	"github.com/libsv/libsv/keys"
)

func TestPrivateKeyToWif1(t *testing.T) {
	seed := "0C28FCA386C7A227600B2FE50B7CAE11EC86D3BF1FBE471BE89827E19D72AA1D"
	hex, err := hex.DecodeString(seed)
	if err != nil {
		t.Error(err)
	}

	wif := keys.PrivateKeyToWIF(hex)
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

	wif := keys.PrivateKeyToWIF(hex)
	if wif != "5KehCbbxxMsPomgbYqJf2VXKtiD8UKVuaHStjaUyRsZ1X2KjmFZ" {
		t.Errorf("PrivateKeyToWIF failed")
	}
}

// func TestValidateExternalMessage(t *testing.T) {
// 	// privBytes := base58.Decode("L4PPagW8MXCuDRdNiuv8aWeftc1cpPfjMRiedqyzeerrCjh51eMR")

// 	// priv, pub := btcec.PrivKeyFromBytes(btcec.S256(), privBytes)

// 	// t.Logf("%v, %v", priv, pub)
// 	message := "Hello world"
// 	address := "1DnmCuL9xMnWBy9rUwWUyz1vi57LMM2AfJ"

// 	sigBytes, err := base64.StdEncoding.DecodeString("IMuSQkuYbEZkrK5x4/7b4SffUsTArsammM7i6kpwn7LiUUWcsTio+QY0qWzQaZ2ujBsKvCVJiWyWkCYUQChuDGw=")
// 	if err != nil {
// 		t.Error(err)
// 		return
// 	}

// 	addressDecoded, netID, err := base58.CheckDecode(address)
// 	if err != nil {
// 		t.Error(err)
// 		return
// 	}
// 	if netID != 0 {
// 		t.Errorf("NetID should be 0")
// 		return
// 	}

// 	pubKey, ok, err := btcec.RecoverCompact(btcec.S256(), sigBytes, addressDecoded)
// 	if err != nil {
// 		t.Error(err)
// 		return
// 	}

// 	t.Log(pubKey, ok)

// 	signature, err := btcec.ParseDERSignature(sigBytes, btcec.S256())
// 	if err != nil {
// 		t.Error(err)
// 		return
// 	}

// 	// messageHash := chainhash.DoubleHashB([]byte(message))
// 	verified := signature.Verify([]byte(message), pubKey)
// 	t.Logf("Signature Verified? %t", verified)

// }
