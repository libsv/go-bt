package address

import (
	"encoding/hex"
	"testing"

	"github.com/libsv/libsv/utils"
)

func TestXPubEncode(t *testing.T) {
	input, _ := hex.DecodeString("0488b21e000000000000000000362f7a9030543db8751401c387d6a71e870f1895b3a62569d455e8ee5f5f5e5f03036624c6df96984db6b4e625b6707c017eb0e0d137cd13a0c989bfa77a4473fd")
	res2 := utils.Base58Encode(input)
	t.Log(res2)
	res3 := EncodeToString(input)
	t.Log(res3)

	res3e, _ := DecodeString(res3)
	t.Log(res3e)
	res2e := utils.Base58Decode(res2)
	t.Log(res2e)

	res00 := Base58EncodeMissingChecksum(input)
	t.Log(res00)
}

func TestBase58(t *testing.T) {
	addr := "1E7ucTTWRTahCyViPhxSMor2pj4VGQdFMr"
	in, _ := DecodeString(addr)
	t.Log(hex.EncodeToString(in))
	out := utils.Base58Decode(addr)
	t.Log(hex.EncodeToString(out))

	in2 := EncodeToString(in)
	t.Log(in2)
	out2 := utils.Base58Encode(out)
	t.Log(out2)

	res00 := Base58EncodeMissingChecksum(in)
	t.Log(res00)
}
