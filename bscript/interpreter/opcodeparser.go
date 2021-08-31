package interpreter

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/libsv/go-bt/v2/bscript"
)

// OpcodeParser parses *bscript.Script into a ParsedScript, and unparsing back
type OpcodeParser interface {
	Parse(*bscript.Script) (ParsedScript, error)
	Unparse(ParsedScript) (*bscript.Script, error)
}

// ParsedScript is a slice of ParsedOp
type ParsedScript []ParsedOp

type parser struct{}

// NewOpcodeParser returns an OpcodeParser
func NewOpcodeParser() OpcodeParser {
	return &parser{}
}

// ParsedOp is a parsed opcode
type ParsedOp struct {
	Op   opcode
	Data []byte
}

// Name returns the human readible name for the current opcode
func (o *ParsedOp) Name() string {
	return o.Op.name
}

// IsDisabled returns true if the op is disabled
func (o *ParsedOp) IsDisabled() bool {
	switch o.Op.val {
	case bscript.Op2MUL, bscript.Op2DIV:
		return true
	default:
		return false
	}
}

// AlwaysIllegal returns true if the op is always illegal
func (o *ParsedOp) AlwaysIllegal() bool {
	switch o.Op.val {
	case bscript.OpVERIF, bscript.OpVERNOTIF:
		return true
	default:
		return false
	}
}

// IsConditional returns true if the op is a conditional
func (o *ParsedOp) IsConditional() bool {
	switch o.Op.val {
	case bscript.OpIF, bscript.OpNOTIF, bscript.OpELSE, bscript.OpENDIF, bscript.OpVERIF, bscript.OpVERNOTIF:
		return true
	default:
		return false
	}
}

// EnforceMinimumDataPush checks that the op is pushing only the needed amount of data.
// Errs if not the case.
func (o *ParsedOp) EnforceMinimumDataPush() error {
	dataLen := len(o.Data)
	if dataLen == 0 && o.Op.val != bscript.Op0 {
		return NewErrMinimalDataPush(o.Op, opcodeArray[bscript.Op0], 0)
	}
	if dataLen == 1 && (1 <= o.Data[0] && o.Data[0] <= 16) && o.Op.val != bscript.Op1+o.Data[0]-1 {
		return NewErrMinimalDataPush(o.Op, opcodeArray[bscript.Op1+o.Data[0]-1], int(o.Data[0]))
	}
	if dataLen == 1 && o.Data[0] == 0x81 && o.Op.val != bscript.Op1NEGATE {
		return NewErrMinimalDataPush(o.Op, opcodeArray[bscript.Op1NEGATE], -1)
	}
	if dataLen <= 75 {
		if int(o.Op.val) != dataLen {
			return NewErrMinimalDataPush(o.Op, opcodeArray[byte(dataLen)], dataLen)
		}
	} else if dataLen <= 255 {
		if o.Op.val != bscript.OpPUSHDATA1 {
			return NewErrMinimalDataPush(o.Op, opcodeArray[bscript.OpPUSHDATA1], dataLen)
		}
	} else if dataLen <= 65535 {
		if o.Op.val != bscript.OpPUSHDATA2 {
			return NewErrMinimalDataPush(o.Op, opcodeArray[bscript.OpPUSHDATA2], dataLen)
		}
	}
	return nil
}

func (p *parser) Parse(s *bscript.Script) (ParsedScript, error) {
	script := *s
	parsedOps := make([]ParsedOp, 0, len(script))

	for i := 0; i < len(script); {
		instruction := script[i]

		parsedOp := ParsedOp{Op: opcodeArray[instruction]}

		switch {
		case parsedOp.Op.length == 1:
			i++
		case parsedOp.Op.length > 1:
			if len(script[i:]) < parsedOp.Op.length {
				return nil, scriptError(ErrMalformedPush, fmt.Sprintf("opcode %s required %d bytes, script has %d remaining",
					parsedOp.Name(), parsedOp.Op.length, len(script[i:])))
			}
			parsedOp.Data = script[i+1 : i+parsedOp.Op.length]
			i += parsedOp.Op.length
		case parsedOp.Op.length < 0:
			var l uint
			offset := i + 1
			if len(script[offset:]) < -parsedOp.Op.length {
				return nil, scriptError(ErrMalformedPush, fmt.Sprintf("opcode %s required %d bytes, script has %d remaining",
					parsedOp.Name(), parsedOp.Op.length, len(script[offset:])))
			}
			// Next -length bytes are little endian length of data.
			switch parsedOp.Op.length {
			case -1:
				l = uint(script[offset])
			case -2:
				l = ((uint(script[offset+1]) << 8) |
					uint(script[offset]))
			case -4:
				l = ((uint(script[offset+3]) << 24) |
					(uint(script[offset+2]) << 16) |
					(uint(script[offset+1]) << 8) |
					uint(script[offset]))
			default:
				return nil, scriptError(ErrMalformedPush, fmt.Sprintf("invalid opcode length %d", parsedOp.Op.length))
			}

			offset += -parsedOp.Op.length
			if int(l) > len(script[offset:]) || int(l) < 0 {
				return nil, scriptError(ErrMalformedPush, fmt.Sprintf("opcode %s pushes %d bytes, script has %d remaining",
					parsedOp.Name(), l, len(script[offset:])))
			}

			parsedOp.Data = script[offset : offset+int(l)]
			i += 1 - parsedOp.Op.length + int(l)
		}

		parsedOps = append(parsedOps, parsedOp)
	}
	return parsedOps, nil
}

