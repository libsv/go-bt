package bt_test

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/libsv/go-bt"
	"github.com/libsv/go-bt/bscript"
	"github.com/stretchr/testify/assert"
)

const outputHexStr = "8a08ac4a000000001976a9148bf10d323ac757268eb715e613cb8e8e1d1793aa88ac00000000"

func TestNewOutputFromBytes(t *testing.T) {
	t.Parallel()

	t.Run("invalid output, too short", func(t *testing.T) {
		o, s, err := bt.NewOutputFromBytes([]byte(""))
		assert.Error(t, err)
		assert.Nil(t, o)
		assert.Equal(t, 0, s)
	})

	t.Run("invalid output, too short + script", func(t *testing.T) {
		o, s, err := bt.NewOutputFromBytes([]byte("0000000000000"))
		assert.Error(t, err)
		assert.Nil(t, o)
		assert.Equal(t, 0, s)
	})

	t.Run("valid output", func(t *testing.T) {
		bytes, err := hex.DecodeString(outputHexStr)
		assert.NoError(t, err)

		var o *bt.Output
		var s int
		o, s, err = bt.NewOutputFromBytes(bytes)
		assert.NoError(t, err)
		assert.NotNil(t, o)

		assert.Equal(t, 34, s)
		assert.Equal(t, uint64(1252788362), o.Satoshis)
		assert.Equal(t, 25, len(*o.LockingScript))
		assert.Equal(t, "76a9148bf10d323ac757268eb715e613cb8e8e1d1793aa88ac", o.LockingScriptHexString())
	})
}

func TestOutput_String(t *testing.T) {
	t.Run("compare string output", func(t *testing.T) {

		bytes, err := hex.DecodeString(outputHexStr)
		assert.NoError(t, err)

		var o *bt.Output
		o, _, err = bt.NewOutputFromBytes(bytes)
		assert.NoError(t, err)
		assert.NotNil(t, o)

		assert.Equal(t, "value:     1252788362\nscriptLen: 25\nscript:    &76a9148bf10d323ac757268eb715e613cb8e8e1d1793aa88ac\n", o.String())
	})
}

func TestNewP2PKHOutputFromPubKeyHashStr(t *testing.T) {
	t.Parallel()

	t.Run("empty pubkey hash", func(t *testing.T) {
		o, err := bt.NewP2PKHOutputFromPubKeyHashStr(
			"",
			uint64(5000),
		)
		assert.NoError(t, err)
		assert.NotNil(t, o)
		assert.Equal(t,
			"76a91488ac",
			o.LockingScriptHexString(),
		)
	})

	t.Run("invalid pubkey hash", func(t *testing.T) {
		o, err := bt.NewP2PKHOutputFromPubKeyHashStr(
			"0",
			uint64(5000),
		)
		assert.Error(t, err)
		assert.Nil(t, o)
	})

	t.Run("valid output", func(t *testing.T) {
		// This is the PKH for address mtdruWYVEV1wz5yL7GvpBj4MgifCB7yhPd
		o, err := bt.NewP2PKHOutputFromPubKeyHashStr(
			"8fe80c75c9560e8b56ed64ea3c26e18d2c52211b",
			uint64(5000),
		)
		assert.NoError(t, err)
		assert.NotNil(t, o)
		assert.Equal(t,
			"76a9148fe80c75c9560e8b56ed64ea3c26e18d2c52211b88ac",
			o.LockingScriptHexString(),
		)
	})
}

func TestNewHashPuzzleOutput(t *testing.T) {
	t.Parallel()

	t.Run("invalid public key", func(t *testing.T) {
		o, err := bt.NewHashPuzzleOutput("", "0", uint64(5000))
		assert.Error(t, err)
		assert.Nil(t, o)
	})

	t.Run("missing secret and public key", func(t *testing.T) {
		o, err := bt.NewHashPuzzleOutput("", "", uint64(5000))
		assert.NoError(t, err)
		assert.Equal(t,
			"a914b472a266d0bd89c13706a4132ccfb16f7c3b9fcb8876a90088ac",
			o.LockingScriptHexString(),
		)
	})

	t.Run("valid puzzle output", func(t *testing.T) {
		addr, err := bscript.NewAddressFromString("myFhJggmsaA2S8Qe6ZQDEcVCwC4wLkvC4e")
		assert.NoError(t, err)
		assert.NotNil(t, addr)

		var o *bt.Output
		o, err = bt.NewHashPuzzleOutput("secret1", addr.PublicKeyHash, uint64(5000))
		assert.NoError(t, err)
		assert.NotNil(t, o)
		assert.Equal(t,
			"a914d3f9e3d971764be5838307b175ee4e08ba427b908876a914c28f832c3d539933e0c719297340b34eee0f4c3488ac",
			o.LockingScriptHexString(),
		)
	})
}

func TestNewOpReturnOutput(t *testing.T) {
	t.Parallel()

	data := "On February 4th, 2020 The Return to Genesis was activated to restore the Satoshi Vision for Bitcoin. " +
		"It is locked in irrevocably by this transaction. Bitcoin can finally be Bitcoin again and the miners can " +
		"continue to write the Chronicle of everything. Thank you and goodnight from team SV."
	dataBytes := []byte(data)
	o, err := bt.NewOpReturnOutput(dataBytes)
	assert.NoError(t, err)
	assert.NotNil(t, o)

	script := o.LockingScriptHexString()
	dataLength := bt.VarInt(uint64(len(dataBytes)))

	assert.Equal(t, "006a4d2201"+hex.EncodeToString(dataBytes), script)
	assert.Equal(t, "fd2201", fmt.Sprintf("%x", dataLength))
}

func TestNewOpReturnPartsOutput(t *testing.T) {
	t.Parallel()

	dataBytes := [][]byte{[]byte("hi"), []byte("how"), []byte("are"), []byte("you")}
	o, err := bt.NewOpReturnPartsOutput(dataBytes)
	assert.NoError(t, err)
	assert.NotNil(t, o)

	assert.Equal(t, "006a02686903686f770361726503796f75", o.LockingScriptHexString())
}
