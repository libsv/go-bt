package script

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/libsv/libsv/crypto"
	"github.com/libsv/libsv/script/address"
)

// Script type
type Script []byte

// NewFromHexString creates a new script from a hex encoded string.
func NewFromHexString(s string) (*Script, error) {
	b, err := hex.DecodeString(s)
	if err != nil {
		return nil, err
	}
	return NewFromBytes(b), nil
}

// NewFromBytes wraps a byte slice with the Script type.
func NewFromBytes(b []byte) *Script {
	s := Script(b)
	return &s
}

// NewFromASM creates a new script from a BitCoin ASM formatted string.
func NewFromASM(str string) (*Script, error) {
	sections := strings.Split(str, " ")

	s := &Script{}

	for _, section := range sections {
		if val, ok := opCodeStrings[section]; ok {
			s.AppendOpCode(val)
		} else {
			err := s.AppendPushDataHexString(section)
			if err != nil {
				return nil, errors.New("invalid opcode data")
			}
		}
	}

	return s, nil
}

// NewP2PKHFromPubKeyStr takes a public key hex string (in
// compressed format) and creates a P2PKH script from it.
func NewP2PKHFromPubKeyStr(pubKey string) (*Script, error) {
	pubKeyBytes, err := hex.DecodeString(pubKey)
	if err != nil {
		return nil, err
	}
	hash := crypto.Hash160(pubKeyBytes)

	b := []byte{
		OP_DUP,
		OP_HASH160,
		0x14,
	}
	b = append(b, hash...)
	b = append(b, OP_EQUALVERIFY)
	b = append(b, OP_CHECKSIG)

	s := Script(b)
	return &s, nil
}

// NewP2PKHFromPubKeyHashStr takes a public key hex string (in
// compressed format) and creates a P2PKH script from it.
func NewP2PKHFromPubKeyHashStr(pubKeyHash string) (*Script, error) {
	hash, err := hex.DecodeString(pubKeyHash)
	if err != nil {
		return nil, err
	}

	b := []byte{
		OP_DUP,
		OP_HASH160,
		0x14,
	}
	b = append(b, hash...)
	b = append(b, OP_EQUALVERIFY)
	b = append(b, OP_CHECKSIG)

	s := Script(b)
	return &s, nil
}

// NewP2PKHFromAddress takes an address
// and creates a P2PKH script from it.
func NewP2PKHFromAddress(addr string) (*Script, error) {

	a, err := address.NewFromString(addr)
	if err != nil {
		return nil, err
	}

	publicKeyHashBytes, err := hex.DecodeString(a.PublicKeyHash)
	if err != nil {
		return nil, err
	}

	s := &Script{}
	s.AppendOpCode(OP_DUP)
	s.AppendOpCode(OP_HASH160)
	err = s.AppendPushData(publicKeyHashBytes)
	if err != nil {
		return nil, err
	}
	s.AppendOpCode(OP_EQUALVERIFY)
	s.AppendOpCode(OP_CHECKSIG)

	return s, nil
}

// ToString returns hex string of script.
func (s *Script) ToString() string {
	return hex.EncodeToString(*s)
}

// AppendPushData takes data bytes and appends them to the script with proper PUSHDATA prefixes
func (s *Script) AppendPushData(d []byte) error {
	p, err := EncodeParts([][]byte{d})
	if err != nil {
		return err
	}

	*s = append(*s, p...)
	return nil
}

// AppendPushDataHexString takes a hex string and appends them to the script with proper PUSHDATA prefixes
func (s *Script) AppendPushDataHexString(str string) error {
	h, err := hex.DecodeString(str)
	if err != nil {
		return err
	}

	return s.AppendPushData(h)
}

// AppendPushDataString takes a string and appends its UTF-8 encoding to the script with proper PUSHDATA prefixes
func (s *Script) AppendPushDataString(str string) error {
	return s.AppendPushData([]byte(str))
}

// AppendPushDataArray takes an array of data bytes and appends them to the script with proper PUSHDATA prefixes
func (s *Script) AppendPushDataArray(d [][]byte) error {
	p, err := EncodeParts(d)
	if err != nil {
		return err
	}

	*s = append(*s, p...)
	return nil
}

// AppendPushDataStrings takes an array of strings and appends their UTF-8 encoding to the script with proper PUSHDATA prefixes
func (s *Script) AppendPushDataStrings(strs []string) error {
	dataBytes := make([][]byte, 0)
	for _, str := range strs {
		strBytes := []byte(str)
		dataBytes = append(dataBytes, strBytes)
	}

	return s.AppendPushDataArray(dataBytes)
}

// AppendOpCode appends an opcode type to the script
func (s *Script) AppendOpCode(o uint8) {
	*s = append(*s, o)
}

// IsPublicKeyHashOut returns true if this is a pay to pubkey hash output script.
func (s *Script) IsPublicKeyHashOut() bool {
	b := []byte(*s)
	return len(b) == 25 &&
		b[0] == OP_DUP &&
		b[1] == OP_HASH160 &&
		b[2] == 0x14 &&
		b[23] == OP_EQUALVERIFY &&
		b[24] == OP_CHECKSIG
}

// IsPublicKeyOut returns true if this is a public key output script.
func (s *Script) IsPublicKeyOut() bool {
	parts, err := DecodeParts(*s)
	if err != nil {
		return false
	}

	if len(parts) == 2 &&
		len(parts[0]) > 0 &&
		parts[1][0] == OP_CHECKSIG {

		pubkey := parts[0]
		version := pubkey[0]

		if (version == 0x04 || version == 0x06 || version == 0x07) && len(pubkey) == 65 {
			return true
		} else if (version == 0x03 || version == 0x02) && len(pubkey) == 33 {
			return true
		}
	}
	return false
}

// IsScriptHashOut returns true if this is a p2sh output script.
func (s *Script) IsScriptHashOut() bool {
	b := []byte(*s)

	return len(b) == 23 &&
		b[0] == OP_HASH160 &&
		b[1] == 0x14 &&
		b[22] == OP_EQUAL
}

// IsMultisigOut returns true if this is a multisig output script.
func (s *Script) IsMultisigOut() bool {
	parts, err := DecodeParts(*s)
	if err != nil {
		return false
	}

	if len(parts) < 3 {
		return false
	}

	if isSmallIntOp(parts[0][0]) == false {
		return false
	}

	for i := 1; i < len(parts)-2; i++ {
		if len(parts[i]) < 1 {
			return false
		}
	}

	return isSmallIntOp(parts[len(parts)-2][0]) &&
		parts[len(parts)-1][0] == OP_CHECKMULTISIG
}

func isSmallIntOp(opcode byte) bool {
	return opcode == OP_ZERO || (opcode >= OP_ONE && opcode <= OP_16)
}

// GetPublicKeyHash returns a public key hash byte array if the script is a P2PKH script
func (s *Script) GetPublicKeyHash() ([]byte, error) {
	if s == nil || len(*s) == 0 {
		return nil, fmt.Errorf("Script is empty")
	}

	if (*s)[0] != 0x76 || (*s)[1] != 0xa9 {
		return nil, fmt.Errorf("Not a P2PKH")
	}

	parts, err := DecodeParts((*s)[2:])
	if err != nil {
		return nil, err
	}

	return parts[0], nil
}
