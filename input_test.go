package bt_test

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/libsv/go-bt"
	"github.com/libsv/go-bt/bscript"
	"github.com/stretchr/testify/assert"
)

const inputHexStr = "4c6ec863cf3e0284b407a1a1b8138c76f98280812cb9653231f385a0305fc76f010000006b483045022100f01c1a1679c9437398d691c8497f278fa2d615efc05115688bf2c3335b45c88602201b54437e54fb53bc50545de44ea8c64e9e583952771fcc663c8687dc2638f7854121037e87bbd3b680748a74372640628a8f32d3a841ceeef6f75626ab030c1a04824fffffffff"

func TestNewInput(t *testing.T) {
	t.Parallel()

	b, err := hex.DecodeString(inputHexStr)
	assert.NoError(t, err)

	var i *bt.Input
	var s int
	i, s, err = bt.NewInputFromBytes(b)
	assert.NoError(t, err)
	assert.NotNil(t, i)
	assert.Equal(t, 148, s)
	assert.Equal(t, uint32(1), i.PreviousTxOutIndex)
	assert.Equal(t, 107, len(*i.UnlockingScript))
	assert.Equal(t, uint32(0xFFFFFFFF), i.SequenceNumber)
}

func TestNewInputFromUTXO(t *testing.T) {
	t.Parallel()

	i, err := bt.NewInputFromUTXO("a61021694ee0fd7c3d441aab7b387e356f5552957d5a01705a66766fe86ec9e5", 4, 5064, "a9149cbe9f5e72fa286ac8a38052d1d5337aa363ea7f88ac", 0xffffffff)
	assert.NoError(t, err)

	assert.Equal(t, "a61021694ee0fd7c3d441aab7b387e356f5552957d5a01705a66766fe86ec9e5", i.PreviousTxID)
	assert.Equal(t, uint32(4), i.PreviousTxOutIndex)
	assert.Equal(t, uint64(5064), i.PreviousTxSatoshis)

	var es *bscript.Script
	es, err = bscript.NewFromHexString("a9149cbe9f5e72fa286ac8a38052d1d5337aa363ea7f88ac")
	assert.NoError(t, err)
	assert.NotNil(t, es)

	if !bytes.Equal(*i.PreviousTxScript, *es) {
		t.Errorf("Expected a9149cbe9f5e72fa286ac8a38052d1d5337aa363ea7f88ac, got %x", *i.PreviousTxScript)
	}

	assert.Equal(t, uint32(0xFFFFFFFF), i.SequenceNumber)
}
