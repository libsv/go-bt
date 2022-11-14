package bscript

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/libsv/go-bk/bec"
	"github.com/libsv/go-bk/bip32"
	"github.com/libsv/go-bk/crypto"
)

// ScriptKey types.
const (
	ScriptTypePubKey      = "pubkey"
	ScriptTypePubKeyHash  = "pubkeyhash"
	ScriptTypeNonStandard = "nonstandard"
	ScriptTypeEmpty       = "empty"
	ScriptTypeSecureHash  = "securehash"
	ScriptTypeMultiSig    = "multisig"
	ScriptTypeNullData    = "nulldata"
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
	s := Script{}

	for _, section := range strings.Split(str, " ") {
		if val, ok := opCodeStrings[section]; ok {
			_ = s.AppendOpcodes(val)
		} else {
			if err := s.AppendPushDataHexString(section); err != nil {
				return nil, ErrInvalidOpCode
			}
		}
	}

	return &s, nil
}

// NewP2PKHFromPubKeyEC takes a public key hex string (in
// compressed format) and creates a P2PKH script from it.
func NewP2PKHFromPubKeyEC(pubKey *bec.PublicKey) (*Script, error) {
	return NewP2PKHFromPubKeyBytes(pubKey.SerialiseCompressed())
}

// NewP2PKHFromPubKeyStr takes a public key hex string (in
// compressed format) and creates a P2PKH script from it.
func NewP2PKHFromPubKeyStr(pubKey string) (*Script, error) {
	pubKeyBytes, err := hex.DecodeString(pubKey)
	if err != nil {
		return nil, err
	}
	return NewP2PKHFromPubKeyBytes(pubKeyBytes)
}

// NewP2PKHFromPubKeyBytes takes public key bytes (in
// compressed format) and creates a P2PKH script from it.
func NewP2PKHFromPubKeyBytes(pubKeyBytes []byte) (*Script, error) {
	if len(pubKeyBytes) != 33 {
		return nil, ErrInvalidPKLen
	}
	return NewP2PKHFromPubKeyHash(crypto.Hash160(pubKeyBytes))
}

// NewP2PKHFromPubKeyHash takes a public key hex string (in
// compressed format) and creates a P2PKH script from it.
func NewP2PKHFromPubKeyHash(pubKeyHash []byte) (*Script, error) {
	b := []byte{
		OpDUP,
		OpHASH160,
		OpDATA20,
	}
	b = append(b, pubKeyHash...)
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

	return NewP2PKHFromPubKeyHash(hash)
}

// NewP2PKHFromAddress takes an address
// and creates a P2PKH script from it.
func NewP2PKHFromAddress(addr string) (*Script, error) {
	a, err := NewAddressFromString(addr)
	if err != nil {
		return nil, err
	}

	var publicKeyHashBytes []byte
	if publicKeyHashBytes, err = hex.DecodeString(a.PublicKeyHash); err != nil {
		return nil, err
	}

	s := new(Script)
	_ = s.AppendOpcodes(OpDUP, OpHASH160)
	if err = s.AppendPushData(publicKeyHashBytes); err != nil {
		return nil, err
	}
	_ = s.AppendOpcodes(OpEQUALVERIFY, OpCHECKSIG)

	return s, nil
}

// NewP2PKHFromBip32ExtKey takes a *bip32.ExtendedKey and creates a P2PKH script from it,
// using an internally random generated seed, returning the script and derivation path used.
func NewP2PKHFromBip32ExtKey(privKey *bip32.ExtendedKey) (*Script, string, error) {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return nil, "", err
	}

	derivationPath := bip32.DerivePath(binary.LittleEndian.Uint64(b[:]))
	pubKey, err := privKey.DerivePublicKeyFromPath(derivationPath)
	if err != nil {
		return nil, "", err
	}

	lockingScript, err := NewP2PKHFromPubKeyBytes(pubKey)
	if err != nil {
		return nil, "", err
	}

	return lockingScript, derivationPath, nil
}

// AppendPushData takes data bytes and appends them to the script
// with proper PUSHDATA prefixes
func (s *Script) AppendPushData(d []byte) error {
	p, err := EncodeParts([][]byte{d})
	if err != nil {
		return err
	}

	*s = append(*s, p...)
	return nil
}

