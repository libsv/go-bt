package bip39

import (
	"crypto/rand"
	"crypto/sha256"
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
	cs := ent / 32
	ms := ent + cs
	sb := strings.Builder{}
	for _, b := range bb {
		sb.WriteString(fmt.Sprintf("%08b", b))
	}
	sb.WriteString(fmt.Sprintf("%08b", sha256.Sum256(bb)[0])[:cs-1])
	bitString := sb.String()
	words := make([]string, 0, ms/11)
	for i := 10; i < int(ms); i += 11 {
		output, err := strconv.ParseInt(bitString[i-10:i], 2, 32)
		if err != nil{
			return nil, fmt.Errorf("failed to convert binary to int %w", err)
		}
		words = append(words, English[output])
	}
	return words, nil
}
