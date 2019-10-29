package v2

// ECDSA signer implementation as well as various crypto-related utility functions.

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
	"math/big"

	"github.com/btcsuite/btcd/btcec"
	"golang.org/x/crypto/ripemd160"
)

const (
	alphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"
)

var (
	indexes  []int
	bigRadix = big.NewInt(58)
	bigZero  = big.NewInt(0)
)

func encodeAddress(hash160 []byte, key byte) string {
	tosum := make([]byte, 21)
	tosum[0] = key
	copy(tosum[1:], hash160)
	cksum := doubleHash(tosum)

	// Address before base58 encoding is 1 byte for netID, ripemd160 hash
	// size, plus 4 bytes of checksum (total 25).
	b := make([]byte, 25)
	b[0] = key
	copy(b[1:], hash160)
	copy(b[21:], cksum[:4])

	return base58Encode(b)
}

func pubKeyFromPrivate(private []byte) []byte {
	_, pubkey := btcec.PrivKeyFromBytes(btcec.S256(), private)
	//pubkeyaddr  := &pubkey
	return pubkey.SerializeCompressed()
}

func hash160(data []byte) []byte {
	if len(data) == 1 && data[0] == 0 {
		data = []byte{}
	}
	h1 := sha256.Sum256(data)
	h2 := ripemd160.New()
	h2.Write(h1[:])
	return h2.Sum(nil)
}

func doubleHash(data []byte) []byte {
	h1 := sha256.Sum256(data)
	h2 := sha256.Sum256(h1[:])
	return h2[:]
}

func encrypt(ciph cipher.Block, text []byte) ([]byte, error) {
	b := base64.StdEncoding.EncodeToString(text)
	ciphertext := make([]byte, aes.BlockSize+len(b))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}
	cfb := cipher.NewCFBEncrypter(ciph, iv)
	cfb.XORKeyStream(ciphertext[aes.BlockSize:], []byte(b))
	return ciphertext, nil
}

func decrypt(ciph cipher.Block, ciphertext []byte) ([]byte, error) {
	if len(ciphertext) < aes.BlockSize {
		return nil, errors.New("ciphertext too short")
	}
	iv := ciphertext[:aes.BlockSize]
	cfb := cipher.NewCFBDecrypter(ciph, iv)
	text := ciphertext[aes.BlockSize:]
	cfb.XORKeyStream(text, text)
	data, err := base64.StdEncoding.DecodeString(string(text))
	if err != nil {
		return nil, err
	}
	return data, nil
}

func base58Decode(b string) []byte {
	if indexes == nil {
		indexes = make([]int, 128)
		for i := 0; i < len(indexes); i++ {
			indexes[i] = -1
		}
		for i := 0; i < len(alphabet); i++ {
			indexes[alphabet[i]] = i
		}
	}

	if len(b) == 0 {
		return []byte{}
	}
	input58 := make([]byte, len(b))
	for n, ch := range b {
		digit58 := -1
		if ch >= 0 && ch < 128 {
			digit58 = indexes[ch]
		}
		if digit58 < 0 {
			return []byte{}
		}

		input58[n] = byte(digit58)
	}
	zeroCount := 0
	for zeroCount < len(input58) && input58[zeroCount] == 0 {
		zeroCount++
	}

	// The encoding
	temp := make([]byte, len(b))
	j := len(temp)

	startAt := zeroCount
	for startAt < len(input58) {
		mod := divmod256(input58, startAt)
		if input58[startAt] == 0 {
			startAt++
		}

		j--
		temp[j] = mod
	}
	// Do no add extra leading zeroes, move j to first non null byte.
	for j < len(temp) && temp[j] == 0 {
		j++
	}

	return temp[j-zeroCount : len(temp)]
}

func divmod256(number58 []byte, startAt int) byte {
	remainder := 0
	for i := startAt; i < len(number58); i++ {
		digit58 := int(number58[i] & 0xFF)
		temp := remainder*58 + digit58

		number58[i] = byte(temp / 256)
		remainder = temp % 256
	}

	return byte(remainder)
}

// Base58Encode encodes a byte slice to a modified base58 string.
func base58Encode(b []byte) string {
	x := new(big.Int)
	x.SetBytes(b)

	answer := make([]byte, 0)
	for x.Cmp(bigZero) > 0 {
		mod := new(big.Int)
		x.DivMod(x, bigRadix, mod)
		answer = append(answer, alphabet[mod.Int64()])
	}

	// leading zero bytes
	for _, i := range b {
		if i != 0 {
			break
		}
		answer = append(answer, alphabet[0])
	}

	// reverse
	alen := len(answer)
	for i := 0; i < alen/2; i++ {
		answer[i], answer[alen-1-i] = answer[alen-1-i], answer[i]
	}

	return string(answer)
}
