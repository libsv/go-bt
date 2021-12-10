package bscript_test

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"strings"
	"testing"

	"github.com/libsv/go-bk/bec"
	"github.com/libsv/go-bk/bip32"
	"github.com/libsv/go-bk/chaincfg"
	"github.com/stretchr/testify/assert"

	"github.com/libsv/go-bt/v2/bscript"
)

func TestNewP2PKHFromPubKeyStr(t *testing.T) {
	t.Parallel()

	scriptP2PKH, err := bscript.NewP2PKHFromPubKeyStr(
		"023717efaec6761e457f55c8417815505b695209d0bbfed8c3265be425b373c2d6",
	)
	assert.NoError(t, err)
	assert.NotNil(t, scriptP2PKH)
	assert.Equal(t,
		"76a9144d5d1920331b71735a97a606d9734aed83cb3dfa88ac",
		hex.EncodeToString(*scriptP2PKH),
	)
}

func TestNewP2PKHFromPubKey(t *testing.T) {
	t.Parallel()

	pk, _ := hex.DecodeString("023717efaec6761e457f55c8417815505b695209d0bbfed8c3265be425b373c2d6")

	pubkey, err := bec.ParsePubKey(pk, bec.S256())
	assert.NoError(t, err)

	scriptP2PKH, err := bscript.NewP2PKHFromPubKeyEC(pubkey)
	assert.NoError(t, err)
	assert.NotNil(t, scriptP2PKH)
	assert.Equal(t,
		"76a9144d5d1920331b71735a97a606d9734aed83cb3dfa88ac",
		hex.EncodeToString(*scriptP2PKH),
	)
}

func TestNewP2PKHFromBip32ExtKey(t *testing.T) {
	t.Parallel()

	t.Run("output is added", func(t *testing.T) {
		var b [64]byte
		_, err := rand.Read(b[:])
		assert.NoError(t, err)

		key, err := bip32.NewMaster(b[:], &chaincfg.TestNet)
		assert.NoError(t, err)

		script, derivationPath, err := bscript.NewP2PKHFromBip32ExtKey(key)

		assert.NoError(t, err)
		assert.NotEmpty(t, derivationPath)
		assert.NotNil(t, script)
		assert.True(t, script.IsP2PKH())
	})

	t.Run("invalid key errors", func(t *testing.T) {
		var b [64]byte
		_, err := rand.Read(b[:])
		assert.NoError(t, err)

		script, derivationPath, err := bscript.NewP2PKHFromBip32ExtKey(&bip32.ExtendedKey{})

		assert.Error(t, err)
		assert.Empty(t, derivationPath)
		assert.Nil(t, script)
	})
}

func TestNewFromHexString(t *testing.T) {
	t.Parallel()

	s, err := bscript.NewFromHexString("76a914e2a623699e81b291c0327f408fea765d534baa2a88ac")
	assert.NoError(t, err)
	assert.NotNil(t, s)
	assert.Equal(t,
		"76a914e2a623699e81b291c0327f408fea765d534baa2a88ac",
		hex.EncodeToString(*s),
	)
}

func TestScript_ToASM(t *testing.T) {
	t.Parallel()

	s, err := bscript.NewFromHexString("76a914e2a623699e81b291c0327f408fea765d534baa2a88ac")
	assert.NoError(t, err)
	assert.NotNil(t, s)

	var res string
	res, err = s.ToASM()
	assert.NoError(t, err)
	assert.Equal(t,
		"OP_DUP OP_HASH160 e2a623699e81b291c0327f408fea765d534baa2a OP_EQUALVERIFY OP_CHECKSIG",
		res,
	)
}

func TestNewFromASM(t *testing.T) {
	t.Parallel()

	s, err := bscript.NewFromASM("OP_DUP OP_HASH160 e2a623699e81b291c0327f408fea765d534baa2a OP_EQUALVERIFY OP_CHECKSIG")
	assert.NoError(t, err)
	assert.NotNil(t, s)
	assert.Equal(t,
		"76a914e2a623699e81b291c0327f408fea765d534baa2a88ac",
		hex.EncodeToString(*s),
	)
}

func TestScript_IsP2PKH(t *testing.T) {
	t.Parallel()

	b, err := hex.DecodeString("76a91403ececf2d12a7f614aef4c82ecf13c303bd9975d88ac")
	assert.NoError(t, err)

	scriptPub := bscript.NewFromBytes(b)
	assert.NotNil(t, scriptPub)
	assert.Equal(t, true, scriptPub.IsP2PKH())
}

