package output_test

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/libsv/libsv/bt/output"
	"github.com/libsv/libsv/script/address"
	"github.com/libsv/libsv/utils"
)

const outputHexStr = "8a08ac4a000000001976a9148bf10d323ac757268eb715e613cb8e8e1d1793aa88ac00000000"

func TestNewOutput(t *testing.T) {
	bytes, _ := hex.DecodeString(outputHexStr)
	o, s, err := output.NewFromBytes(bytes)
	if err != nil {
		t.Errorf("Invalid outputHexStr")
	}

	// t.Errorf("\n%s\n", o)
	if s != 34 {
		t.Errorf("Expected 25, got %d", s)
	}

	if o.Satoshis != 1252788362 {
		t.Errorf("Expected 1252788362, got %d", o.Satoshis)
	}

	if len(*o.LockingScript) != 25 {
		t.Errorf("Expected 25, got %d", len(*o.LockingScript))
	}

	script := o.GetLockingScriptHexString()
	if script != "76a9148bf10d323ac757268eb715e613cb8e8e1d1793aa88ac" {
		t.Errorf("Expected 76a9148bf10d323ac757268eb715e613cb8e8e1d1793aa88ac, got %x", script)
	}
}

func TestNewOutputForPublicKeyHash(t *testing.T) {
	publicKeyhash := "8fe80c75c9560e8b56ed64ea3c26e18d2c52211b" // This is the PKH for address mtdruWYVEV1wz5yL7GvpBj4MgifCB7yhPd
	value := uint64(5000)
	o, err := output.NewP2PkhFromPubKeyHash(publicKeyhash, value)
	if err != nil {
		t.Error("Error")
	}
	expected := "76a9148fe80c75c9560e8b56ed64ea3c26e18d2c52211b88ac"
	if o.GetLockingScriptHexString() != expected {
		t.Errorf("Error script not correct\nExpected: %s\n     Got: %s\n", expected, o.GetLockingScriptHexString())
	}
}

func TestNewOutputForHashPuzzle(t *testing.T) {
	secret := "secret1"
	addr, _ := address.NewFromString("myFhJggmsaA2S8Qe6ZQDEcVCwC4wLkvC4e")
	value := uint64(5000)
	o, err := output.NewHashPuzzle(secret, addr.PublicKeyHash, value)
	if err != nil {
		t.Error("Error")
	}
	expected := "a914d3f9e3d971764be5838307b175ee4e08ba427b908876a914c28f832c3d539933e0c719297340b34eee0f4c3488ac"
	if o.GetLockingScriptHexString() != expected {
		t.Errorf("Error script not correct\nExpected: %s\n     Got: %s\n", expected, o.GetLockingScriptHexString())
	}
}

func TestNewOutputOpReturn(t *testing.T) {
	data := "On February 4th, 2020 The Return to Genesis was activated to restore the Satoshi Vision for Bitcoin. It is locked in irrevocably by this transaction. Bitcoin can finally be Bitcoin again and the miners can continue to write the Chronicle of everything. Thank you and goodnight from team SV."
	dataBytes := []byte(data)
	o, err := output.NewOpReturn(dataBytes)
	if err != nil {
		t.Error(err)
		return
	}
	dataHexStr := hex.EncodeToString(dataBytes)
	script := o.GetLockingScriptHexString()
	dataLength := utils.VarInt(uint64(len(dataBytes)))
	fmt.Printf("%x", dataLength)
	expectedScript := "006a4d2201" + dataHexStr

	if script != expectedScript {
		t.Errorf("Error op return hex expected %s, got %s", expectedScript, script)
	}
}

func TestNewOutputOpReturnPush(t *testing.T) {
	data1 := "hi"
	data2 := "how"
	data3 := "are"
	data4 := "you"
	dataBytes := [][]byte{[]byte(data1), []byte(data2), []byte(data3), []byte(data4)}
	o, err := output.NewOpReturnParts(dataBytes)
	if err != nil {
		t.Error(err)
		return
	}

	script := o.GetLockingScriptHexString()
	expectedScript := "006a02686903686f770361726503796f75"

	if script != expectedScript {
		t.Errorf("Error op return hex expected %s, got %s", expectedScript, script)
	}
}
