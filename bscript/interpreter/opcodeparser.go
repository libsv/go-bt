package interpreter

import (
	"bytes"
	"encoding/binary"

	"github.com/libsv/go-bt/v2/bscript"
	"github.com/libsv/go-bt/v2/bscript/interpreter/errs"
)

// OpcodeParser parses *bscript.Script into a ParsedScript, and unparsing back
type OpcodeParser interface {
	Parse(*bscript.Script) (ParsedScript, error)
	Unparse(ParsedScript) (*bscript.Script, error)
}

// ParsedScript is a slice of ParsedOp
type ParsedScript []ParsedOpcode

// DefaultOpcodeParser is a standard parser which can be used from zero value.
type DefaultOpcodeParser struct {
	ErrorOnCheckSig bool
}

// ParsedOpcode is a parsed opcode.
type ParsedOpcode struct {
	op   opcode
	Data []byte
}

// Name returns the human readable name for the current opcode.
func (o ParsedOpcode) Name() string {
	return o.op.name
}

// Value returns the byte value of the opcode.
func (o ParsedOpcode) Value() byte {
	return o.op.val
}

// Length returns the data length of the opcode.
func (o ParsedOpcode) Length() int {
	return o.op.length
}

// IsDisabled returns true if the op is disabled.
func (o *ParsedOpcode) IsDisabled() bool {
	switch o.op.val {
	case bscript.Op2MUL, bscript.Op2DIV:
		return true
	default:
		return false
	}
}

// RequiresTx returns true if the op is checksig.
func (o *ParsedOpcode) RequiresTx() bool {
	switch o.op.val {
	case bscript.OpCHECKSIG, bscript.OpCHECKSIGVERIFY,
		bscript.OpCHECKMULTISIG, bscript.OpCHECKMULTISIGVERIFY, bscript.OpCHECKSEQUENCEVERIFY:
		return true
	default:
		return false
	}
}

// AlwaysIllegal returns true if the op is always illegal.
func (o *ParsedOpcode) AlwaysIllegal() bool {
	switch o.op.val {
	case bscript.OpVERIF, bscript.OpVERNOTIF:
		return true
	default:
		return false
	}
}

// IsConditional returns true if the op is a conditional.
func (o *ParsedOpcode) IsConditional() bool {
	switch o.op.val {
	case bscript.OpIF, bscript.OpNOTIF, bscript.OpELSE, bscript.OpENDIF, bscript.OpVERIF, bscript.OpVERNOTIF:
		return true
	default:
		return false
	}
}

// enforceMinimumDataPush checks that the op is pushing only the needed amount of data.
// Errs if not the case.
func (o *ParsedOpcode) enforceMinimumDataPush() error {
	dataLen := len(o.Data)
	if dataLen == 0 && o.op.val != bscript.Op0 {
		return errs.NewError(
			errs.ErrMinimalData,
			"zero length data push is encoded with opcode %s instead of OP_0",
			o.op.name,
		)
	}
	if dataLen == 1 && (1 <= o.Data[0] && o.Data[0] <= 16) && o.op.val != bscript.Op1+o.Data[0]-1 {
		return errs.NewError(
			errs.ErrMinimalData,
			"data push of the value %d encoded with opcode %s instead of OP_%d", o.Data[0], o.op.name, o.Data[0],
		)
	}
	if dataLen == 1 && o.Data[0] == 0x81 && o.op.val != bscript.Op1NEGATE {
		return errs.NewError(
			errs.ErrMinimalData,
			"data push of the value -1 encoded with opcode %s instead of OP_1NEGATE", o.op.name,
		)
	}
	if dataLen <= 75 {
		if int(o.op.val) != dataLen {
			return errs.NewError(
				errs.ErrMinimalData,
				"data push of %d bytes encoded with opcode %s instead of OP_DATA_%d", dataLen, o.op.name, dataLen,
			)
		}
	} else if dataLen <= 255 {
		if o.op.val != bscript.OpPUSHDATA1 {
			return errs.NewError(
				errs.ErrMinimalData,
				"data push of %d bytes encoded with opcode %s instead of OP_PUSHDATA1", dataLen, o.op.name,
			)
		}
	} else if dataLen <= 65535 {
		if o.op.val != bscript.OpPUSHDATA2 {
			return errs.NewError(
				errs.ErrMinimalData,
				"data push of %d bytes encoded with opcode %s instead of OP_PUSHDATA2", dataLen, o.op.name,
			)
		}
	}
	return nil
}

