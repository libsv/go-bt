package bscript_test

import (
	"fmt"
	"testing"

	"github.com/libsv/go-bt/bscript"
	"github.com/stretchr/testify/assert"
)

func TestEncode(t *testing.T) {
	s := bscript.EncodeBIP276(
		bscript.PrefixScript,
		bscript.NetworkMainnet,
		bscript.CurrentVersion,
		[]byte("fake script"),
	)

	assert.Equal(t, "bitcoin-script:010166616b65207363726970746f0cd86a", s)
}

func TestDecode(t *testing.T) {
	prefix, network, version, data, err := bscript.DecodeBIP276("bitcoin-script:010166616b65207363726970746f0cd86a")
	assert.NoError(t, err)
	assert.Equal(t, `"bitcoin-script"`, fmt.Sprintf("%q", prefix))
	assert.Equal(t, 1, network)
	assert.Equal(t, 1, version)
	assert.Equal(t, "fake script", fmt.Sprintf("%s", data))
}