// AppendPushDataHexString takes a hex string and appends them to the
// script with proper PUSHDATA prefixes
func (s *Script) AppendPushDataHexString(str string) error {
	h, err := hex.DecodeString(str)
	if err != nil {
		return err
	}

	return s.AppendPushData(h)
}

// AppendPushDataString takes a string and appends its UTF-8 encoding
// to the script with proper PUSHDATA prefixes
func (s *Script) AppendPushDataString(str string) error {
	return s.AppendPushData([]byte(str))
}

// AppendPushDataArray takes an array of data bytes and appends them
// to the script with proper PUSHDATA prefixes
func (s *Script) AppendPushDataArray(d [][]byte) error {
	p, err := EncodeParts(d)
	if err != nil {
		return err
	}

	*s = append(*s, p...)
	return nil
}

// AppendPushDataStrings takes an array of strings and appends their
// UTF-8 encoding to the script with proper PUSHDATA prefixes
func (s *Script) AppendPushDataStrings(pushDataStrings []string) error {
	dataBytes := make([][]byte, 0)
	for _, str := range pushDataStrings {
		strBytes := []byte(str)
		dataBytes = append(dataBytes, strBytes)
	}
	return s.AppendPushDataArray(dataBytes)
}

// AppendOpcodes appends opcodes type to the script.
// This does not support appending OP_PUSHDATA opcodes, so use `Script.AppendPushData` instead.
func (s *Script) AppendOpcodes(oo ...uint8) error {
	for _, o := range oo {
		if OpDATA1 <= o && o <= OpPUSHDATA4 {
			return fmt.Errorf("%w: %s", ErrInvalidOpcodeType, opCodeValues[o])
		}
	}
	*s = append(*s, oo...)
	return nil
}

// String implements the stringer interface and returns the hex string of script.
func (s *Script) String() string {
	return hex.EncodeToString(*s)
}

// ToASM returns the string ASM opcodes of the script.
func (s *Script) ToASM() (string, error) {
	if s == nil || len(*s) == 0 {
		return "", nil
	}
	parts, err := DecodeParts(*s)
	// if err != nil, we will append [error] to the ASM script below (as done in the node).

	data := false
	if len(*s) > 1 && ((*s)[0] == OpRETURN || ((*s)[0] == OpFALSE && (*s)[1] == OpRETURN)) {
		data = true
	}

	var asm strings.Builder

	for _, p := range parts {
		asm.WriteRune(' ')
		if len(p) == 1 {
			if data && p[0] != 0x6a {
				asm.WriteString(fmt.Sprintf("%d", p[0]))
			} else {
				asm.WriteString(opCodeValues[p[0]])
			}
		} else {
			if data && len(p) <= 4 {
				b := make([]byte, 0)
				b = append(b, p...)
				for i := 0; i < 4-len(p); i++ {
					b = append(b, 0)
				}
				asm.WriteString(fmt.Sprintf("%d", binary.LittleEndian.Uint32(b)))
			} else {
				asm.WriteString(hex.EncodeToString(p))
			}
		}
	}

	if err != nil {
		asm.WriteString(" [error]")
	}

	return asm.String()[1:], nil
}

// IsP2PKH returns true if this is a pay to pubkey hash output script.
func (s *Script) IsP2PKH() bool {
	b := []byte(*s)
	return len(b) == 25 &&
		b[0] == OpDUP &&
		b[1] == OpHASH160 &&
		b[2] == OpDATA20 &&
		b[23] == OpEQUALVERIFY &&
		b[24] == OpCHECKSIG
}

