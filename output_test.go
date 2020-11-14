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
	assert.Equal(t, "76a9148bf10d323ac757268eb715e613cb8e8e1d1793aa88ac", o.GetLockingScriptHexString())
}

func TestNewP2PKHOutputFromPubKeyHash(t *testing.T) {
	t.Parallel()

	// This is the PKH for address mtdruWYVEV1wz5yL7GvpBj4MgifCB7yhPd
	o, err := bt.NewP2PKHOutputFromPubKeyHash(
		"8fe80c75c9560e8b56ed64ea3c26e18d2c52211b",
		uint64(5000),
	)
	assert.NoError(t, err)
	assert.NotNil(t, o)
	assert.Equal(t,
		"76a9148fe80c75c9560e8b56ed64ea3c26e18d2c52211b88ac",
		o.GetLockingScriptHexString(),
	)
}

func TestNewHashPuzzleOutput(t *testing.T) {
	t.Parallel()

	addr, err := bscript.NewAddressFromString("myFhJggmsaA2S8Qe6ZQDEcVCwC4wLkvC4e")
	assert.NoError(t, err)
	assert.NotNil(t, addr)

	var o *bt.Output
	o, err = bt.NewHashPuzzleOutput("secret1", addr.PublicKeyHash, uint64(5000))
	assert.NoError(t, err)
	assert.NotNil(t, addr)
	assert.Equal(t,
		"a914d3f9e3d971764be5838307b175ee4e08ba427b908876a914c28f832c3d539933e0c719297340b34eee0f4c3488ac",
		o.GetLockingScriptHexString(),
	)
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

	script := o.GetLockingScriptHexString()
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

	assert.Equal(t, "006a02686903686f770361726503796f75", o.GetLockingScriptHexString())
}
