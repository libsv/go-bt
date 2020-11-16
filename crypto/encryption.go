package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

// Encrypt is an encrypt function
// todo: write test and documentation
func Encrypt(cipherBlock cipher.Block, text []byte) ([]byte, error) {
	b := base64.StdEncoding.EncodeToString(text)
	ciphertext := make([]byte, aes.BlockSize+len(b))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}
	cfb := cipher.NewCFBEncrypter(cipherBlock, iv)
	cfb.XORKeyStream(ciphertext[aes.BlockSize:], []byte(b))
	return ciphertext, nil
}

// Decrypt is a decrypt function
// todo: write test and documentation
func Decrypt(cipherBlock cipher.Block, ciphertext []byte) ([]byte, error) {
	if len(ciphertext) < aes.BlockSize {
		return nil, errors.New("ciphertext too short")
	}
	iv := ciphertext[:aes.BlockSize]
	cfb := cipher.NewCFBDecrypter(cipherBlock, iv)
	text := ciphertext[aes.BlockSize:]
	cfb.XORKeyStream(text, text)
	return base64.StdEncoding.DecodeString(string(text))
}