// IsP2PK returns true if this is a public key output script.
func (s *Script) IsP2PK() bool {
	parts, err := DecodeParts(*s)
	if err != nil {
		return false
	}

	if len(parts) == 2 && len(parts[0]) > 0 && parts[1][0] == OpCHECKSIG {
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

// IsP2SH returns true if this is a p2sh output script.
// TODO: remove all p2sh stuff from repo
func (s *Script) IsP2SH() bool {
	b := []byte(*s)

	return len(b) == 23 &&
		b[0] == OpHASH160 &&
		b[1] == OpDATA20 &&
		b[22] == OpEQUAL
}

// IsData returns true if this is a data output script. This
// means the script starts with OP_RETURN or OP_FALSE OP_RETURN.
func (s *Script) IsData() bool {
	b := []byte(*s)

	return (len(b) > 0 && b[0] == OpRETURN) ||
		(len(b) > 1 && b[0] == OpFALSE && b[1] == OpRETURN)
}

// IsMultiSigOut returns true if this is a multisig output script.
func (s *Script) IsMultiSigOut() bool {
	parts, err := DecodeParts(*s)
	if err != nil {
		return false
	}

	if len(parts) < 3 {
		return false
	}

	if !isSmallIntOp(parts[0][0]) {
		return false
	}

	for i := 1; i < len(parts)-2; i++ {
		if len(parts[i]) < 1 {
			return false
		}
	}

	return len(parts[len(parts)-2]) > 0 && isSmallIntOp(parts[len(parts)-2][0]) && len(parts[len(parts)-1]) > 0 &&
		parts[len(parts)-1][0] == OpCHECKMULTISIG
}

func isSmallIntOp(opcode byte) bool {
	return opcode == OpZERO || (opcode >= OpONE && opcode <= Op16)
}

// PublicKeyHash returns a public key hash byte array if the script is a P2PKH script.
func (s *Script) PublicKeyHash() ([]byte, error) {
	if s == nil || len(*s) == 0 {
		return nil, ErrEmptyScript
	}

	if (*s)[0] != OpDUP || len(*s) <= 2 || (*s)[1] != OpHASH160 {
		return nil, ErrNotP2PKH
	}

	parts, err := DecodeParts((*s)[2:])
	if err != nil {
		return nil, err
	}

	return parts[0], nil
}

// ScriptType returns the type of script this is as a string.
func (s *Script) ScriptType() string {
	if len(*s) == 0 {
		return ScriptTypeEmpty
	}
	if s.IsP2PKH() {
		return ScriptTypePubKeyHash
	}
	if s.IsP2PK() {
		return ScriptTypePubKey
	}
	if s.IsMultiSigOut() {
		return ScriptTypeMultiSig
	}
	if s.IsData() {
		return ScriptTypeNullData
	}
	return ScriptTypeNonStandard
}

// Addresses will return all addresses found in the script, if any.
func (s *Script) Addresses() ([]string, error) {
	addresses := make([]string, 0)
	if s.IsP2PKH() {
		pkh, err := s.PublicKeyHash()
		if err != nil {
			return nil, err
		}
		a, err := NewAddressFromPublicKeyHash(pkh, true)
		if err != nil {
			return nil, err
		}
		addresses = []string{a.AddressString}
	}
	// TODO: handle multisig, and other outputs
	// https://github.com/libsv/go-bt/issues/6
	return addresses, nil
}

// Equals will compare the script to b and return true if they match.
func (s *Script) Equals(b *Script) bool {
	return bytes.Equal(*s, *b)
}

// EqualsBytes will compare the script to a byte representation of a
// script, b, and return true if they match.
func (s *Script) EqualsBytes(b []byte) bool {
	return bytes.Equal(*s, b)
}

// EqualsHex will compare the script to a hex string h,
// if they match then true is returned otherwise false.
func (s *Script) EqualsHex(h string) bool {
	return s.String() == h
}

// MinPushSize returns the minimum size of a push operation of the given data.
func MinPushSize(bb []byte) int {
	l := len(bb)

	// data length is larger than max supported
	if l > 0xffffffff {
		return 0
	}

	if l == 0 {
		return 1
	}

	if l == 1 {
		// data can be represented as Op1 to Op16, or OpNegate
		if bb[0] <= 16 || bb[0] == 0x81 {
			// OpX
			return 1
		}
		// OP_DATA_1 + data
		return 2
	}

	// OP_DATA_X + data
	if l <= 75 {
		return l + 1
	}
	// OP_PUSHDATA1 + length byte + data
	if l <= 0xff {
		return l + 2
	}
	// OP_PUSHDATA2 + two length bytes + data
	if l <= 0xffff {
		return l + 3
	}

	// OP_PUSHDATA4 + four length bytes + data
	return l + 5
}

// MarshalJSON convert script into json.
func (s *Script) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, s.String())), nil
}

// UnmarshalJSON covert from json into *bscript.Script.
func (s *Script) UnmarshalJSON(bb []byte) error {
	ss, err := NewFromHexString(string(bytes.Trim(bb, `"`)))
	if err != nil {
		return err
	}

	*s = *ss
	return nil
}