// Parse takes a *bscript.Script and returns a []interpreter.ParsedOp
func (p *DefaultOpcodeParser) Parse(s *bscript.Script) (ParsedScript, error) {
	script := *s
	parsedOps := make([]ParsedOpcode, 0, len(script))

	for i := 0; i < len(script); {
		instruction := script[i]

		parsedOp := ParsedOpcode{op: opcodeArray[instruction]}
		if p.ErrorOnCheckSig && parsedOp.RequiresTx() {
			return nil, errs.NewError(errs.ErrInvalidParams, "tx and previous output must be supplied for checksig")
		}

		switch {
		case parsedOp.op.length == 1:
			i++
		case parsedOp.op.length > 1:
			if len(script[i:]) < parsedOp.op.length {
				return nil, errs.NewError(errs.ErrMalformedPush, "opcode %s required %d bytes, script has %d remaining",
					parsedOp.Name(), parsedOp.op.length, len(script[i:]))
			}
			parsedOp.Data = script[i+1 : i+parsedOp.op.length]
			i += parsedOp.op.length
		case parsedOp.op.length < 0:
			var l uint
			offset := i + 1
			if len(script[offset:]) < -parsedOp.op.length {
				return nil, errs.NewError(errs.ErrMalformedPush, "opcode %s required %d bytes, script has %d remaining",
					parsedOp.Name(), parsedOp.op.length, len(script[offset:]))
			}
			// Next -length bytes are little endian length of data.
			switch parsedOp.op.length {
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
				return nil, errs.NewError(errs.ErrMalformedPush, "invalid opcode length %d", parsedOp.op.length)
			}

			offset += -parsedOp.op.length
			if int(l) > len(script[offset:]) || int(l) < 0 {
				return nil, errs.NewError(errs.ErrMalformedPush, "opcode %s pushes %d bytes, script has %d remaining",
					parsedOp.Name(), l, len(script[offset:]))
			}

			parsedOp.Data = script[offset : offset+int(l)]
			i += 1 - parsedOp.op.length + int(l)
		}

		parsedOps = append(parsedOps, parsedOp)
	}
	return parsedOps, nil
}

// Unparse reverses the action of Parse and returns the
// ParsedScript as a *bscript.Script
func (p *DefaultOpcodeParser) Unparse(pscr ParsedScript) (*bscript.Script, error) {
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
		if op.op.val > bscript.Op16 {
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
		if pop.op.val != opcode {
			retScript = append(retScript, pop)
		}
	}

	return retScript
}

// canonicalPush returns true if the object is either not a push instruction
// or the push instruction contained wherein is matches the canonical form
// or using the smallest instruction to do the job. False otherwise.
func (o ParsedOpcode) canonicalPush() bool {
	opcode := o.op.val
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
func (o *ParsedOpcode) bytes() ([]byte, error) {
	var retbytes []byte
	if o.op.length > 0 {
		retbytes = make([]byte, 1, o.op.length)
	} else {
		retbytes = make([]byte, 1, 1+len(o.Data)-
			o.op.length)
	}

	retbytes[0] = o.op.val
	if o.op.length == 1 {
		if len(o.Data) != 0 {
			return nil, errs.NewError(
				errs.ErrInternal,
				"internal consistency error - parsed opcode %s has data length %d when %d was expected",
				o.Name(), len(o.Data), 0,
			)
		}
		return retbytes, nil
	}
	nbytes := o.op.length
	if o.op.length < 0 {
		l := len(o.Data)
		// tempting just to hardcode to avoid the complexity here.
		switch o.op.length {
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
		return nil, errs.NewError(errs.ErrInternal,
			"internal consistency error - parsed opcode %s has data length %d when %d was expected",
			o.Name(), len(retbytes), nbytes,
		)
	}

	return retbytes, nil
}
