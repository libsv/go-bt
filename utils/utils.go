package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/btcsuite/btcd/btcec"
	"golang.org/x/crypto/ripemd160"
	"io"
	"math/big"
	"strconv"
)

//region V2 Utils
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

	return Base58Encode(b)
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

	return temp[j-zeroCount:]
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
func Base58Encode(b []byte) string {
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

//endregion

//region V1 Utils

// DifficultyFromBits returns the mining difficulty from the nBits field in the block header.
func DifficultyFromBits(bits string) (float64, error) {
	b, _ := hex.DecodeString(bits)
	ib := binary.BigEndian.Uint32(b)
	return targetToDifficulty(toCompactSize(ib))
}

func toCompactSize(bits uint32) *big.Int {
	t := big.NewInt(int64(bits % 0x01000000))
	t.Mul(t, big.NewInt(2).Exp(big.NewInt(2), big.NewInt(8*(int64(bits/0x01000000)-3)), nil))

	return t
}

func targetToDifficulty(target *big.Int) (float64, error) {
	a := float64(0xFFFF0000000000000000000000000000000000000000000000000000) // genesis difficulty
	b, err := strconv.ParseFloat(target.String(), 64)
	if err != nil {
		return 0.0, err
	}
	return a / b, nil
}

// GetLittleEndianBytes returns a byte array in little endian from an unsigned integer of 32 bytes.
func GetLittleEndianBytes(v uint32, l uint32) []byte {
	// TODO: is v hex encoded?
	buf := make([]byte, l)

	binary.LittleEndian.PutUint32(buf, v)

	return buf
}

// VarInt takes an unsiged integer and  returns a byte array in VarInt format.
// See http://learnmeabitcoin.com/glossary/varint
func VarInt(i uint64) []byte {
	b := make([]byte, 9)
	if i < 0xfd {
		b[0] = byte(i)
		return b[:1]
	}
	if i < 0x10000 {
		b[0] = 0xfd
		binary.LittleEndian.PutUint16(b[1:3], uint16(i))
		return b[:3]
	}
	if i < 0x100000000 {
		b[0] = 0xfe
		binary.LittleEndian.PutUint32(b[1:5], uint32(i))
		return b[:5]
	}
	b[0] = 0xff
	binary.LittleEndian.PutUint64(b[1:9], i)
	return b
}

// DecodeVarInt takes a byte array in VarInt format and returns the
// decoded unsiged integer value and it's size in bytes.
// See http://learnmeabitcoin.com/glossary/varint
func DecodeVarInt(b []byte) (result uint64, size int) {
	switch b[0] {
	case 0xff:
		result = binary.LittleEndian.Uint64(b[1:9])
		size = 9

	case 0xfe:
		result = uint64(binary.LittleEndian.Uint32(b[1:5]))
		size = 5

	case 0xfd:
		result = uint64(binary.LittleEndian.Uint16(b[1:3]))
		size = 3

	default:
		result = uint64(binary.LittleEndian.Uint16([]byte{b[0], 0x00}))
		size = 1
	}

	return
}

// EncodeParts takes a slice of slices and returns a single slice with the appropriate OP_PUSH commands embedded.
func EncodeParts(parts [][]byte) ([]byte, error) {
	b := make([]byte, 0)

	for i, part := range parts {
		l := int64(len(part))

		if l <= 75 {
			b = append(b, byte(len(part)))
			b = append(b, part...)

		} else if l <= 0xFF {
			b = append(b, 0x4c) // OP_PUSHDATA1
			b = append(b, byte(len(part)))
			b = append(b, part...)

		} else if l <= 0xFFFF {
			b = append(b, 0x4d) // OP_PUSHDATA2
			lenBuf := make([]byte, 2)
			binary.LittleEndian.PutUint16(lenBuf, uint16(len(part)))
			b = append(b, lenBuf...)
			b = append(b, part...)

		} else if l <= 0xFFFFFFFF {
			b = append(b, 0x4e) // OP_PUSHDATA4
			lenBuf := make([]byte, 4)
			binary.LittleEndian.PutUint32(lenBuf, uint32(len(part)))
			b = append(b, lenBuf...)
			b = append(b, part...)

		} else {
			return nil, fmt.Errorf("Part %d is too big", i)
		}
	}

	return b, nil
}

// DecodeStringParts calls DecodeParts.
func DecodeStringParts(s string) ([][]byte, error) {
	b, err := hex.DecodeString(s)
	if err != nil {
		return nil, err
	}
	return DecodeParts(b)
}

// DecodeParts returns an array of strings...
func DecodeParts(b []byte) ([][]byte, error) {
	var r [][]byte
	for len(b) > 0 {
		// Handle OP codes
		switch b[0] {
		case OpPUSHDATA1:
			length := b[1]
			part := b[2 : 2+length]
			r = append(r, part)
			b = b[2+length:]

		case OpPUSHDATA2:
			length := binary.LittleEndian.Uint16(b[1:])
			part := b[3 : 3+length]
			r = append(r, part)
			b = b[3+length:]

		case OpPUSHDATA4:
			length := binary.LittleEndian.Uint32(b[1:])
			part := b[5 : 5+length]
			r = append(r, part)
			b = b[5+length:]

		default:
			if b[0] >= 0x01 && b[0] <= 0x4e {
				length := b[0]
				part := b[1 : length+1]
				r = append(r, part)
				b = b[1+length:]
			} else {
				r = append(r, []byte{b[0]})
				b = b[1:]
			}
		}
	}

	return r, nil
}

//endregion
