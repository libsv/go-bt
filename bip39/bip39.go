package bip39

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var(
	ErrInvalidLength = errors.New("invalid number of bits requested, should be a multiple of 32 and between 128 and 256 (inclusive)")
)

func Words(ent uint32) ([]string, error) {
	if ent%32 != 0 || ent < 128 || ent > 256 {
		return nil, ErrInvalidLength
	}
	bb := make([]byte, ent/8)
	_, _ = rand.Read(bb)
	fmt.Println(hex.EncodeToString(bb))
	cs := ent / 32
	ms := int(ent + cs)
	bb = append(bb, sha256.Sum256(bb)[0])
	sb := strings.Builder{}
	sb.Grow(ms)
	for _, b := range bb {
		for t := 7; t >= 0; t--{
			if b & (1 << t) != 0{
				sb.WriteString("1")
				continue
			}
			sb.WriteString("0")
		}
	}
	bitString := sb.String()
	words := make([]string, 0, ms/11)
	for i := 11; i <= ms; i += 11 {
		output, err := strconv.ParseInt(bitString[i-11:i], 2, 32)
		if err != nil{
			return nil, fmt.Errorf("failed to convert binary to int %w", err)
		}
		words = append(words, English[output])
	}
	return words, nil
}