func TestScript_IsP2PK(t *testing.T) {
	t.Parallel()

	b, err := hex.DecodeString("2102f0d97c290e79bf2a8660c406aa56b6f189ff79f2245cc5aff82808b58131b4d5ac")
	assert.NoError(t, err)

	scriptPub := bscript.NewFromBytes(b)
	assert.NotNil(t, scriptPub)
	assert.Equal(t, true, scriptPub.IsP2PK())
}

func TestScript_IsP2SH(t *testing.T) {
	t.Parallel()

	b, err := hex.DecodeString("a9149de5aeaff9c48431ba4dd6e8af73d51f38e451cb87")
	assert.NoError(t, err)

	scriptPub := bscript.NewFromBytes(b)
	assert.NotNil(t, scriptPub)
	assert.Equal(t, true, scriptPub.IsP2SH())
}

func TestScript_IsData(t *testing.T) {
	t.Parallel()

	b, err := hex.DecodeString("006a04ac1eed884d53027b2276657273696f6e223a22302e31222c22686569676874223a3634323436302c22707265764d696e65724964223a22303365393264336535633366376264393435646662663438653761393933393362316266623366313166333830616533306432383665376666326165633561323730222c22707265764d696e65724964536967223a2233303435303232313030643736333630653464323133333163613836663031386330343665353763393338663139373735303734373333333533363062653337303438636165316166333032323030626536363034353430323162663934363465393966356139353831613938633963663439353430373539386335396234373334623266646234383262663937222c226d696e65724964223a22303365393264336535633366376264393435646662663438653761393933393362316266623366313166333830616533306432383665376666326165633561323730222c2276637478223a7b2274784964223a2235373962343335393235613930656533396133376265336230306239303631653734633330633832343133663664306132303938653162656137613235313566222c22766f7574223a307d2c226d696e6572436f6e74616374223a7b22656d61696c223a22696e666f407461616c2e636f6d222c226e616d65223a225441414c20446973747269627574656420496e666f726d6174696f6e20546563686e6f6c6f67696573222c226d65726368616e74415049456e64506f696e74223a2268747470733a2f2f6d65726368616e746170692e7461616c2e636f6d2f227d7d46304402206fd1c6d6dd32cc85ddd2f30bc068445dd901c6bd85e394e45bb254716d2bb228022041f0f8b1b33c2e3702aee4ad47155548045ed945738b43dc0faed2e86faa12e4")
	assert.NoError(t, err)

	scriptPub := bscript.NewFromBytes(b)
	assert.NotNil(t, scriptPub)
	assert.Equal(t, true, scriptPub.IsData())
}

func TestScript_IsMultisigOut(t *testing.T) { // TODO: check this
	t.Parallel()

	b, err := hex.DecodeString("5201110122013353ae")
	assert.NoError(t, err)

	scriptPub := bscript.NewFromBytes(b)
	assert.NotNil(t, scriptPub)
	assert.Equal(t, true, scriptPub.IsMultiSigOut())
}

func TestScript_PublicKeyHash(t *testing.T) {
	t.Parallel()

	t.Run("get as bytes", func(t *testing.T) {
		b, err := hex.DecodeString("76a91404d03f746652cfcb6cb55119ab473a045137d26588ac")
		assert.NoError(t, err)

		s := bscript.NewFromBytes(b)
		assert.NotNil(t, s)

		var pkh []byte
		pkh, err = s.PublicKeyHash()
		assert.NoError(t, err)
		assert.Equal(t, "04d03f746652cfcb6cb55119ab473a045137d265", hex.EncodeToString(pkh))
	})

	t.Run("get as string", func(t *testing.T) {
		s, err := bscript.NewFromHexString("76a91404d03f746652cfcb6cb55119ab473a045137d26588ac")
		assert.NoError(t, err)
		assert.NotNil(t, s)

		var pkh []byte
		pkh, err = s.PublicKeyHash()
		assert.NoError(t, err)
		assert.Equal(t, "04d03f746652cfcb6cb55119ab473a045137d265", hex.EncodeToString(pkh))
	})

	t.Run("empty script", func(t *testing.T) {
		s := &bscript.Script{}

		_, err := s.PublicKeyHash()
		assert.Error(t, err)
		assert.EqualError(t, err, "script is empty")
	})
}

func TestErrorIsAppended(t *testing.T) {
	script, _ := hex.DecodeString("6a0548656c6c6f0548656c6c")
	s := bscript.Script(script)

	asm, err := s.ToASM()
	assert.NoError(t, err)
	assert.True(t, strings.HasSuffix(asm, "[error]"), "toASM() should end with [error]")
}

