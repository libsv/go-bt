package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

const testKey = "2b7e151628aed2a6abf7158809cf4f3c"

func TestEncrypt(t *testing.T) {
	t.Parallel()

	t.Run("valid aes encryption", func(t *testing.T) {
		key, err := hex.DecodeString(testKey)
		assert.NoError(t, err)

		testStr := "7468697320697320612074657374"
		var testData []byte
		testData, err = hex.DecodeString(testStr)
		assert.NoError(t, err)

		var block cipher.Block
		block, err = aes.NewCipher(key)
		assert.NoError(t, err)
		assert.NotNil(t, block)

		var encrypted []byte
		encrypted, err = Encrypt(block, testData)
		assert.NoError(t, err)
		// t.Logf("%x", encrypted)

		var decrypted []byte
		decrypted, err = Decrypt(block, encrypted)
		assert.NoError(t, err)
		assert.Equal(t, "this is a test", string(decrypted))
		assert.Equal(t, "7468697320697320612074657374", hex.EncodeToString(decrypted))
	})
}

func TestDecrypt(t *testing.T) {
	t.Parallel()

	t.Run("valid aes decryption", func(t *testing.T) {
		key, err := hex.DecodeString(testKey)
		assert.NoError(t, err)

		encryptedString := "16c9de9d806edf8bf7512f1654f0d72c63e4698d61714d1e7c394ada99ef10d8e43c0b22"
		var encryptedData []byte
		encryptedData, err = hex.DecodeString(encryptedString)
		assert.NoError(t, err)

		var block cipher.Block
		block, err = aes.NewCipher(key)
		assert.NoError(t, err)
		assert.NotNil(t, block)

		var decrypted []byte
		decrypted, err = Decrypt(block, encryptedData)
		assert.NoError(t, err)
		assert.Equal(t, "this is a test", string(decrypted))
		assert.Equal(t, "7468697320697320612074657374", hex.EncodeToString(decrypted))
	})

	t.Run("invalid cipher text", func(t *testing.T) {

		key, err := hex.DecodeString(testKey)
		assert.NoError(t, err)

		encryptedString := "000000"
		var encryptedData []byte
		encryptedData, err = hex.DecodeString(encryptedString)
		assert.NoError(t, err)

		var block cipher.Block
		block, err = aes.NewCipher(key)
		assert.NoError(t, err)
		assert.NotNil(t, block)

		_, err = Decrypt(block, encryptedData)
		assert.Error(t, err)
	})
}
