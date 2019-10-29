package transaction

import "bitbucket.org/simon_ordish/cryptolib"

// Script type
type Script []byte

// NewScript wraps a byte slice with the Script type
func NewScript(b []byte) *Script {
	s := Script(b)
	return &s
}

// IsPublicKeyHashOut returns true if this is a pay to pubkey hash output script
func (s *Script) IsPublicKeyHashOut() bool {
	b := []byte(*s)
	return len(b) == 25 &&
		b[0] == opDUP &&
		b[1] == opHASH160 &&
		b[2] == 0x14 &&
		b[23] == opEQUALVERIFY &&
		b[24] == opCHECKSIG
}

// IsPublicKeyOut returns true if this is a public key output script
func (s *Script) IsPublicKeyOut() bool {
	parts, err := cryptolib.DecodeParts(*s)
	if err != nil {
		return false
	}

	if len(parts) == 2 &&
		len(parts[0]) > 0 &&
		parts[1][0] == opCHECKSIG {

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

// IsScriptHashOut returns true if this is a p2sh output script
func (s *Script) IsScriptHashOut() bool {
	b := []byte(*s)

	return len(b) == 23 &&
		b[0] == opHASH160 &&
		b[1] == 0x14 &&
		b[22] == opEQUAL
}

// IsMultisigOut returns true if this is a multisig output script
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
		parts[len(parts)-1][0] == opCHECKMULTISIG
}

func isSmallIntOp(opcode byte) bool {
	return opcode == opZERO || (opcode >= opONE && opcode <= opSIXTEEN)
}

func (s *Script) getPublicKeyHash() []byte {
	parts, err := cryptolib.DecodeParts(*s)
	if err != nil {
		return nil
	}

	if len(parts) < 3 {
		return nil
	}

	return parts[2]
}
