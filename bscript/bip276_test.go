package bscript_test

import (
	"fmt"
	"testing"

	"github.com/libsv/go-bt/bscript"
	"github.com/stretchr/testify/assert"
)

func TestEncodeBIP276(t *testing.T) {
	t.Parallel()

	t.Run("valid encode (mainnet)", func(t *testing.T) {
		s := bscript.EncodeBIP276(
			bscript.PrefixScript,
			bscript.NetworkMainnet,
			bscript.CurrentVersion,
			[]byte("fake script"),
		)

		assert.Equal(t, "bitcoin-script:010166616b65207363726970746f0cd86a", s)
	})

	t.Run("valid encode (testnet)", func(t *testing.T) {
		s := bscript.EncodeBIP276(
			bscript.PrefixScript,
			bscript.NetworkTestnet,
			bscript.CurrentVersion,
			[]byte("fake script"),
		)

		assert.Equal(t, "bitcoin-script:020166616b65207363726970742577a444", s)
	})

	t.Run("invalid version = 0", func(t *testing.T) {
		s := bscript.EncodeBIP276(
			bscript.PrefixScript,
			bscript.NetworkMainnet,
			0,
			[]byte("fake script"),
		)

		assert.Equal(t, "ERROR", s)
	})

	t.Run("invalid version > 255", func(t *testing.T) {
		s := bscript.EncodeBIP276(
			bscript.PrefixScript,
			bscript.NetworkMainnet,
			256,
			[]byte("fake script"),
		)

		assert.Equal(t, "ERROR", s)
	})

	t.Run("invalid network = 0", func(t *testing.T) {
		s := bscript.EncodeBIP276(
			bscript.PrefixScript,
			0,
			bscript.CurrentVersion,
			[]byte("fake script"),
		)

		assert.Equal(t, "ERROR", s)
	})

	t.Run("invalid version > 255", func(t *testing.T) {
		s := bscript.EncodeBIP276(
			bscript.PrefixScript,
			256,
			bscript.CurrentVersion,
			[]byte("fake script"),
		)

		assert.Equal(t, "ERROR", s)
	})

	t.Run("different prefix", func(t *testing.T) {
		s := bscript.EncodeBIP276(
			"different-prefix",
			bscript.NetworkMainnet,
			bscript.CurrentVersion,
			[]byte("fake script"),
		)

		assert.Equal(t, "different-prefix:010166616b6520736372697074effdb090", s)
	})

	t.Run("template prefix", func(t *testing.T) {
		s := bscript.EncodeBIP276(
			bscript.PrefixTemplate,
			bscript.NetworkMainnet,
			bscript.CurrentVersion,
			[]byte("fake script"),
		)

		assert.Equal(t, "bitcoin-template:010166616b65207363726970749e31aa72", s)
	})
}

func TestDecodeBIP276(t *testing.T) {
	t.Parallel()

	t.Run("valid decode", func(t *testing.T) {
		prefix, network, version, data, err := bscript.DecodeBIP276("bitcoin-script:010166616b65207363726970746f0cd86a")
		assert.NoError(t, err)
		assert.Equal(t, `"bitcoin-script"`, fmt.Sprintf("%q", prefix))
		assert.Equal(t, 1, network)
		assert.Equal(t, 1, version)
		assert.Equal(t, "fake script", fmt.Sprintf("%s", data))
	})

	t.Run("panic - invalid decode", func(t *testing.T) {
		assert.Panics(t, func() {
			_, _, _, _, err := bscript.DecodeBIP276("bitcoin-script:01")
			assert.Error(t, err)
		})
	})

	t.Run("valid format, bad checksum", func(t *testing.T) {
		prefix, network, version, data, err := bscript.DecodeBIP276("bitcoin-script:010166616b65207363726970746f0cd8")
		assert.Error(t, err)
		assert.Equal(t, `"bitcoin-script"`, fmt.Sprintf("%q", prefix))
		assert.Equal(t, 1, network)
		assert.Equal(t, 1, version)
		assert.Equal(t, "fake scrip", fmt.Sprintf("%s", data))
	})
}