func TestScript_Equals(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		script1 *bscript.Script
		script2 *bscript.Script
		exp     bool
	}{
		"P2PKH scripts that equal should return true": {
			script1: func() *bscript.Script {
				s, err := bscript.NewP2PKHFromAddress("n2wmGVP89x3DsLNqk3NvctfQy9m9pvt7mk")
				assert.NoError(t, err)
				return s
			}(),
			script2: func() *bscript.Script {
				s, err := bscript.NewP2PKHFromAddress("n2wmGVP89x3DsLNqk3NvctfQy9m9pvt7mk")
				assert.NoError(t, err)
				return s
			}(),
			exp: true,
		}, "scripts from bytes equal should return true": {
			script1: func() *bscript.Script {
				b, err := hex.DecodeString("5201110122013353ae")
				assert.NoError(t, err)

				return bscript.NewFromBytes(b)
			}(),
			script2: func() *bscript.Script {
				b, err := hex.DecodeString("5201110122013353ae")
				assert.NoError(t, err)

				return bscript.NewFromBytes(b)
			}(),
			exp: true,
		}, "scripts from hex, equal should return true": {
			script1: func() *bscript.Script {
				s, err := bscript.NewFromHexString("76a91404d03f746652cfcb6cb55119ab473a045137d26588ac")
				assert.NoError(t, err)
				assert.NotNil(t, s)
				return s
			}(),
			script2: func() *bscript.Script {
				s, err := bscript.NewFromHexString("76a91404d03f746652cfcb6cb55119ab473a045137d26588ac")
				assert.NoError(t, err)
				assert.NotNil(t, s)
				return s
			}(),
			exp: true,
		}, "scripts from hex, not equal should return false": {
			script1: func() *bscript.Script {
				s, err := bscript.NewFromHexString("76a91404d03f746652cfcb6cb55119ab473a045137d26566ac")
				assert.NoError(t, err)
				assert.NotNil(t, s)
				return s
			}(),
			script2: func() *bscript.Script {
				s, err := bscript.NewFromHexString("76a91404d03f746652cfcb6cb55119ab473a045137d26588ac")
				assert.NoError(t, err)
				assert.NotNil(t, s)
				return s
			}(),
			exp: false,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.exp, test.script1.Equals(test.script2))
			assert.Equal(t, test.exp, test.script1.EqualsBytes(*test.script2))
			assert.Equal(t, test.exp, test.script1.EqualsHex(test.script2.String()))
		})
	}
}