// unparseScript reversed the action of parseScript and returns the
// parsedOpcodes as a list of bytes
func (p *parser) Unparse(pscr ParsedScript) (*bscript.Script, error) {
	script := make(bscript.Script, 0, len(pscr))
	for _, pop := range pscr {
		b, err := pop.bytes()
		if err != nil {
			return nil, err
		}
		script = append(script, b...)
	}
	return &script, nil
}

// IsPushOnly returns true if the ParsedScript only contains push commands
func (p ParsedScript) IsPushOnly() bool {
	for _, op := range p {
		if op.Op.val > bscript.Op16 {
			return false
		}
	}

	return true
}

// removeOpcodeByData will return the script minus any opcodes that would push
// the passed data to the stack.
func (p ParsedScript) removeOpcodeByData(data []byte) ParsedScript {
	retScript := make(ParsedScript, 0, len(p))
	for _, pop := range p {
		if !pop.canonicalPush() || !bytes.Contains(pop.Data, data) {
			retScript = append(retScript, pop)
		}
	}

	return retScript
}

func (p ParsedScript) removeOpcode(opcode byte) ParsedScript {
	retScript := make(ParsedScript, 0, len(p))
	for _, pop := range p {
		if pop.Op.val != opcode {
			retScript = append(retScript, pop)
		}
	}

	return retScript
}

// canonicalPush returns true if the object is either not a push instruction
// or the push instruction contained wherein is matches the canonical form
// or using the smallest instruction to do the job. False otherwise.
func (o ParsedOp) canonicalPush() bool {
	opcode := o.Op.val
	data := o.Data
	dataLen := len(o.Data)
	if opcode > bscript.Op16 {
		return true
	}

	if opcode < bscript.OpPUSHDATA1 && opcode > bscript.Op0 && (dataLen == 1 && data[0] <= 16) {
		return false
	}
	if opcode == bscript.OpPUSHDATA1 && dataLen < int(bscript.OpPUSHDATA1) {
		return false
	}
	if opcode == bscript.OpPUSHDATA2 && dataLen <= 0xff {
		return false
	}
	if opcode == bscript.OpPUSHDATA4 && dataLen <= 0xffff {
		return false
	}
	return true
}

// bytes returns any data associated with the opcode encoded as it would be in
// a script.  This is used for unparsing scripts from parsed opcodes.
func (o *ParsedOp) bytes() ([]byte, error) {
	var retbytes []byte
	if o.Op.length > 0 {
		retbytes = make([]byte, 1, o.Op.length)
	} else {
		retbytes = make([]byte, 1, 1+len(o.Data)-
			o.Op.length)
	}

	retbytes[0] = o.Op.val
	if o.Op.length == 1 {
		if len(o.Data) != 0 {
			str := fmt.Sprintf("internal consistency error - "+
				"parsed opcode %s has data length %d when %d "+
				"was expected", o.Name(), len(o.Data),
				0)
			return nil, scriptError(ErrInternal, str)
		}
		return retbytes, nil
	}
	nbytes := o.Op.length
	if o.Op.length < 0 {
		l := len(o.Data)
		// tempting just to hardcode to avoid the complexity here.
		switch o.Op.length {
		case -1:
			retbytes = append(retbytes, byte(l))
			nbytes = int(retbytes[1]) + len(retbytes)
		case -2:
			retbytes = append(retbytes, byte(l&0xff),
				byte(l>>8&0xff))
			nbytes = int(binary.LittleEndian.Uint16(retbytes[1:])) +
				len(retbytes)
		case -4:
			retbytes = append(retbytes, byte(l&0xff),
				byte((l>>8)&0xff), byte((l>>16)&0xff),
				byte((l>>24)&0xff))
			nbytes = int(binary.LittleEndian.Uint32(retbytes[1:])) +
				len(retbytes)
		}
	}

	retbytes = append(retbytes, o.Data...)

	if len(retbytes) != nbytes {
		str := fmt.Sprintf("internal consistency error - "+
			"parsed opcode %s has data length %d when %d was "+
			"expected", o.Name(), len(retbytes), nbytes)
		return nil, scriptError(ErrInternal, str)
	}

	return retbytes, nil
}
