package script

import (
	"encoding/hex"
	"fmt"

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
func NewFromASM(s string) (*Script, error) {
	return &Script{}, nil // TODO:
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
		OpDUP,
		OpHASH160,
		0x14,
	}
	b = append(b, hash...)
	b = append(b, OpEQUALVERIFY)
	b = append(b, OpCHECKSIG)

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
		OpDUP,
		OpHASH160,
		0x14,
	}
	b = append(b, hash...)
	b = append(b, OpEQUALVERIFY)
	b = append(b, OpCHECKSIG)

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
	s.AppendOpCode(OpDUP)
	s.AppendOpCode(OpHASH160)
	err = s.AppendPushData(publicKeyHashBytes)
	if err != nil {
		return nil, err
	}
	s.AppendOpCode(OpEQUALVERIFY)
	s.AppendOpCode(OpCHECKSIG)

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

// AppendPushDataString takes a string and appends them to the script with proper PUSHDATA prefixes
func (s *Script) AppendPushDataString(str string) error {
	err := s.AppendPushData([]byte(str))
	return err
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

// AppendPushDataStrings takes an array of strings and appends them to the script with proper PUSHDATA prefixes
func (s *Script) AppendPushDataStrings(strs []string) error {
	dataBytes := make([][]byte, 0)
	for _, str := range strs {
		strBytes := []byte(str)
		dataBytes = append(dataBytes, strBytes)
	}

	err := s.AppendPushDataArray(dataBytes)
	return err
}

// AppendOpCode appends an opcode type to the script
func (s *Script) AppendOpCode(o uint8) {
	*s = append(*s, o)
}

// IsPublicKeyHashOut returns true if this is a pay to pubkey hash output script.
func (s *Script) IsPublicKeyHashOut() bool {
	b := []byte(*s)
	return len(b) == 25 &&
		b[0] == OpDUP &&
		b[1] == OpHASH160 &&
		b[2] == 0x14 &&
		b[23] == OpEQUALVERIFY &&
		b[24] == OpCHECKSIG
}

// IsPublicKeyOut returns true if this is a public key output script.
func (s *Script) IsPublicKeyOut() bool {
	parts, err := DecodeParts(*s)
	if err != nil {
		return false
	}

	if len(parts) == 2 &&
		len(parts[0]) > 0 &&
		parts[1][0] == OpCHECKSIG {

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
		b[0] == OpHASH160 &&
		b[1] == 0x14 &&
		b[22] == OpEQUAL
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
		parts[len(parts)-1][0] == OpCHECKMULTISIG
}

func isSmallIntOp(opcode byte) bool {
	return opcode == OpZERO || (opcode >= OpONE && opcode <= OpSIXTEEN)
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