func TestScript_MinPushSize(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		data   [][]byte
		expLen int
	}{
		"OpX / OpNeg returns 1": {
			data: [][]byte{
				{1}, {2}, {3}, {4}, {5}, {6}, {7}, {8}, {9},
				{10}, {11}, {12}, {13}, {14}, {15}, {16}, {0x81},
			},
			expLen: 1,
		},
		"OP_DATA_1 + data returns 2": {
			data: [][]byte{
				{0x17}, {0x18}, {0x19}, {0x20}, {0x21}, {0x22}, {0x23}, {0x24}, {0x25}, {0x26},
				{0x27}, {0x28}, {0x29}, {0x30}, {0x31}, {0x32}, {0x33}, {0x34}, {0x35}, {0x36},
				{0x37}, {0x38}, {0x39}, {0x40}, {0x41}, {0x42}, {0x43}, {0x44}, {0x45}, {0x46},
				{0x47}, {0x48}, {0x49}, {0x50}, {0x51}, {0x52}, {0x53}, {0x54}, {0x55}, {0x56},
				{0x57}, {0x58}, {0x59}, {0x60}, {0x61}, {0x62}, {0x63}, {0x64}, {0x65}, {0x66},
				{0x67}, {0x68}, {0x69}, {0x70}, {0x71}, {0x72}, {0x73}, {0x74}, {0x75}, {0x76},
				{0x78}, {0x79}, {0x7a}, {0x7b}, {0x7c}, {0x7d}, {0x7e}, {0x7f}, {0x80},
				{0x82}, {0x83}, {0x84}, {0x85}, {0x86}, {0x87}, {0x88}, {0x89}, {0x8a}, {0x8b},
				{0x8c}, {0x8d}, {0x8e}, {0x8f}, {0x90}, {0x91}, {0x92}, {0x93}, {0x94}, {0x95},
				{0x96}, {0x97}, {0x98}, {0x99}, {0x9a}, {0x9b}, {0x9c}, {0x9d}, {0x9e}, {0x9f},
				{0xa0}, {0xa1}, {0xa2}, {0xa3}, {0xa4}, {0xa5}, {0xa6}, {0xa7}, {0xa8}, {0xa9},
				{0xaa}, {0xab}, {0xac}, {0xad}, {0xae}, {0xaf}, {0xb0}, {0xb1}, {0xb2}, {0xb3},
				{0xb4}, {0xb5}, {0xb6}, {0xb7}, {0xb8}, {0xb9}, {0xba}, {0xbb}, {0xbc}, {0xbd},
				{0xbe}, {0xbf}, {0xc0}, {0xc1}, {0xc2}, {0xc3}, {0xc4}, {0xc5}, {0xc6}, {0xc7},
				{0xc8}, {0xc9}, {0xca}, {0xcb}, {0xcc}, {0xcd}, {0xce}, {0xcf}, {0xd0}, {0xd1},
				{0xd2}, {0xd3}, {0xd4}, {0xd5}, {0xd6}, {0xd7}, {0xd8}, {0xd9}, {0xda}, {0xdb},
				{0xdc}, {0xdd}, {0xde}, {0xdf}, {0xe0}, {0xe1}, {0xe2}, {0xe3}, {0xe4}, {0xe5},
				{0xe6}, {0xe7}, {0xe8}, {0xe9}, {0xea}, {0xeb}, {0xec}, {0xed}, {0xee}, {0xef},
				{0xf0}, {0xf1}, {0xf2}, {0xf3}, {0xf4}, {0xf5}, {0xf6}, {0xf7}, {0xf8}, {0xf9},
				{0xfa}, {0xfb}, {0xfc}, {0xfd}, {0xfe}, {0xff},
			},
			expLen: 2,
		},
		"OP_DATA_2 onward returns len(data)+1": {
			data: [][]byte{func() []byte {
				return bytes.Repeat([]byte{0x00}, 23)
			}()},
			expLen: 23 + 1,
		},
		"OP_DATA_75 returns len(data)+1 (max)": {
			data: [][]byte{func() []byte {
				return bytes.Repeat([]byte{0x00}, 75)
			}()},
			expLen: 75 + 1,
		},
		"OP_PUSHDATA1 + length byte + data returns len(data)+2": {
			data: [][]byte{func() []byte {
				return bytes.Repeat([]byte{0x00}, 86)
			}()},
			expLen: 86 + 2,
		},
		"OP_PUSHDATA1 + length byte + data returns len(data)+2 (max)": {
			data: [][]byte{func() []byte {
				return bytes.Repeat([]byte{0x00}, 255)
			}()},
			expLen: 255 + 2,
		},
		"OP_PUSHDATA2 + length byte + data returns len(data)+3": {
			data: [][]byte{func() []byte {
				return bytes.Repeat([]byte{0x00}, 256)
			}()},
			expLen: 256 + 3,
		},
		"OP_PUSHDATA2 + length byte + data returns len(data)+3 (max)": {
			data: [][]byte{func() []byte {
				return bytes.Repeat([]byte{0x00}, 65535)
			}()},
			expLen: 65535 + 3,
		},
		"OP_PUSHDATA4 + length byte + data returns len(data)+5": {
			data: [][]byte{func() []byte {
				return bytes.Repeat([]byte{0x00}, 65536)
			}()},
			expLen: 65536 + 5,
		},
		// These tests cause the CI to OOM due to the massive slices being created
		//"OP_PUSHDATA4 + length byte + data returns len(data)+5 (max)": {
		//	data: [][]byte{func() []byte {
		//		return bytes.Repeat([]byte{0x00}, 0xffffffff)
		//	}()},
		//	expLen: 0xffffffff + 5,
		//},
		//"data too large returns 0": {
		//	data: [][]byte{func() []byte {
		//		return bytes.Repeat([]byte{0x00}, 0xffffffff+1)
		//	}()},
		//	expLen: 0,
		//},
		"Op0 returns 1": {
			data:   [][]byte{},
			expLen: 1,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			for _, data := range test.data {
				assert.Equal(t, test.expLen, bscript.MinPushSize(data), "data: %x", data)
			}
		})
	}
}

func TestScript_MarshalJSON(t *testing.T) {
	script, err := bscript.NewFromASM("OP_2 OP_2 OP_ADD OP_4 OP_EQUALVERIFY")
	assert.NoError(t, err)

	bb, err := json.Marshal(script)
	assert.NoError(t, err)

	assert.Equal(t, `"5252935488"`, string(bb))
}

func TestScript_UnmarshalJSON(t *testing.T) {
	tests := map[string]struct {
		jsonString string
		exp        string
	}{
		"script with content": {
			jsonString: `"5252935488"`,
			exp:        "5252935488",
		},
		"empty script": {
			jsonString: `""`,
			exp:        "",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			var out *bscript.Script
			assert.NoError(t, json.Unmarshal([]byte(test.jsonString), &out))
			assert.Equal(t, test.exp, out.String())
		})
	}
}
