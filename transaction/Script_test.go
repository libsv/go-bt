package transaction

import (
	"encoding/hex"
	"testing"
)

func TestIsPublicKeyHashOut(t *testing.T) {
	b, _ := hex.DecodeString("76a91403ececf2d12a7f614aef4c82ecf13c303bd9975d88ac")
	scriptPub := NewScript(b)

	ret := scriptPub.IsPublicKeyHashOut()

	t.Log(ret)
}

func TestIsPublicKeyOut(t *testing.T) {
	b, _ := hex.DecodeString("2102f0d97c290e79bf2a8660c406aa56b6f189ff79f2245cc5aff82808b58131b4d5ac")
	scriptPub := NewScript(b)

	ret := scriptPub.IsPublicKeyOut()

	t.Log(ret)
}

func TestIsScriptHashOut(t *testing.T) {
	b, _ := hex.DecodeString("a9149de5aeaff9c48431ba4dd6e8af73d51f38e451cb87")
	scriptPub := NewScript(b)

	ret := scriptPub.IsScriptHashOut()

	t.Log(ret)
}

func TestIsMultisigOut(t *testing.T) {
	b, _ := hex.DecodeString("5201110122013353ae")
	scriptPub := NewScript(b)

	ret := scriptPub.IsMultisigOut()

	t.Log(ret)
}
