package transaction

import (
	"encoding/hex"
	"fmt"

	"bitbucket.org/simon_ordish/cryptolib"
)

// Script type
type Script []byte

// NewScript creates a new Script.
func NewScript() *Script {
	s := Script(make([]byte, 0))
	return &s
}

// NewScriptFromString creates a new script from a hex encoded string.
func NewScriptFromString(s string) *Script {
	b, _ := hex.DecodeString(s)
	return NewScriptFromBytes(b)
}

// NewScriptFromBytes wraps a byte slice with the Script type.
func NewScriptFromBytes(b []byte) *Script {
	s := Script(b)
	return &s
}

// ToString returns hex string of script.
func (s *Script) ToString() string {
	return hex.EncodeToString(*s)
}

// AppendPushDataToScript takes an array of data bytes and appends them to the script with proper PUSHDATA prefixes
func (s *Script) AppendPushDataToScript(d [][]byte) error {
	p, err := cryptolib.EncodeParts(d)
	if err != nil {
		return err
	}

	*s = append(*s, p...)
	return nil
}

// AppendPushDataStringsToScript takes an array of strings and appends them to the script with proper PUSHDATA prefixes
func (s *Script) AppendPushDataStringsToScript(strs []string) error {
	dataBytes := make([][]byte, 0)
	for _, str := range strs {
		strBytes := []byte(str)
		dataBytes = append(dataBytes, strBytes)
	}

	err := s.AppendPushDataToScript(dataBytes)
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
		b[0] == cryptolib.OpDUP &&
		b[1] == cryptolib.OpHASH160 &&
		b[2] == 0x14 &&
		b[23] == cryptolib.OpEQUALVERIFY &&
		b[24] == cryptolib.OpCHECKSIG
}

// IsPublicKeyOut returns true if this is a public key output script.
func (s *Script) IsPublicKeyOut() bool {
	parts, err := cryptolib.DecodeParts(*s)
	if err != nil {
		return false
	}

	if len(parts) == 2 &&
		len(parts[0]) > 0 &&
		parts[1][0] == cryptolib.OpCHECKSIG {

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
		b[0] == cryptolib.OpHASH160 &&
		b[1] == 0x14 &&
		b[22] == cryptolib.OpEQUAL
}

// IsMultisigOut returns true if this is a multisig output script.
func (s *Script) IsMultisigOut() bool {
	parts, err := cryptolib.DecodeParts(*s)
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
		parts[len(parts)-1][0] == cryptolib.OpCHECKMULTISIG
}

func isSmallIntOp(opcode byte) bool {
	return opcode == cryptolib.OpZERO || (opcode >= cryptolib.OpONE && opcode <= cryptolib.OpSIXTEEN)
}

// GetPublicKeyHash returns a public key hash byte array if the script is a P2PKH script
func (s *Script) GetPublicKeyHash() ([]byte, error) {
	if s == nil || len(*s) == 0 {
		return nil, fmt.Errorf("Script is empty")
	}

	if (*s)[0] != 0x76 || (*s)[1] != 0xa9 {
		return nil, fmt.Errorf("Not a P2PKH")
	}

	parts, err := cryptolib.DecodeParts((*s)[2:])
	if err != nil {
		return nil, err
	}

	return parts[0], nil
}
