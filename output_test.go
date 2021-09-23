package bt_test

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/libsv/go-bt/v2"
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

		assert.Equal(t, "value:     1252788362\nscriptLen: 25\nscript:    76a9148bf10d323ac757268eb715e613cb8e8e1d1793aa88ac\n", o.String())
	})
}
