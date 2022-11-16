package bt

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewInputFromReader(t *testing.T) {
	t.Parallel()

	t.Run("valid tx", func(t *testing.T) {
		rawHex := "4c6ec863cf3e0284b407a1a1b8138c76f98280812cb9653231f385a0305fc76f010000006b483045022100f01c1a1679c9437398d691c8497f278fa2d615efc05115688bf2c3335b45c88602201b54437e54fb53bc50545de44ea8c64e9e583952771fcc663c8687dc2638f7854121037e87bbd3b680748a74372640628a8f32d3a841ceeef6f75626ab030c1a04824fffffffff"
		b, err := hex.DecodeString(rawHex)
		assert.NoError(t, err)

		i := &Input{}
		var s int64
		s, err = i.readFrom(bytes.NewReader(b), false)

		assert.NoError(t, err)
		assert.NotNil(t, i)
		assert.Equal(t, int64(148), s)
		assert.Equal(t, uint32(1), i.PreviousTxOutIndex)
		assert.Equal(t, 107, len(*i.UnlockingScript))
		assert.Equal(t, DefaultSequenceNumber, i.SequenceNumber)
	})

	t.Run("empty bytes", func(t *testing.T) {
		i := &Input{}

		s, err := i.readFrom(bytes.NewReader([]byte("")), false)
		assert.Error(t, err)
		assert.Equal(t, int64(0), s)
	})

	t.Run("invalid input, too short", func(t *testing.T) {
		i := &Input{}
		s, err := i.readFrom(bytes.NewReader([]byte("invalid")), false)
		assert.Error(t, err)
		assert.Equal(t, int64(7), s)
	})

	t.Run("invalid input, too short + script", func(t *testing.T) {
		i := &Input{}
		s, err := i.readFrom(bytes.NewReader([]byte("000000000000000000000000000000000000000000000000000000000000000000000000")), false)
		assert.Error(t, err)
		assert.Equal(t, int64(72), s)
	})
}

func TestInput_String(t *testing.T) {
	t.Run("valid tx", func(t *testing.T) {
		rawHex := "4c6ec863cf3e0284b407a1a1b8138c76f98280812cb9653231f385a0305fc76f010000006b483045022100f01c1a1679c9437398d691c8497f278fa2d615efc05115688bf2c3335b45c88602201b54437e54fb53bc50545de44ea8c64e9e583952771fcc663c8687dc2638f7854121037e87bbd3b680748a74372640628a8f32d3a841ceeef6f75626ab030c1a04824fffffffff"
		b, err := hex.DecodeString(rawHex)
		assert.NoError(t, err)

		i := &Input{}
		var s int64

		s, err = i.readFrom(bytes.NewReader(b), false)
		assert.NoError(t, err)
		assert.NotNil(t, i)
		assert.Equal(t, int64(148), s)

		assert.Equal(t,
			"prevTxHash:   6fc75f30a085f3313265b92c818082f9768c13b8a1a107b484023ecf63c86e4c\nprevOutIndex: 1\nscriptLen:    107\nscript:       483045022100f01c1a1679c9437398d691c8497f278fa2d615efc05115688bf2c3335b45c88602201b54437e54fb53bc50545de44ea8c64e9e583952771fcc663c8687dc2638f7854121037e87bbd3b680748a74372640628a8f32d3a841ceeef6f75626ab030c1a04824f\nsequence:     ffffffff\n",
			i.String(),
		)
	})
}
