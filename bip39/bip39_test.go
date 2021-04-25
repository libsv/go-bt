package bip39

import (
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/pbkdf2"
)

func TestEntropy(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		ent uint32
		expLen int
		err error
	}{
		"successful run should return no errors": {
			ent: 128,
			expLen: 12,
			err: nil,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			out, err := Words(test.ent)
			fmt.Println(err)
			assert.EqualError(t, err, "")
			assert.Equal(t, test.expLen, len(out))
			fmt.Printf("%+v\n", out)
			fmt.Println(strings.Join(out, " "))
			privkey := pbkdf2.Key([]byte(strings.Join(out, " ")), []byte("mnemonic"+"test"),2048, 64, sha512.New)
			fmt.Println(hex.EncodeToString(privkey))

		})
	}
}
