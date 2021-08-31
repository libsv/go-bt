package interpreter

import (
	"bytes"
	"crypto/sha1" //nolint:gosec // OP_SHA1 support requires this
	"crypto/sha256"
	"fmt"
	"hash"

	"github.com/libsv/go-bk/bec"
	"github.com/libsv/go-bk/crypto"
	"github.com/libsv/go-bt/v2"
	"github.com/libsv/go-bt/v2/bscript"
	"github.com/libsv/go-bt/v2/sighash"
	"golang.org/x/crypto/ripemd160"
)

// Conditional execution constants.
const (
	OpCondFalse = 0
	OpCondTrue  = 1
	OpCondSkip  = 2
)

type opcode struct {
	val    byte
	name   string
	length int
	exec   func(*ParsedOp, *thread) error
}

func (o opcode) Name() string {
	return o.name
}

// opcodeArray associates an opcode with its respective function, and defines them in order as to
// be correctly placed in an array
var opcodeArray = [256]opcode{
	// Data push opcodes.
	bscript.OpFALSE:     {bscript.OpFALSE, "OP_0", 1, opcodeFalse},
	bscript.OpDATA1:     {bscript.OpDATA1, "OP_DATA_1", 2, opcodePushData},
	bscript.OpDATA2:     {bscript.OpDATA2, "OP_DATA_2", 3, opcodePushData},
	bscript.OpDATA3:     {bscript.OpDATA3, "OP_DATA_3", 4, opcodePushData},
	bscript.OpDATA4:     {bscript.OpDATA4, "OP_DATA_4", 5, opcodePushData},
	bscript.OpDATA5:     {bscript.OpDATA5, "OP_DATA_5", 6, opcodePushData},
	bscript.OpDATA6:     {bscript.OpDATA6, "OP_DATA_6", 7, opcodePushData},
	bscript.OpDATA7:     {bscript.OpDATA7, "OP_DATA_7", 8, opcodePushData},
	bscript.OpDATA8:     {bscript.OpDATA8, "OP_DATA_8", 9, opcodePushData},
	bscript.OpDATA9:     {bscript.OpDATA9, "OP_DATA_9", 10, opcodePushData},
	bscript.OpDATA10:    {bscript.OpDATA10, "OP_DATA_10", 11, opcodePushData},
	bscript.OpDATA11:    {bscript.OpDATA11, "OP_DATA_11", 12, opcodePushData},
	bscript.OpDATA12:    {bscript.OpDATA12, "OP_DATA_12", 13, opcodePushData},
	bscript.OpDATA13:    {bscript.OpDATA13, "OP_DATA_13", 14, opcodePushData},
	bscript.OpDATA14:    {bscript.OpDATA14, "OP_DATA_14", 15, opcodePushData},
	bscript.OpDATA15:    {bscript.OpDATA15, "OP_DATA_15", 16, opcodePushData},
	bscript.OpDATA16:    {bscript.OpDATA16, "OP_DATA_16", 17, opcodePushData},
	bscript.OpDATA17:    {bscript.OpDATA17, "OP_DATA_17", 18, opcodePushData},
	bscript.OpDATA18:    {bscript.OpDATA18, "OP_DATA_18", 19, opcodePushData},
	bscript.OpDATA19:    {bscript.OpDATA19, "OP_DATA_19", 20, opcodePushData},
	bscript.OpDATA20:    {bscript.OpDATA20, "OP_DATA_20", 21, opcodePushData},
	bscript.OpDATA21:    {bscript.OpDATA21, "OP_DATA_21", 22, opcodePushData},
	bscript.OpDATA22:    {bscript.OpDATA22, "OP_DATA_22", 23, opcodePushData},
	bscript.OpDATA23:    {bscript.OpDATA23, "OP_DATA_23", 24, opcodePushData},
	bscript.OpDATA24:    {bscript.OpDATA24, "OP_DATA_24", 25, opcodePushData},
	bscript.OpDATA25:    {bscript.OpDATA25, "OP_DATA_25", 26, opcodePushData},
	bscript.OpDATA26:    {bscript.OpDATA26, "OP_DATA_26", 27, opcodePushData},
	bscript.OpDATA27:    {bscript.OpDATA27, "OP_DATA_27", 28, opcodePushData},
	bscript.OpDATA28:    {bscript.OpDATA28, "OP_DATA_28", 29, opcodePushData},
	bscript.OpDATA29:    {bscript.OpDATA29, "OP_DATA_29", 30, opcodePushData},
	bscript.OpDATA30:    {bscript.OpDATA30, "OP_DATA_30", 31, opcodePushData},
	bscript.OpDATA31:    {bscript.OpDATA31, "OP_DATA_31", 32, opcodePushData},
	bscript.OpDATA32:    {bscript.OpDATA32, "OP_DATA_32", 33, opcodePushData},
	bscript.OpDATA33:    {bscript.OpDATA33, "OP_DATA_33", 34, opcodePushData},
	bscript.OpDATA34:    {bscript.OpDATA34, "OP_DATA_34", 35, opcodePushData},
	bscript.OpDATA35:    {bscript.OpDATA35, "OP_DATA_35", 36, opcodePushData},
	bscript.OpDATA36:    {bscript.OpDATA36, "OP_DATA_36", 37, opcodePushData},
	bscript.OpDATA37:    {bscript.OpDATA37, "OP_DATA_37", 38, opcodePushData},
	bscript.OpDATA38:    {bscript.OpDATA38, "OP_DATA_38", 39, opcodePushData},
	bscript.OpDATA39:    {bscript.OpDATA39, "OP_DATA_39", 40, opcodePushData},
	bscript.OpDATA40:    {bscript.OpDATA40, "OP_DATA_40", 41, opcodePushData},
	bscript.OpDATA41:    {bscript.OpDATA41, "OP_DATA_41", 42, opcodePushData},
	bscript.OpDATA42:    {bscript.OpDATA42, "OP_DATA_42", 43, opcodePushData},
	bscript.OpDATA43:    {bscript.OpDATA43, "OP_DATA_43", 44, opcodePushData},
	bscript.OpDATA44:    {bscript.OpDATA44, "OP_DATA_44", 45, opcodePushData},
	bscript.OpDATA45:    {bscript.OpDATA45, "OP_DATA_45", 46, opcodePushData},
	bscript.OpDATA46:    {bscript.OpDATA46, "OP_DATA_46", 47, opcodePushData},
	bscript.OpDATA47:    {bscript.OpDATA47, "OP_DATA_47", 48, opcodePushData},
	bscript.OpDATA48:    {bscript.OpDATA48, "OP_DATA_48", 49, opcodePushData},
	bscript.OpDATA49:    {bscript.OpDATA49, "OP_DATA_49", 50, opcodePushData},
	bscript.OpDATA50:    {bscript.OpDATA50, "OP_DATA_50", 51, opcodePushData},
	bscript.OpDATA51:    {bscript.OpDATA51, "OP_DATA_51", 52, opcodePushData},
	bscript.OpDATA52:    {bscript.OpDATA52, "OP_DATA_52", 53, opcodePushData},
	bscript.OpDATA53:    {bscript.OpDATA53, "OP_DATA_53", 54, opcodePushData},
	bscript.OpDATA54:    {bscript.OpDATA54, "OP_DATA_54", 55, opcodePushData},
	bscript.OpDATA55:    {bscript.OpDATA55, "OP_DATA_55", 56, opcodePushData},
	bscript.OpDATA56:    {bscript.OpDATA56, "OP_DATA_56", 57, opcodePushData},
	bscript.OpDATA57:    {bscript.OpDATA57, "OP_DATA_57", 58, opcodePushData},
	bscript.OpDATA58:    {bscript.OpDATA58, "OP_DATA_58", 59, opcodePushData},
	bscript.OpDATA59:    {bscript.OpDATA59, "OP_DATA_59", 60, opcodePushData},
	bscript.OpDATA60:    {bscript.OpDATA60, "OP_DATA_60", 61, opcodePushData},
	bscript.OpDATA61:    {bscript.OpDATA61, "OP_DATA_61", 62, opcodePushData},
	bscript.OpDATA62:    {bscript.OpDATA62, "OP_DATA_62", 63, opcodePushData},
	bscript.OpDATA63:    {bscript.OpDATA63, "OP_DATA_63", 64, opcodePushData},
	bscript.OpDATA64:    {bscript.OpDATA64, "OP_DATA_64", 65, opcodePushData},
	bscript.OpDATA65:    {bscript.OpDATA65, "OP_DATA_65", 66, opcodePushData},
	bscript.OpDATA66:    {bscript.OpDATA66, "OP_DATA_66", 67, opcodePushData},
	bscript.OpDATA67:    {bscript.OpDATA67, "OP_DATA_67", 68, opcodePushData},
	bscript.OpDATA68:    {bscript.OpDATA68, "OP_DATA_68", 69, opcodePushData},
	bscript.OpDATA69:    {bscript.OpDATA69, "OP_DATA_69", 70, opcodePushData},
	bscript.OpDATA70:    {bscript.OpDATA70, "OP_DATA_70", 71, opcodePushData},
	bscript.OpDATA71:    {bscript.OpDATA71, "OP_DATA_71", 72, opcodePushData},
	bscript.OpDATA72:    {bscript.OpDATA72, "OP_DATA_72", 73, opcodePushData},
	bscript.OpDATA73:    {bscript.OpDATA73, "OP_DATA_73", 74, opcodePushData},
	bscript.OpDATA74:    {bscript.OpDATA74, "OP_DATA_74", 75, opcodePushData},
	bscript.OpDATA75:    {bscript.OpDATA75, "OP_DATA_75", 76, opcodePushData},
	bscript.OpPUSHDATA1: {bscript.OpPUSHDATA1, "OP_PUSHDATA1", -1, opcodePushData},
	bscript.OpPUSHDATA2: {bscript.OpPUSHDATA2, "OP_PUSHDATA2", -2, opcodePushData},
	bscript.OpPUSHDATA4: {bscript.OpPUSHDATA4, "OP_PUSHDATA4", -4, opcodePushData},
	bscript.Op1NEGATE:   {bscript.Op1NEGATE, "OP_1NEGATE", 1, opcode1Negate},
	bscript.OpRESERVED:  {bscript.OpRESERVED, "OP_RESERVED", 1, opcodeReserved},
	bscript.OpTRUE:      {bscript.OpTRUE, "OP_1", 1, opcodeN},
	bscript.Op2:         {bscript.Op2, "OP_2", 1, opcodeN},
	bscript.Op3:         {bscript.Op3, "OP_3", 1, opcodeN},
	bscript.Op4:         {bscript.Op4, "OP_4", 1, opcodeN},
	bscript.Op5:         {bscript.Op5, "OP_5", 1, opcodeN},
	bscript.Op6:         {bscript.Op6, "OP_6", 1, opcodeN},
	bscript.Op7:         {bscript.Op7, "OP_7", 1, opcodeN},
	bscript.Op8:         {bscript.Op8, "OP_8", 1, opcodeN},
	bscript.Op9:         {bscript.Op9, "OP_9", 1, opcodeN},
	bscript.Op10:        {bscript.Op10, "OP_10", 1, opcodeN},
	bscript.Op11:        {bscript.Op11, "OP_11", 1, opcodeN},
	bscript.Op12:        {bscript.Op12, "OP_12", 1, opcodeN},
	bscript.Op13:        {bscript.Op13, "OP_13", 1, opcodeN},
	bscript.Op14:        {bscript.Op14, "OP_14", 1, opcodeN},
	bscript.Op15:        {bscript.Op15, "OP_15", 1, opcodeN},
	bscript.Op16:        {bscript.Op16, "OP_16", 1, opcodeN},

	// Control opcodes.
	bscript.OpNOP:                 {bscript.OpNOP, "OP_NOP", 1, opcodeNop},
	bscript.OpVER:                 {bscript.OpVER, "OP_VER", 1, opcodeReserved},
	bscript.OpIF:                  {bscript.OpIF, "OP_IF", 1, opcodeIf},
	bscript.OpNOTIF:               {bscript.OpNOTIF, "OP_NOTIF", 1, opcodeNotIf},
	bscript.OpVERIF:               {bscript.OpVERIF, "OP_VERIF", 1, opcodeVerConditional},
	bscript.OpVERNOTIF:            {bscript.OpVERNOTIF, "OP_VERNOTIF", 1, opcodeVerConditional},
	bscript.OpELSE:                {bscript.OpELSE, "OP_ELSE", 1, opcodeElse},
	bscript.OpENDIF:               {bscript.OpENDIF, "OP_ENDIF", 1, opcodeEndif},
	bscript.OpVERIFY:              {bscript.OpVERIFY, "OP_VERIFY", 1, opcodeVerify},
	bscript.OpRETURN:              {bscript.OpRETURN, "OP_RETURN", 1, opcodeReturn},
	bscript.OpCHECKLOCKTIMEVERIFY: {bscript.OpCHECKLOCKTIMEVERIFY, "OP_CHECKLOCKTIMEVERIFY", 1, opcodeCheckLockTimeVerify},
	bscript.OpCHECKSEQUENCEVERIFY: {bscript.OpCHECKSEQUENCEVERIFY, "OP_CHECKSEQUENCEVERIFY", 1, opcodeCheckSequenceVerify},

	// Stack opcodes.
	bscript.OpTOALTSTACK:   {bscript.OpTOALTSTACK, "OP_TOALTSTACK", 1, opcodeToAltStack},
	bscript.OpFROMALTSTACK: {bscript.OpFROMALTSTACK, "OP_FROMALTSTACK", 1, opcodeFromAltStack},
	bscript.Op2DROP:        {bscript.Op2DROP, "OP_2DROP", 1, opcode2Drop},
	bscript.Op2DUP:         {bscript.Op2DUP, "OP_2DUP", 1, opcode2Dup},
	bscript.Op3DUP:         {bscript.Op3DUP, "OP_3DUP", 1, opcode3Dup},
	bscript.Op2OVER:        {bscript.Op2OVER, "OP_2OVER", 1, opcode2Over},
	bscript.Op2ROT:         {bscript.Op2ROT, "OP_2ROT", 1, opcode2Rot},
	bscript.Op2SWAP:        {bscript.Op2SWAP, "OP_2SWAP", 1, opcode2Swap},
	bscript.OpIFDUP:        {bscript.OpIFDUP, "OP_IFDUP", 1, opcodeIfDup},
	bscript.OpDEPTH:        {bscript.OpDEPTH, "OP_DEPTH", 1, opcodeDepth},
	bscript.OpDROP:         {bscript.OpDROP, "OP_DROP", 1, opcodeDrop},
	bscript.OpDUP:          {bscript.OpDUP, "OP_DUP", 1, opcodeDup},
	bscript.OpNIP:          {bscript.OpNIP, "OP_NIP", 1, opcodeNip},
	bscript.OpOVER:         {bscript.OpOVER, "OP_OVER", 1, opcodeOver},
	bscript.OpPICK:         {bscript.OpPICK, "OP_PICK", 1, opcodePick},
	bscript.OpROLL:         {bscript.OpROLL, "OP_ROLL", 1, opcodeRoll},
	bscript.OpROT:          {bscript.OpROT, "OP_ROT", 1, opcodeRot},
	bscript.OpSWAP:         {bscript.OpSWAP, "OP_SWAP", 1, opcodeSwap},
	bscript.OpTUCK:         {bscript.OpTUCK, "OP_TUCK", 1, opcodeTuck},

	// Splice opcodes.
	bscript.OpCAT:     {bscript.OpCAT, "OP_CAT", 1, opcodeCat},
	bscript.OpSPLIT:   {bscript.OpSPLIT, "OP_SPLIT", 1, opcodeSplit},
	bscript.OpNUM2BIN: {bscript.OpNUM2BIN, "OP_NUM2BIN", 1, opcodeNum2bin},
	bscript.OpBIN2NUM: {bscript.OpBIN2NUM, "OP_BIN2NUM", 1, opcodeBin2num},
	bscript.OpSIZE:    {bscript.OpSIZE, "OP_SIZE", 1, opcodeSize},

	// Bitwise logic opcodes.
	bscript.OpINVERT:      {bscript.OpINVERT, "OP_INVERT", 1, opcodeInvert},
	bscript.OpAND:         {bscript.OpAND, "OP_AND", 1, opcodeAnd},
	bscript.OpOR:          {bscript.OpOR, "OP_OR", 1, opcodeOr},
	bscript.OpXOR:         {bscript.OpXOR, "OP_XOR", 1, opcodeXor},
	bscript.OpEQUAL:       {bscript.OpEQUAL, "OP_EQUAL", 1, opcodeEqual},
	bscript.OpEQUALVERIFY: {bscript.OpEQUALVERIFY, "OP_EQUALVERIFY", 1, opcodeEqualVerify},
	bscript.OpRESERVED1:   {bscript.OpRESERVED1, "OP_RESERVED1", 1, opcodeReserved},
	bscript.OpRESERVED2:   {bscript.OpRESERVED2, "OP_RESERVED2", 1, opcodeReserved},

	// Numeric related opcodes.
	bscript.Op1ADD:               {bscript.Op1ADD, "OP_1ADD", 1, opcode1Add},
	bscript.Op1SUB:               {bscript.Op1SUB, "OP_1SUB", 1, opcode1Sub},
	bscript.Op2MUL:               {bscript.Op2MUL, "OP_2MUL", 1, opcodeDisabled},
	bscript.Op2DIV:               {bscript.Op2DIV, "OP_2DIV", 1, opcodeDisabled},
	bscript.OpNEGATE:             {bscript.OpNEGATE, "OP_NEGATE", 1, opcodeNegate},
	bscript.OpABS:                {bscript.OpABS, "OP_ABS", 1, opcodeAbs},
	bscript.OpNOT:                {bscript.OpNOT, "OP_NOT", 1, opcodeNot},
	bscript.Op0NOTEQUAL:          {bscript.Op0NOTEQUAL, "OP_0NOTEQUAL", 1, opcode0NotEqual},
	bscript.OpADD:                {bscript.OpADD, "OP_ADD", 1, opcodeAdd},
	bscript.OpSUB:                {bscript.OpSUB, "OP_SUB", 1, opcodeSub},
	bscript.OpMUL:                {bscript.OpMUL, "OP_MUL", 1, opcodeMul},
	bscript.OpDIV:                {bscript.OpDIV, "OP_DIV", 1, opcodeDiv},
	bscript.OpMOD:                {bscript.OpMOD, "OP_MOD", 1, opcodeMod},
	bscript.OpLSHIFT:             {bscript.OpLSHIFT, "OP_LSHIFT", 1, opcodeLShift},
	bscript.OpRSHIFT:             {bscript.OpRSHIFT, "OP_RSHIFT", 1, opcodeRShift},
	bscript.OpBOOLAND:            {bscript.OpBOOLAND, "OP_BOOLAND", 1, opcodeBoolAnd},
	bscript.OpBOOLOR:             {bscript.OpBOOLOR, "OP_BOOLOR", 1, opcodeBoolOr},
	bscript.OpNUMEQUAL:           {bscript.OpNUMEQUAL, "OP_NUMEQUAL", 1, opcodeNumEqual},
	bscript.OpNUMEQUALVERIFY:     {bscript.OpNUMEQUALVERIFY, "OP_NUMEQUALVERIFY", 1, opcodeNumEqualVerify},
	bscript.OpNUMNOTEQUAL:        {bscript.OpNUMNOTEQUAL, "OP_NUMNOTEQUAL", 1, opcodeNumNotEqual},
	bscript.OpLESSTHAN:           {bscript.OpLESSTHAN, "OP_LESSTHAN", 1, opcodeLessThan},
	bscript.OpGREATERTHAN:        {bscript.OpGREATERTHAN, "OP_GREATERTHAN", 1, opcodeGreaterThan},
	bscript.OpLESSTHANOREQUAL:    {bscript.OpLESSTHANOREQUAL, "OP_LESSTHANOREQUAL", 1, opcodeLessThanOrEqual},
	bscript.OpGREATERTHANOREQUAL: {bscript.OpGREATERTHANOREQUAL, "OP_GREATERTHANOREQUAL", 1, opcodeGreaterThanOrEqual},
	bscript.OpMIN:                {bscript.OpMIN, "OP_MIN", 1, opcodeMin},
	bscript.OpMAX:                {bscript.OpMAX, "OP_MAX", 1, opcodeMax},
	bscript.OpWITHIN:             {bscript.OpWITHIN, "OP_WITHIN", 1, opcodeWithin},

	// Crypto opcodes.
	bscript.OpRIPEMD160:           {bscript.OpRIPEMD160, "OP_RIPEMD160", 1, opcodeRipemd160},
	bscript.OpSHA1:                {bscript.OpSHA1, "OP_SHA1", 1, opcodeSha1},
	bscript.OpSHA256:              {bscript.OpSHA256, "OP_SHA256", 1, opcodeSha256},
	bscript.OpHASH160:             {bscript.OpHASH160, "OP_HASH160", 1, opcodeHash160},
	bscript.OpHASH256:             {bscript.OpHASH256, "OP_HASH256", 1, opcodeHash256},
	bscript.OpCODESEPARATOR:       {bscript.OpCODESEPARATOR, "OP_CODESEPARATOR", 1, opcodeCodeSeparator},
	bscript.OpCHECKSIG:            {bscript.OpCHECKSIG, "OP_CHECKSIG", 1, opcodeCheckSig},
	bscript.OpCHECKSIGVERIFY:      {bscript.OpCHECKSIGVERIFY, "OP_CHECKSIGVERIFY", 1, opcodeCheckSigVerify},
	bscript.OpCHECKMULTISIG:       {bscript.OpCHECKMULTISIG, "OP_CHECKMULTISIG", 1, opcodeCheckMultiSig},
	bscript.OpCHECKMULTISIGVERIFY: {bscript.OpCHECKMULTISIGVERIFY, "OP_CHECKMULTISIGVERIFY", 1, opcodeCheckMultiSigVerify},

	// Reserved opcodes.
	bscript.OpNOP1:  {bscript.OpNOP1, "OP_NOP1", 1, opcodeNop},
	bscript.OpNOP4:  {bscript.OpNOP4, "OP_NOP4", 1, opcodeNop},
	bscript.OpNOP5:  {bscript.OpNOP5, "OP_NOP5", 1, opcodeNop},
	bscript.OpNOP6:  {bscript.OpNOP6, "OP_NOP6", 1, opcodeNop},
	bscript.OpNOP7:  {bscript.OpNOP7, "OP_NOP7", 1, opcodeNop},
	bscript.OpNOP8:  {bscript.OpNOP8, "OP_NOP8", 1, opcodeNop},
	bscript.OpNOP9:  {bscript.OpNOP9, "OP_NOP9", 1, opcodeNop},
	bscript.OpNOP10: {bscript.OpNOP10, "OP_NOP10", 1, opcodeNop},

	// Undefined opcodes.
	bscript.OpUNKNOWN186: {bscript.OpUNKNOWN186, "OP_UNKNOWN186", 1, opcodeInvalid},
	bscript.OpUNKNOWN187: {bscript.OpUNKNOWN187, "OP_UNKNOWN187", 1, opcodeInvalid},
	bscript.OpUNKNOWN188: {bscript.OpUNKNOWN188, "OP_UNKNOWN188", 1, opcodeInvalid},
	bscript.OpUNKNOWN189: {bscript.OpUNKNOWN189, "OP_UNKNOWN189", 1, opcodeInvalid},
	bscript.OpUNKNOWN190: {bscript.OpUNKNOWN190, "OP_UNKNOWN190", 1, opcodeInvalid},
	bscript.OpUNKNOWN191: {bscript.OpUNKNOWN191, "OP_UNKNOWN191", 1, opcodeInvalid},
	bscript.OpUNKNOWN192: {bscript.OpUNKNOWN192, "OP_UNKNOWN192", 1, opcodeInvalid},
	bscript.OpUNKNOWN193: {bscript.OpUNKNOWN193, "OP_UNKNOWN193", 1, opcodeInvalid},
	bscript.OpUNKNOWN194: {bscript.OpUNKNOWN194, "OP_UNKNOWN194", 1, opcodeInvalid},
	bscript.OpUNKNOWN195: {bscript.OpUNKNOWN195, "OP_UNKNOWN195", 1, opcodeInvalid},
	bscript.OpUNKNOWN196: {bscript.OpUNKNOWN196, "OP_UNKNOWN196", 1, opcodeInvalid},
	bscript.OpUNKNOWN197: {bscript.OpUNKNOWN197, "OP_UNKNOWN197", 1, opcodeInvalid},
	bscript.OpUNKNOWN198: {bscript.OpUNKNOWN198, "OP_UNKNOWN198", 1, opcodeInvalid},
	bscript.OpUNKNOWN199: {bscript.OpUNKNOWN199, "OP_UNKNOWN199", 1, opcodeInvalid},
	bscript.OpUNKNOWN200: {bscript.OpUNKNOWN200, "OP_UNKNOWN200", 1, opcodeInvalid},
	bscript.OpUNKNOWN201: {bscript.OpUNKNOWN201, "OP_UNKNOWN201", 1, opcodeInvalid},
	bscript.OpUNKNOWN202: {bscript.OpUNKNOWN202, "OP_UNKNOWN202", 1, opcodeInvalid},
	bscript.OpUNKNOWN203: {bscript.OpUNKNOWN203, "OP_UNKNOWN203", 1, opcodeInvalid},
	bscript.OpUNKNOWN204: {bscript.OpUNKNOWN204, "OP_UNKNOWN204", 1, opcodeInvalid},
	bscript.OpUNKNOWN205: {bscript.OpUNKNOWN205, "OP_UNKNOWN205", 1, opcodeInvalid},
	bscript.OpUNKNOWN206: {bscript.OpUNKNOWN206, "OP_UNKNOWN206", 1, opcodeInvalid},
	bscript.OpUNKNOWN207: {bscript.OpUNKNOWN207, "OP_UNKNOWN207", 1, opcodeInvalid},
	bscript.OpUNKNOWN208: {bscript.OpUNKNOWN208, "OP_UNKNOWN208", 1, opcodeInvalid},
	bscript.OpUNKNOWN209: {bscript.OpUNKNOWN209, "OP_UNKNOWN209", 1, opcodeInvalid},
	bscript.OpUNKNOWN210: {bscript.OpUNKNOWN210, "OP_UNKNOWN210", 1, opcodeInvalid},
	bscript.OpUNKNOWN211: {bscript.OpUNKNOWN211, "OP_UNKNOWN211", 1, opcodeInvalid},
	bscript.OpUNKNOWN212: {bscript.OpUNKNOWN212, "OP_UNKNOWN212", 1, opcodeInvalid},
	bscript.OpUNKNOWN213: {bscript.OpUNKNOWN213, "OP_UNKNOWN213", 1, opcodeInvalid},
	bscript.OpUNKNOWN214: {bscript.OpUNKNOWN214, "OP_UNKNOWN214", 1, opcodeInvalid},
	bscript.OpUNKNOWN215: {bscript.OpUNKNOWN215, "OP_UNKNOWN215", 1, opcodeInvalid},
	bscript.OpUNKNOWN216: {bscript.OpUNKNOWN216, "OP_UNKNOWN216", 1, opcodeInvalid},
	bscript.OpUNKNOWN217: {bscript.OpUNKNOWN217, "OP_UNKNOWN217", 1, opcodeInvalid},
	bscript.OpUNKNOWN218: {bscript.OpUNKNOWN218, "OP_UNKNOWN218", 1, opcodeInvalid},
	bscript.OpUNKNOWN219: {bscript.OpUNKNOWN219, "OP_UNKNOWN219", 1, opcodeInvalid},
	bscript.OpUNKNOWN220: {bscript.OpUNKNOWN220, "OP_UNKNOWN220", 1, opcodeInvalid},
	bscript.OpUNKNOWN221: {bscript.OpUNKNOWN221, "OP_UNKNOWN221", 1, opcodeInvalid},
	bscript.OpUNKNOWN222: {bscript.OpUNKNOWN222, "OP_UNKNOWN222", 1, opcodeInvalid},
	bscript.OpUNKNOWN223: {bscript.OpUNKNOWN223, "OP_UNKNOWN223", 1, opcodeInvalid},
	bscript.OpUNKNOWN224: {bscript.OpUNKNOWN224, "OP_UNKNOWN224", 1, opcodeInvalid},
	bscript.OpUNKNOWN225: {bscript.OpUNKNOWN225, "OP_UNKNOWN225", 1, opcodeInvalid},
	bscript.OpUNKNOWN226: {bscript.OpUNKNOWN226, "OP_UNKNOWN226", 1, opcodeInvalid},
	bscript.OpUNKNOWN227: {bscript.OpUNKNOWN227, "OP_UNKNOWN227", 1, opcodeInvalid},
	bscript.OpUNKNOWN228: {bscript.OpUNKNOWN228, "OP_UNKNOWN228", 1, opcodeInvalid},
	bscript.OpUNKNOWN229: {bscript.OpUNKNOWN229, "OP_UNKNOWN229", 1, opcodeInvalid},
	bscript.OpUNKNOWN230: {bscript.OpUNKNOWN230, "OP_UNKNOWN230", 1, opcodeInvalid},
	bscript.OpUNKNOWN231: {bscript.OpUNKNOWN231, "OP_UNKNOWN231", 1, opcodeInvalid},
	bscript.OpUNKNOWN232: {bscript.OpUNKNOWN232, "OP_UNKNOWN232", 1, opcodeInvalid},
	bscript.OpUNKNOWN233: {bscript.OpUNKNOWN233, "OP_UNKNOWN233", 1, opcodeInvalid},
	bscript.OpUNKNOWN234: {bscript.OpUNKNOWN234, "OP_UNKNOWN234", 1, opcodeInvalid},
	bscript.OpUNKNOWN235: {bscript.OpUNKNOWN235, "OP_UNKNOWN235", 1, opcodeInvalid},
	bscript.OpUNKNOWN236: {bscript.OpUNKNOWN236, "OP_UNKNOWN236", 1, opcodeInvalid},
	bscript.OpUNKNOWN237: {bscript.OpUNKNOWN237, "OP_UNKNOWN237", 1, opcodeInvalid},
	bscript.OpUNKNOWN238: {bscript.OpUNKNOWN238, "OP_UNKNOWN238", 1, opcodeInvalid},
	bscript.OpUNKNOWN239: {bscript.OpUNKNOWN239, "OP_UNKNOWN239", 1, opcodeInvalid},
	bscript.OpUNKNOWN240: {bscript.OpUNKNOWN240, "OP_UNKNOWN240", 1, opcodeInvalid},
	bscript.OpUNKNOWN241: {bscript.OpUNKNOWN241, "OP_UNKNOWN241", 1, opcodeInvalid},
	bscript.OpUNKNOWN242: {bscript.OpUNKNOWN242, "OP_UNKNOWN242", 1, opcodeInvalid},
	bscript.OpUNKNOWN243: {bscript.OpUNKNOWN243, "OP_UNKNOWN243", 1, opcodeInvalid},
	bscript.OpUNKNOWN244: {bscript.OpUNKNOWN244, "OP_UNKNOWN244", 1, opcodeInvalid},
	bscript.OpUNKNOWN245: {bscript.OpUNKNOWN245, "OP_UNKNOWN245", 1, opcodeInvalid},
	bscript.OpUNKNOWN246: {bscript.OpUNKNOWN246, "OP_UNKNOWN246", 1, opcodeInvalid},
	bscript.OpUNKNOWN247: {bscript.OpUNKNOWN247, "OP_UNKNOWN247", 1, opcodeInvalid},
	bscript.OpUNKNOWN248: {bscript.OpUNKNOWN248, "OP_UNKNOWN248", 1, opcodeInvalid},
	bscript.OpUNKNOWN249: {bscript.OpUNKNOWN249, "OP_UNKNOWN249", 1, opcodeInvalid},

	// Bitcoin Core internal use opcode.  Defined here for completeness.
	bscript.OpSMALLINTEGER: {bscript.OpSMALLINTEGER, "OP_SMALLINTEGER", 1, opcodeInvalid},
	bscript.OpPUBKEYS:      {bscript.OpPUBKEYS, "OP_PUBKEYS", 1, opcodeInvalid},
	bscript.OpUNKNOWN252:   {bscript.OpUNKNOWN252, "OP_UNKNOWN252", 1, opcodeInvalid},
	bscript.OpPUBKEYHASH:   {bscript.OpPUBKEYHASH, "OP_PUBKEYHASH", 1, opcodeInvalid},
	bscript.OpPUBKEY:       {bscript.OpPUBKEY, "OP_PUBKEY", 1, opcodeInvalid},

	bscript.OpINVALIDOPCODE: {bscript.OpINVALIDOPCODE, "OP_INVALIDOPCODE", 1, opcodeInvalid},
}

// *******************************************
// Opcode implementation functions start here.
// *******************************************

// opcodeDisabled is a common handler for disabled opcodes.  It returns an
// appropriate error indicating the opcode is disabled.  While it would
// ordinarily make more sense to detect if the script contains any disabled
// opcodes before executing in an initial parse step, the consensus rules
// dictate the script doesn't fail until the program counter passes over a
// disabled opcode (even when they appear in a branch that is not executed).
func opcodeDisabled(op *ParsedOp, t *thread) error {
	return scriptError(ErrDisabledOpcode, "attempt to execute disabled opcode %s", op.Name())
}

func opcodeVerConditional(op *ParsedOp, t *thread) error {
	if t.afterGenesis && !t.shouldExec(*op) {
		return nil
	}
	return opcodeReserved(op, t)
}

// opcodeReserved is a common handler for all reserved opcodes.  It returns an
// appropriate error indicating the opcode is reserved.
func opcodeReserved(op *ParsedOp, t *thread) error {
	return scriptError(ErrReservedOpcode, "attempt to execute reserved opcode %s", op.Name())
}

// opcodeInvalid is a common handler for all invalid opcodes.  It returns an
// appropriate error indicating the opcode is invalid.
func opcodeInvalid(op *ParsedOp, t *thread) error {
	return scriptError(ErrReservedOpcode, "attempt to execute invalid opcode %s", op.Name())
}

// opcodeFalse pushes an empty array to the data stack to represent false.  Note
// that 0, when encoded as a number according to the numeric encoding consensus
// rules, is an empty array.
func opcodeFalse(op *ParsedOp, t *thread) error {
	t.dstack.PushByteArray(nil)
	return nil
}

// opcodePushData is a common handler for the vast majority of opcodes that push
// raw data (bytes) to the data stack.
func opcodePushData(op *ParsedOp, t *thread) error {
	t.dstack.PushByteArray(op.Data)
	return nil
}

// opcode1Negate pushes -1, encoded as a number, to the data stack.
func opcode1Negate(op *ParsedOp, t *thread) error {
	t.dstack.PushInt(-1)
	return nil
}

// opcodeN is a common handler for the small integer data push opcodes.  It
// pushes the numeric value the opcode represents (which will be from 1 to 16)
// onto the data stack.
func opcodeN(op *ParsedOp, t *thread) error {
	// The opcodes are all defined consecutively, so the numeric value is
	// the difference.
	t.dstack.PushInt(scriptNum((op.Op.val - (bscript.Op1 - 1))))
	return nil
}

// opcodeNop is a common handler for the NOP family of opcodes.  As the name
// implies it generally does nothing, however, it will return an error when
// the flag to discourage use of NOPs is set for select opcodes.
func opcodeNop(op *ParsedOp, t *thread) error {
	switch op.Op.val {
	case bscript.OpNOP1, bscript.OpNOP4, bscript.OpNOP5,
		bscript.OpNOP6, bscript.OpNOP7, bscript.OpNOP8, bscript.OpNOP9, bscript.OpNOP10:
		if t.hasFlag(ScriptDiscourageUpgradableNops) {
			return scriptError(
				ErrDiscourageUpgradableNOPs,
				"bscript.OpNOP%d reserved for soft-fork upgrades",
				op.Op.val-(bscript.OpNOP1-1),
			)
		}
	}

	return nil
}

// popIfBool pops the top item off the stack and returns a bool
func popIfBool(t *thread) (bool, error) {
	if t.hasFlag(ScriptVerifyMinimalIf) {
		b, err := t.dstack.PopByteArray()
		if err != nil {
			return false, err
		}

		if len(b) > 1 {
			return false, scriptError(ErrMinimalIf, "conditionl has data of length %d", len(b))
		}
		if len(b) == 1 && b[0] != 1 {
			return false, scriptError(ErrMinimalIf, "conditional failed")
		}

		return asBool(b), nil
	}

	return t.dstack.PopBool()
}

// opcodeIf treats the top item on the data stack as a boolean and removes it.
//
// An appropriate entry is added to the conditional stack depending on whether
// the boolean is true and whether this if is on an executing branch in order
// to allow proper execution of further opcodes depending on the conditional
// logic.  When the boolean is true, the first branch will be executed (unless
// this opcode is nested in a non-executed branch).
//
// <expression> if [statements] [else [statements]] endif
//
// Note that, unlike for all non-conditional opcodes, this is executed even when
// it is on a non-executing branch so proper nesting is maintained.
//
// Data stack transformation: [... bool] -> [...]
// Conditional stack transformation: [...] -> [... OpCondValue]
func opcodeIf(op *ParsedOp, t *thread) error {
	condVal := OpCondFalse
	if t.shouldExec(*op) {
		if t.isBranchExecuting() {
			ok, err := popIfBool(t)
			if err != nil {
				return err
			}

			if ok {
				condVal = OpCondTrue
			}
		} else {
			condVal = OpCondSkip
		}
	}

	t.condStack = append(t.condStack, condVal)
	t.elseStack.PushBool(false)
	return nil
}

// opcodeNotIf treats the top item on the data stack as a boolean and removes
// it.
//
// An appropriate entry is added to the conditional stack depending on whether
// the boolean is true and whether this if is on an executing branch in order
// to allow proper execution of further opcodes depending on the conditional
// logic.  When the boolean is false, the first branch will be executed (unless
// this opcode is nested in a non-executed branch).
//
// <expression> notif [statements] [else [statements]] endif
//
// Note that, unlike for all non-conditional opcodes, this is executed even when
// it is on a non-executing branch so proper nesting is maintained.
//
// Data stack transformation: [... bool] -> [...]
// Conditional stack transformation: [...] -> [... OpCondValue]
func opcodeNotIf(op *ParsedOp, t *thread) error {
	condVal := OpCondFalse
	if t.shouldExec(*op) {
		if t.isBranchExecuting() {
			ok, err := popIfBool(t)
			if err != nil {
				return err
			}

			if !ok {
				condVal = OpCondTrue
			}
		} else {
			condVal = OpCondSkip
		}
	}

	t.condStack = append(t.condStack, condVal)
	t.elseStack.PushBool(false)
	return nil
}

// opcodeElse inverts conditional execution for other half of if/else/endif.
//
// An error is returned if there has not already been a matching bscript.OpIF.
//
// Conditional stack transformation: [... OpCondValue] -> [... !OpCondValue]
func opcodeElse(op *ParsedOp, t *thread) error {
	if len(t.condStack) == 0 {
		return scriptError(ErrUnbalancedConditional,
			"encountered opcode %s with no matching opcode to begin conditional execution", op.Name())
	}

	// Only one ELSE allowed in IF after genesis
	ok, err := t.elseStack.PopBool()
	if err != nil {
		return err
	}
	if ok {
		return scriptError(ErrUnbalancedConditional,
			"encountered opcode %s with no matching opcode to begin conditional execution", op.Name())
	}

	conditionalIdx := len(t.condStack) - 1
	switch t.condStack[conditionalIdx] {
	case OpCondTrue:
		t.condStack[conditionalIdx] = OpCondFalse
	case OpCondFalse:
		t.condStack[conditionalIdx] = OpCondTrue
	case OpCondSkip:
		// Value doesn't change in skip since it indicates this opcode
		// is nested in a non-executed branch.
	}

	t.elseStack.PushBool(true)
	return nil
}

// opcodeEndif terminates a conditional block, removing the value from the
// conditional execution stack.
//
// An error is returned if there has not already been a matching bscript.OpIF.
//
// Conditional stack transformation: [... OpCondValue] -> [...]
func opcodeEndif(op *ParsedOp, t *thread) error {
	if len(t.condStack) == 0 {
		return scriptError(ErrUnbalancedConditional,
			"encountered opcode %s with no matching opcode to begin conditional execution", op.Name())
	}

	t.condStack = t.condStack[:len(t.condStack)-1]
	if _, err := t.elseStack.PopBool(); err != nil {
		return err
	}

	return nil
}

// abstractVerify examines the top item on the data stack as a boolean value and
// verifies it evaluates to true.  An error is returned either when there is no
// item on the stack or when that item evaluates to false.  In the latter case
// where the verification fails specifically due to the top item evaluating
// to false, the returned error will use the passed error code.
func abstractVerify(op *ParsedOp, t *thread, c ErrorCode) error {
	verified, err := t.dstack.PopBool()
	if err != nil {
		return err
	}
	if !verified {
		return scriptError(c, "%s failed", op.Name())
	}

	return nil
}

// opcodeVerify examines the top item on the data stack as a boolean value and
// verifies it evaluates to true.  An error is returned if it does not.
func opcodeVerify(op *ParsedOp, t *thread) error {
	return abstractVerify(op, t, ErrVerify)
}

// opcodeReturn returns an appropriate error since it is always an error to
// return early from a script.
func opcodeReturn(op *ParsedOp, t *thread) error {
	if !t.afterGenesis {
		return scriptError(ErrEarlyReturn, "script returned early")
	}

	t.earlyReturnAfterGenesis = true
	if len(t.condStack) == 0 {
		// Terminate the execution as successful. The remaining of the script does not affect the validity (even in
		// presence of unbalanced IFs, invalid opcodes etc)
		return success()
	}

	return nil
}

// verifyLockTime is a helper function used to validate locktimes.
func verifyLockTime(txLockTime, threshold, lockTime int64) error {
	// The lockTimes in both the script and transaction must be of the same
	// type.
	if !((txLockTime < threshold && lockTime < threshold) ||
		(txLockTime >= threshold && lockTime >= threshold)) {
		return scriptError(ErrUnsatisfiedLockTime,
			"mismatched locktime types -- tx locktime %d, stack locktime %d", txLockTime, lockTime)
	}

	if lockTime > txLockTime {
		return scriptError(ErrUnsatisfiedLockTime,
			"locktime requirement not satisfied -- locktime is greater than the transaction locktime: %d > %d",
			lockTime, txLockTime)
	}

	return nil
}

// opcodeCheckLockTimeVerify compares the top item on the data stack to the
// LockTime field of the transaction containing the script signature
// validating if the transaction outputs are spendable yet.  If flag
// ScriptVerifyCheckLockTimeVerify is not set, the code continues as if bscript.OpNOP2
// were executed.
func opcodeCheckLockTimeVerify(op *ParsedOp, t *thread) error {
	// If the ScriptVerifyCheckLockTimeVerify script flag is not set, treat
	// opcode as bscript.OpNOP2 instead.
	if !t.hasFlag(ScriptVerifyCheckLockTimeVerify) || t.afterGenesis {
		if t.hasFlag(ScriptDiscourageUpgradableNops) {
			return scriptError(ErrDiscourageUpgradableNOPs, "bscript.OpNOP2 reserved for soft-fork upgrades")
		}

		return nil
	}

	// The current transaction locktime is a uint32 resulting in a maximum
	// locktime of 2^32-1 (the year 2106).  However, scriptNums are signed
	// and therefore a standard 4-byte scriptNum would only support up to a
	// maximum of 2^31-1 (the year 2038).  Thus, a 5-byte scriptNum is used
	// here since it will support up to 2^39-1 which allows dates beyond the
	// current locktime limit.
	//
	// PeekByteArray is used here instead of PeekInt because we do not want
	// to be limited to a 4-byte integer for reasons specified above.
	so, err := t.dstack.PeekByteArray(0)
	if err != nil {
		return err
	}
	lockTime, err := makeScriptNum(so, t.dstack.verifyMinimalData, 5)
	if err != nil {
		return err
	}

	// In the rare event that the argument needs to be < 0 due to some
	// arithmetic being done first, you can always use
	// 0 bscript.OpMAX bscript.OpCHECKLOCKTIMEVERIFY.
	if lockTime < 0 {
		return scriptError(ErrNegativeLockTime, "negative lock time: %d", lockTime)
	}

	// The lock time field of a transaction is either a block height at
	// which the transaction is finalised or a timestamp depending on if the
	// value is before the interpreter.LockTimeThreshold.  When it is under the
	// threshold it is a block height.
	if err = verifyLockTime(int64(t.tx.LockTime), LockTimeThreshold, int64(lockTime)); err != nil {
		return err
	}

	// The lock time feature can also be disabled, thereby bypassing
	// bscript.OpCHECKLOCKTIMEVERIFY, if every transaction input has been finalised by
	// setting its sequence to the maximum value (bt.MaxTxInSequenceNum).  This
	// condition would result in the transaction being allowed into the blockchain
	// making the opcode ineffective.
	//
	// This condition is prevented by enforcing that the input being used by
	// the opcode is unlocked (its sequence number is less than the max
	// value).  This is sufficient to prove correctness without having to
	// check every input.
	//
	// NOTE: This implies that even if the transaction is not finalised due to
	// another input being unlocked, the opcode execution will still fail when the
	// input being used by the opcode is locked.
	if t.tx.Inputs[t.inputIdx].SequenceNumber == bt.MaxTxInSequenceNum {
		return scriptError(ErrUnsatisfiedLockTime, "transaction input is finalised")
	}

	return nil
}

// opcodeCheckSequenceVerify compares the top item on the data stack to the
// LockTime field of the transaction containing the script signature
// validating if the transaction outputs are spendable yet.  If flag
// ScriptVerifyCheckSequenceVerify is not set, the code continues as if bscript.OpNOP3
// were executed.
func opcodeCheckSequenceVerify(op *ParsedOp, t *thread) error {
	// If the ScriptVerifyCheckSequenceVerify script flag is not set, treat
	// opcode as bscript.OpNOP3 instead.
	if !t.hasFlag(ScriptVerifyCheckSequenceVerify) || t.afterGenesis {
		if t.hasFlag(ScriptDiscourageUpgradableNops) {
			return scriptError(ErrDiscourageUpgradableNOPs, "bscript.OpNOP3 reserved for soft-fork upgrades")
		}

		return nil
	}

	// The current transaction sequence is a uint32 resulting in a maximum
	// sequence of 2^32-1.  However, scriptNums are signed and therefore a
	// standard 4-byte scriptNum would only support up to a maximum of
	// 2^31-1.  Thus, a 5-byte scriptNum is used here since it will support
	// up to 2^39-1 which allows sequences beyond the current sequence
	// limit.
	//
	// PeekByteArray is used here instead of PeekInt because we do not want
	// to be limited to a 4-byte integer for reasons specified above.
	so, err := t.dstack.PeekByteArray(0)
	if err != nil {
		return err
	}
	stackSequence, err := makeScriptNum(so, t.dstack.verifyMinimalData, 5)
	if err != nil {
		return err
	}

	// In the rare event that the argument needs to be < 0 due to some
	// arithmetic being done first, you can always use
	// 0 bscript.OpMAX bscript.OpCHECKSEQUENCEVERIFY.
	if stackSequence < 0 {
		return scriptError(ErrNegativeLockTime, "negative sequence: %d", stackSequence)
	}

	sequence := int64(stackSequence)

	// To provide for future soft-fork extensibility, if the
	// operand has the disabled lock-time flag set,
	// CHECKSEQUENCEVERIFY behaves as a NOP.
	if sequence&int64(bt.SequenceLockTimeDisabled) != 0 {
		return nil
	}

	// Transaction version numbers not high enough to trigger CSV rules must
	// fail.
	if t.tx.Version < 2 {
		return scriptError(ErrUnsatisfiedLockTime, "invalid transaction version: %d", t.tx.Version)
	}

	// Sequence numbers with their most significant bit set are not
	// consensus constrained. Testing that the transaction's sequence
	// number does not have this bit set prevents using this property
	// to get around a CHECKSEQUENCEVERIFY check.
	txSequence := int64(t.tx.Inputs[t.inputIdx].SequenceNumber)
	if txSequence&int64(bt.SequenceLockTimeDisabled) != 0 {
		return scriptError(ErrUnsatisfiedLockTime,
			"transaction sequence has sequence locktime disabled bit set: 0x%x", txSequence)
	}

	// Mask off non-consensus bits before doing comparisons.
	lockTimeMask := int64(bt.SequenceLockTimeIsSeconds | bt.SequenceLockTimeMask)

	return verifyLockTime(txSequence&lockTimeMask, bt.SequenceLockTimeIsSeconds, sequence&lockTimeMask)
}

// opcodeToAltStack removes the top item from the main data stack and pushes it
// onto the alternate data stack.
//
// Main data stack transformation: [... x1 x2 x3] -> [... x1 x2]
// Alt data stack transformation:  [... y1 y2 y3] -> [... y1 y2 y3 x3]
func opcodeToAltStack(op *ParsedOp, t *thread) error {
	so, err := t.dstack.PopByteArray()
	if err != nil {
		return err
	}

	t.astack.PushByteArray(so)

	return nil
}

// opcodeFromAltStack removes the top item from the alternate data stack and
// pushes it onto the main data stack.
//
// Main data stack transformation: [... x1 x2 x3] -> [... x1 x2 x3 y3]
// Alt data stack transformation:  [... y1 y2 y3] -> [... y1 y2]
func opcodeFromAltStack(op *ParsedOp, t *thread) error {
	so, err := t.astack.PopByteArray()
	if err != nil {
		return err
	}

	t.dstack.PushByteArray(so)

	return nil
}

// opcode2Drop removes the top 2 items from the data stack.
//
// Stack transformation: [... x1 x2 x3] -> [... x1]
func opcode2Drop(op *ParsedOp, t *thread) error {
	return t.dstack.DropN(2)
}

// opcode2Dup duplicates the top 2 items on the data stack.
//
// Stack transformation: [... x1 x2 x3] -> [... x1 x2 x3 x2 x3]
func opcode2Dup(op *ParsedOp, t *thread) error {
	return t.dstack.DupN(2)
}

// opcode3Dup duplicates the top 3 items on the data stack.
//
// Stack transformation: [... x1 x2 x3] -> [... x1 x2 x3 x1 x2 x3]
func opcode3Dup(op *ParsedOp, t *thread) error {
	return t.dstack.DupN(3)
}

// opcode2Over duplicates the 2 items before the top 2 items on the data stack.
//
// Stack transformation: [... x1 x2 x3 x4] -> [... x1 x2 x3 x4 x1 x2]
func opcode2Over(op *ParsedOp, t *thread) error {
	return t.dstack.OverN(2)
}

// opcode2Rot rotates the top 6 items on the data stack to the left twice.
//
// Stack transformation: [... x1 x2 x3 x4 x5 x6] -> [... x3 x4 x5 x6 x1 x2]
func opcode2Rot(op *ParsedOp, t *thread) error {
	return t.dstack.RotN(2)
}

// opcode2Swap swaps the top 2 items on the data stack with the 2 that come
// before them.
//
// Stack transformation: [... x1 x2 x3 x4] -> [... x3 x4 x1 x2]
func opcode2Swap(op *ParsedOp, t *thread) error {
	return t.dstack.SwapN(2)
}

// opcodeIfDup duplicates the top item of the stack if it is not zero.
//
// Stack transformation (x1==0): [... x1] -> [... x1]
// Stack transformation (x1!=0): [... x1] -> [... x1 x1]
func opcodeIfDup(op *ParsedOp, t *thread) error {
	so, err := t.dstack.PeekByteArray(0)
	if err != nil {
		return err
	}

	// Push copy of data iff it isn't zero
	if asBool(so) {
		t.dstack.PushByteArray(so)
	}

	return nil
}

// opcodeDepth pushes the depth of the data stack prior to executing this
// opcode, encoded as a number, onto the data stack.
//
// Stack transformation: [...] -> [... <num of items on the stack>]
// Example with 2 items: [x1 x2] -> [x1 x2 2]
// Example with 3 items: [x1 x2 x3] -> [x1 x2 x3 3]
func opcodeDepth(op *ParsedOp, t *thread) error {
	t.dstack.PushInt(scriptNum(t.dstack.Depth()))
	return nil
}

// opcodeDrop removes the top item from the data stack.
//
// Stack transformation: [... x1 x2 x3] -> [... x1 x2]
func opcodeDrop(op *ParsedOp, t *thread) error {
	return t.dstack.DropN(1)
}

// opcodeDup duplicates the top item on the data stack.
//
// Stack transformation: [... x1 x2 x3] -> [... x1 x2 x3 x3]
func opcodeDup(op *ParsedOp, t *thread) error {
	return t.dstack.DupN(1)
}

// opcodeNip removes the item before the top item on the data stack.
//
// Stack transformation: [... x1 x2 x3] -> [... x1 x3]
func opcodeNip(op *ParsedOp, t *thread) error {
	return t.dstack.NipN(1)
}

// opcodeOver duplicates the item before the top item on the data stack.
//
// Stack transformation: [... x1 x2 x3] -> [... x1 x2 x3 x2]
func opcodeOver(op *ParsedOp, t *thread) error {
	return t.dstack.OverN(1)
}

// opcodePick treats the top item on the data stack as an integer and duplicates
// the item on the stack that number of items back to the top.
//
// Stack transformation: [xn ... x2 x1 x0 n] -> [xn ... x2 x1 x0 xn]
// Example with n=1: [x2 x1 x0 1] -> [x2 x1 x0 x1]
// Example with n=2: [x2 x1 x0 2] -> [x2 x1 x0 x2]
func opcodePick(op *ParsedOp, t *thread) error {
	val, err := t.dstack.PopInt()
	if err != nil {
		return err
	}

	return t.dstack.PickN(val.Int32())
}

// opcodeRoll treats the top item on the data stack as an integer and moves
// the item on the stack that number of items back to the top.
//
// Stack transformation: [xn ... x2 x1 x0 n] -> [... x2 x1 x0 xn]
// Example with n=1: [x2 x1 x0 1] -> [x2 x0 x1]
// Example with n=2: [x2 x1 x0 2] -> [x1 x0 x2]
func opcodeRoll(op *ParsedOp, t *thread) error {
	val, err := t.dstack.PopInt()
	if err != nil {
		return err
	}

	return t.dstack.RollN(val.Int32())
}

// opcodeRot rotates the top 3 items on the data stack to the left.
//
// Stack transformation: [... x1 x2 x3] -> [... x2 x3 x1]
func opcodeRot(op *ParsedOp, t *thread) error {
	return t.dstack.RotN(1)
}

// opcodeSwap swaps the top two items on the stack.
//
// Stack transformation: [... x1 x2] -> [... x2 x1]
func opcodeSwap(op *ParsedOp, t *thread) error {
	return t.dstack.SwapN(1)
}

// opcodeTuck inserts a duplicate of the top item of the data stack before the
// second-to-top item.
//
// Stack transformation: [... x1 x2] -> [... x2 x1 x2]
func opcodeTuck(op *ParsedOp, t *thread) error {
	return t.dstack.Tuck()
}

// opcodeCat concatenates two byte sequences. The result must
// not be larger than MaxScriptElementSize.
//
// Stack transformation: {Ox11} {0x22, 0x33} bscript.OpCAT -> 0x112233
func opcodeCat(op *ParsedOp, t *thread) error {
	b, err := t.dstack.PopByteArray()
	if err != nil {
		return err
	}

	a, err := t.dstack.PopByteArray()
	if err != nil {
		return err
	}

	c := append(a, b...)
	if len(c) > t.cfg.MaxScriptElementSize() {
		return scriptError(ErrElementTooBig,
			"concatenated size %d exceeds max allowed size %d", len(c), t.cfg.MaxScriptElementSize())
	}

	t.dstack.PushByteArray(c)
	return nil
}

// opcodeSplit splits the operand at the given position.
// This operation is the exact inverse of bscript.OpCAT
//
// Stack transformation: x n bscript.OpSPLIT -> x1 x2
func opcodeSplit(op *ParsedOp, t *thread) error {
	n, err := t.dstack.PopInt()
	if err != nil {
		return err
	}

	c, err := t.dstack.PopByteArray()
	if err != nil {
		return err
	}

	if n.Int32() > int32(len(c)) {
		return scriptError(ErrNumberTooBig, "n is larger than length of array")
	}
	if n < 0 {
		return scriptError(ErrNumberTooSmall, "n is negative")
	}

	a := c[:n]
	b := c[n:]
	t.dstack.PushByteArray(a)
	t.dstack.PushByteArray(b)

	return nil
}

// opcodeNum2Bin converts the numeric value into a byte sequence of a
// certain size, taking account of the sign bit. The byte sequence
// produced uses the little-endian encoding.
//
// Stack transformation: a b bscript.OpNUM2BIN -> x
func opcodeNum2bin(op *ParsedOp, t *thread) error {
	n, err := t.dstack.PopInt()
	if err != nil {
		return err
	}

	a, err := t.dstack.PopByteArray()
	if err != nil {
		return err
	}

	size := int(n.Int32())
	if size > t.cfg.MaxScriptElementSize() {
		return scriptError(ErrNumberTooBig, "n is larger than the max of %d", defaultScriptNumLen)
	}

	// encode a as a script num so that we we take the bytes it
	// will be minimally encoded.
	sn, err := makeScriptNum(a, false, len(a))
	if err != nil {
		return err
	}

	b := sn.Bytes()
	if len(b) > size {
		return scriptError(ErrNumberTooSmall, "cannot fit it into n sized array")
	}
	if len(b) == size {
		t.dstack.PushByteArray(b)
		return nil
	}

	signbit := byte(0x00)
	if len(b) > 0 {
		signbit = b[0] & 0x80
		b[len(b)-1] &= 0x7f
	}

	for len(b) < size-1 {
		b = append(b, 0x00)
	}

	b = append(b, signbit)

	t.dstack.PushByteArray(b)
	return nil
}

// opcodeBin2num converts the byte sequence into a numeric value,
// including minimal encoding. The byte sequence must encode the
// value in little-endian encoding.
//
// Stack transformation: a bscript.OpBIN2NUM -> x
func opcodeBin2num(op *ParsedOp, t *thread) error {
	a, err := t.dstack.PopByteArray()
	if err != nil {
		return err
	}

	n, err := makeScriptNum(a, false, len(a))
	if err != nil {
		return err
	}
	if len(n.Bytes()) > defaultScriptNumLen {
		return scriptError(ErrNumberTooBig,
			fmt.Sprintf("script numbers are limited to %d bytes", defaultScriptNumLen))
	}

	t.dstack.PushInt(n)
	return nil
}

// opcodeSize pushes the size of the top item of the data stack onto the data
// stack.
//
// Stack transformation: [... x1] -> [... x1 len(x1)]
func opcodeSize(op *ParsedOp, t *thread) error {
	so, err := t.dstack.PeekByteArray(0)
	if err != nil {
		return err
	}

	t.dstack.PushInt(scriptNum(len(so)))
	return nil
}

// opcodeInvert flips all of the top stack item's bits
//
// Stack transformation: a -> ~a
func opcodeInvert(op *ParsedOp, t *thread) error {
	ba, err := t.dstack.PeekByteArray(0)
	if err != nil {
		return err
	}

	for i := range ba {
		ba[i] = ba[i] ^ 0xFF
	}

	return nil
}

// opcodeAnd executes a boolean and between each bit in the operands
//
// Stack transformation: x1 x2 bscript.OpAND -> out
func opcodeAnd(op *ParsedOp, t *thread) error {
	a, err := t.dstack.PopByteArray()
	if err != nil {
		return err
	}

	b, err := t.dstack.PopByteArray()
	if err != nil {
		return err
	}

	if len(a) != len(b) {
		return scriptError(ErrInvalidInputLength, "byte arrays are not the same length")
	}

	c := make([]byte, len(a))
	for i := range a {
		c[i] = a[i] & b[i]
	}

	t.dstack.PushByteArray(c)
	return nil
}

// opcodeOr executes a boolean or between each bit in the operands
//
// Stack transformation: x1 x2 bscript.OpOR -> out
func opcodeOr(op *ParsedOp, t *thread) error {
	a, err := t.dstack.PopByteArray()
	if err != nil {
		return err
	}

	b, err := t.dstack.PopByteArray()
	if err != nil {
		return err
	}

	if len(a) != len(b) {
		return scriptError(ErrInvalidInputLength, "byte arrays are not the same length")
	}

	c := make([]byte, len(a))
	for i := range a {
		c[i] = a[i] | b[i]
	}

	t.dstack.PushByteArray(c)
	return nil
}

// opcodeXor executes a boolean xor between each bit in the operands
//
// Stack transformation: x1 x2 bscript.OpXOR -> out
func opcodeXor(op *ParsedOp, t *thread) error {
	a, err := t.dstack.PopByteArray()
	if err != nil {
		return err
	}

	b, err := t.dstack.PopByteArray()
	if err != nil {
		return err
	}

	if len(a) != len(b) {
		return scriptError(ErrInvalidInputLength, "byte arrays are not the same length")
	}

	c := make([]byte, len(a))
	for i := range a {
		c[i] = a[i] ^ b[i]
	}

	t.dstack.PushByteArray(c)
	return nil
}

// opcodeEqual removes the top 2 items of the data stack, compares them as raw
// bytes, and pushes the result, encoded as a boolean, back to the stack.
//
// Stack transformation: [... x1 x2] -> [... bool]
func opcodeEqual(op *ParsedOp, t *thread) error {
	a, err := t.dstack.PopByteArray()
	if err != nil {
		return err
	}

	b, err := t.dstack.PopByteArray()
	if err != nil {
		return err
	}

	t.dstack.PushBool(bytes.Equal(a, b))
	return nil
}

// opcodeEqualVerify is a combination of opcodeEqual and opcodeVerify.
// Specifically, it removes the top 2 items of the data stack, compares them,
// and pushes the result, encoded as a boolean, back to the stack.  Then, it
// examines the top item on the data stack as a boolean value and verifies it
// evaluates to true.  An error is returned if it does not.
//
// Stack transformation: [... x1 x2] -> [... bool] -> [...]
func opcodeEqualVerify(op *ParsedOp, t *thread) error {
	if err := opcodeEqual(op, t); err != nil {
		return err
	}

	return abstractVerify(op, t, ErrEqualVerify)
}

// opcode1Add treats the top item on the data stack as an integer and replaces
// it with its incremented value (plus 1).
//
// Stack transformation: [... x1 x2] -> [... x1 x2+1]
func opcode1Add(op *ParsedOp, t *thread) error {
	m, err := t.dstack.PopInt()
	if err != nil {
		return err
	}

	t.dstack.PushInt(m + 1)
	return nil
}

// opcode1Sub treats the top item on the data stack as an integer and replaces
// it with its decremented value (minus 1).
//
// Stack transformation: [... x1 x2] -> [... x1 x2-1]
func opcode1Sub(op *ParsedOp, t *thread) error {
	m, err := t.dstack.PopInt()
	if err != nil {
		return err
	}

	t.dstack.PushInt(m - 1)
	return nil
}

// opcodeNegate treats the top item on the data stack as an integer and replaces
// it with its negation.
//
// Stack transformation: [... x1 x2] -> [... x1 -x2]
func opcodeNegate(op *ParsedOp, t *thread) error {
	m, err := t.dstack.PopInt()
	if err != nil {
		return err
	}

	t.dstack.PushInt(-m)
	return nil
}

// opcodeAbs treats the top item on the data stack as an integer and replaces it
// it with its absolute value.
//
// Stack transformation: [... x1 x2] -> [... x1 abs(x2)]
func opcodeAbs(op *ParsedOp, t *thread) error {
	m, err := t.dstack.PopInt()
	if err != nil {
		return err
	}

	if m < 0 {
		m = -m
	}

	t.dstack.PushInt(m)
	return nil
}

// opcodeNot treats the top item on the data stack as an integer and replaces
// it with its "inverted" value (0 becomes 1, non-zero becomes 0).
//
// NOTE: While it would probably make more sense to treat the top item as a
// boolean, and push the opposite, which is really what the intention of this
// opcode is, it is extremely important that is not done because integers are
// interpreted differently than booleans and the consensus rules for this opcode
// dictate the item is interpreted as an integer.
//
// Stack transformation (x2==0): [... x1 0] -> [... x1 1]
// Stack transformation (x2!=0): [... x1 1] -> [... x1 0]
// Stack transformation (x2!=0): [... x1 17] -> [... x1 0]
func opcodeNot(op *ParsedOp, t *thread) error {
	m, err := t.dstack.PopInt()
	if err != nil {
		return err
	}

	var n scriptNum
	if m == 0 {
		n = 1
	}

	t.dstack.PushInt(n)
	return nil
}

// opcode0NotEqual treats the top item on the data stack as an integer and
// replaces it with either a 0 if it is zero, or a 1 if it is not zero.
//
// Stack transformation (x2==0): [... x1 0] -> [... x1 0]
// Stack transformation (x2!=0): [... x1 1] -> [... x1 1]
// Stack transformation (x2!=0): [... x1 17] -> [... x1 1]
func opcode0NotEqual(op *ParsedOp, t *thread) error {
	m, err := t.dstack.PopInt()
	if err != nil {
		return err
	}

	if m != 0 {
		m = 1
	}

	t.dstack.PushInt(m)
	return nil
}

// opcodeAdd treats the top two items on the data stack as integers and replaces
// them with their sum.
//
// Stack transformation: [... x1 x2] -> [... x1+x2]
func opcodeAdd(op *ParsedOp, t *thread) error {
	v0, err := t.dstack.PopInt()
	if err != nil {
		return err
	}

	v1, err := t.dstack.PopInt()
	if err != nil {
		return err
	}

	t.dstack.PushInt(v0 + v1)
	return nil
}

// opcodeSub treats the top two items on the data stack as integers and replaces
// them with the result of subtracting the top entry from the second-to-top
// entry.
//
// Stack transformation: [... x1 x2] -> [... x1-x2]
func opcodeSub(op *ParsedOp, t *thread) error {
	v0, err := t.dstack.PopInt()
	if err != nil {
		return err
	}

	v1, err := t.dstack.PopInt()
	if err != nil {
		return err
	}

	t.dstack.PushInt(v1 - v0)
	return nil
}

// opcodeMul treats the top two items on the data stack as integers and replaces
// them with the result of subtracting the top entry from the second-to-top
// entry.
func opcodeMul(op *ParsedOp, t *thread) error {
	n1, err := t.dstack.PopInt()
	if err != nil {
		return err
	}

	n2, err := t.dstack.PopInt()
	if err != nil {
		return err
	}

	n3 := n1.Int64() * n2.Int64()

	t.dstack.PushInt(scriptNum(n3))
	return nil
}

// opcodeDiv return the integer quotient of a and b. If the result
// would be a non-integer it is rounded towards zero.
//
// Stack transformation: a b bscript.OpDIV -> out
func opcodeDiv(op *ParsedOp, t *thread) error {
	b, err := t.dstack.PopInt()
	if err != nil {
		return err
	}

	a, err := t.dstack.PopInt()
	if err != nil {
		return err
	}

	if b == 0 {
		return scriptError(ErrDivideByZero, "divide by zero")
	}

	t.dstack.PushInt(a / b)
	return nil
}

// opcodeMod returns the remainder after dividing a by b. The output will
// be represented using the least number of bytes required.
//
// Stack transformation: a b bscript.OpMOD -> out
func opcodeMod(op *ParsedOp, t *thread) error {
	b, err := t.dstack.PopInt()
	if err != nil {
		return err
	}

	a, err := t.dstack.PopInt()
	if err != nil {
		return err
	}

	if b == 0 {
		return scriptError(ErrDivideByZero, "mod by zero")
	}

	t.dstack.PushInt(a % b)
	return nil
}

func opcodeLShift(op *ParsedOp, t *thread) error {
	n, err := t.dstack.PopInt()
	if err != nil {
		return err
	}

	if n.Int32() < 0 {
		return scriptError(ErrNumberTooSmall, "n less than 0")
	}

	x, err := t.dstack.PopByteArray()
	if err != nil {
		return err
	}

	l := len(x)
	for i := 0; i < l-1; i++ {
		x[i] = x[i]<<n | x[i+1]>>(8-n)
	}
	x[l-1] <<= n

	t.dstack.PushByteArray(x)
	return nil
}

func opcodeRShift(op *ParsedOp, t *thread) error {
	n, err := t.dstack.PopInt()
	if err != nil {
		return err
	}

	if n.Int32() < 0 {
		return scriptError(ErrNumberTooSmall, "n less than 0")
	}

	x, err := t.dstack.PopByteArray()
	if err != nil {
		return err
	}

	l := len(x)
	for i := l - 1; i > 0; i-- {
		x[i] = x[i]>>n | x[i-1]<<(8-n)
	}
	x[0] >>= n

	t.dstack.PushByteArray(x)
	return nil
}

// opcodeBoolAnd treats the top two items on the data stack as integers.  When
// both of them are not zero, they are replaced with a 1, otherwise a 0.
//
// Stack transformation (x1==0, x2==0): [... 0 0] -> [... 0]
// Stack transformation (x1!=0, x2==0): [... 5 0] -> [... 0]
// Stack transformation (x1==0, x2!=0): [... 0 7] -> [... 0]
// Stack transformation (x1!=0, x2!=0): [... 4 8] -> [... 1]
func opcodeBoolAnd(op *ParsedOp, t *thread) error {
	v0, err := t.dstack.PopInt()
	if err != nil {
		return err
	}

	v1, err := t.dstack.PopInt()
	if err != nil {
		return err
	}

	var n scriptNum
	if v0 != 0 && v1 != 0 {
		n = 1
	}

	t.dstack.PushInt(n)
	return nil
}

// opcodeBoolOr treats the top two items on the data stack as integers.  When
// either of them are not zero, they are replaced with a 1, otherwise a 0.
//
// Stack transformation (x1==0, x2==0): [... 0 0] -> [... 0]
// Stack transformation (x1!=0, x2==0): [... 5 0] -> [... 1]
// Stack transformation (x1==0, x2!=0): [... 0 7] -> [... 1]
// Stack transformation (x1!=0, x2!=0): [... 4 8] -> [... 1]
func opcodeBoolOr(op *ParsedOp, t *thread) error {
	v0, err := t.dstack.PopInt()
	if err != nil {
		return err
	}

	v1, err := t.dstack.PopInt()
	if err != nil {
		return err
	}

	var n scriptNum
	if v0 != 0 || v1 != 0 {
		n = 1
	}

	t.dstack.PushInt(n)
	return nil
}

// opcodeNumEqual treats the top two items on the data stack as integers.  When
// they are equal, they are replaced with a 1, otherwise a 0.
//
// Stack transformation (x1==x2): [... 5 5] -> [... 1]
// Stack transformation (x1!=x2): [... 5 7] -> [... 0]
func opcodeNumEqual(op *ParsedOp, t *thread) error {
	v0, err := t.dstack.PopInt()
	if err != nil {
		return err
	}

	v1, err := t.dstack.PopInt()
	if err != nil {
		return err
	}

	var n scriptNum
	if v0 == v1 {
		n = 1
	}

	t.dstack.PushInt(n)
	return nil
}

// opcodeNumEqualVerify is a combination of opcodeNumEqual and opcodeVerify.
//
// Specifically, treats the top two items on the data stack as integers.  When
// they are equal, they are replaced with a 1, otherwise a 0.  Then, it examines
// the top item on the data stack as a boolean value and verifies it evaluates
// to true.  An error is returned if it does not.
//
// Stack transformation: [... x1 x2] -> [... bool] -> [...]
func opcodeNumEqualVerify(op *ParsedOp, t *thread) error {
	if err := opcodeNumEqual(op, t); err != nil {
		return err
	}

	return abstractVerify(op, t, ErrNumEqualVerify)
}

// opcodeNumNotEqual treats the top two items on the data stack as integers.
// When they are NOT equal, they are replaced with a 1, otherwise a 0.
//
// Stack transformation (x1==x2): [... 5 5] -> [... 0]
// Stack transformation (x1!=x2): [... 5 7] -> [... 1]
func opcodeNumNotEqual(op *ParsedOp, t *thread) error {
	v0, err := t.dstack.PopInt()
	if err != nil {
		return err
	}

	v1, err := t.dstack.PopInt()
	if err != nil {
		return err
	}

	var n scriptNum
	if v0 != v1 {
		n = 1
	}

	t.dstack.PushInt(n)
	return nil
}

// opcodeLessThan treats the top two items on the data stack as integers.  When
// the second-to-top item is less than the top item, they are replaced with a 1,
// otherwise a 0.
//
// Stack transformation: [... x1 x2] -> [... bool]
func opcodeLessThan(op *ParsedOp, t *thread) error {
	v0, err := t.dstack.PopInt()
	if err != nil {
		return err
	}

	v1, err := t.dstack.PopInt()
	if err != nil {
		return err
	}

	var n scriptNum
	if v1 < v0 {
		n = 1
	}

	t.dstack.PushInt(n)
	return nil
}

// opcodeGreaterThan treats the top two items on the data stack as integers.
// When the second-to-top item is greater than the top item, they are replaced
// with a 1, otherwise a 0.
//
// Stack transformation: [... x1 x2] -> [... bool]
func opcodeGreaterThan(op *ParsedOp, t *thread) error {
	v0, err := t.dstack.PopInt()
	if err != nil {
		return err
	}

	v1, err := t.dstack.PopInt()
	if err != nil {
		return err
	}

	var n scriptNum
	if v1 > v0 {
		n = 1
	}

	t.dstack.PushInt(n)
	return nil
}

// opcodeLessThanOrEqual treats the top two items on the data stack as integers.
// When the second-to-top item is less than or equal to the top item, they are
// replaced with a 1, otherwise a 0.
//
// Stack transformation: [... x1 x2] -> [... bool]
func opcodeLessThanOrEqual(op *ParsedOp, t *thread) error {
	v0, err := t.dstack.PopInt()
	if err != nil {
		return err
	}

	v1, err := t.dstack.PopInt()
	if err != nil {
		return err
	}

	var n scriptNum
	if v1 <= v0 {
		n = 1
	}

	t.dstack.PushInt(n)
	return nil
}

// opcodeGreaterThanOrEqual treats the top two items on the data stack as
// integers.  When the second-to-top item is greater than or equal to the top
// item, they are replaced with a 1, otherwise a 0.
//
// Stack transformation: [... x1 x2] -> [... bool]
func opcodeGreaterThanOrEqual(op *ParsedOp, t *thread) error {
	v0, err := t.dstack.PopInt()
	if err != nil {
		return err
	}

	v1, err := t.dstack.PopInt()
	if err != nil {
		return err
	}

	var n scriptNum
	if v1 >= v0 {
		n = 1
	}

	t.dstack.PushInt(n)
	return nil
}

// opcodeMin treats the top two items on the data stack as integers and replaces
// them with the minimum of the two.
//
// Stack transformation: [... x1 x2] -> [... min(x1, x2)]
func opcodeMin(op *ParsedOp, t *thread) error {
	v0, err := t.dstack.PopInt()
	if err != nil {
		return err
	}

	v1, err := t.dstack.PopInt()
	if err != nil {
		return err
	}

	n := v0
	if v1 < v0 {
		n = v1
	}

	t.dstack.PushInt(n)
	return nil
}

// opcodeMax treats the top two items on the data stack as integers and replaces
// them with the maximum of the two.
//
// Stack transformation: [... x1 x2] -> [... max(x1, x2)]
func opcodeMax(op *ParsedOp, t *thread) error {
	v0, err := t.dstack.PopInt()
	if err != nil {
		return err
	}

	v1, err := t.dstack.PopInt()
	if err != nil {
		return err
	}

	n := v0
	if v1 > v0 {
		n = v1
	}

	t.dstack.PushInt(n)
	return nil
}

// opcodeWithin treats the top 3 items on the data stack as integers.  When the
// value to test is within the specified range (left inclusive), they are
// replaced with a 1, otherwise a 0.
//
// The top item is the max value, the second-top-item is the minimum value, and
// the third-to-top item is the value to test.
//
// Stack transformation: [... x1 min max] -> [... bool]
func opcodeWithin(op *ParsedOp, t *thread) error {
	maxVal, err := t.dstack.PopInt()
	if err != nil {
		return err
	}

	minVal, err := t.dstack.PopInt()
	if err != nil {
		return err
	}

	x, err := t.dstack.PopInt()
	if err != nil {
		return err
	}

	var n int
	if minVal <= x && x < maxVal {
		n = 1
	}

	t.dstack.PushInt(scriptNum(n))
	return nil
}

// calcHash calculates the hash of hasher over buf.
func calcHash(buf []byte, hasher hash.Hash) []byte {
	hasher.Write(buf)
	return hasher.Sum(nil)
}

// opcodeRipemd160 treats the top item of the data stack as raw bytes and
// replaces it with ripemd160(data).
//
// Stack transformation: [... x1] -> [... ripemd160(x1)]
func opcodeRipemd160(op *ParsedOp, t *thread) error {
	buf, err := t.dstack.PopByteArray()
	if err != nil {
		return err
	}

	t.dstack.PushByteArray(calcHash(buf, ripemd160.New()))
	return nil
}

// opcodeSha1 treats the top item of the data stack as raw bytes and replaces it
// with sha1(data).
//
// Stack transformation: [... x1] -> [... sha1(x1)]
func opcodeSha1(op *ParsedOp, t *thread) error {
	buf, err := t.dstack.PopByteArray()
	if err != nil {
		return err
	}

	hash := sha1.Sum(buf) // nolint:gosec // operation is for sha1
	t.dstack.PushByteArray(hash[:])
	return nil
}

// opcodeSha256 treats the top item of the data stack as raw bytes and replaces
// it with sha256(data).
//
// Stack transformation: [... x1] -> [... sha256(x1)]
func opcodeSha256(op *ParsedOp, t *thread) error {
	buf, err := t.dstack.PopByteArray()
	if err != nil {
		return err
	}

	hash := sha256.Sum256(buf)
	t.dstack.PushByteArray(hash[:])
	return nil
}

// opcodeHash160 treats the top item of the data stack as raw bytes and replaces
// it with ripemd160(sha256(data)).
//
// Stack transformation: [... x1] -> [... ripemd160(sha256(x1))]
func opcodeHash160(op *ParsedOp, t *thread) error {
	buf, err := t.dstack.PopByteArray()
	if err != nil {
		return err
	}

	hash := sha256.Sum256(buf)
	t.dstack.PushByteArray(calcHash(hash[:], ripemd160.New()))
	return nil
}

// opcodeHash256 treats the top item of the data stack as raw bytes and replaces
// it with sha256(sha256(data)).
//
// Stack transformation: [... x1] -> [... sha256(sha256(x1))]
func opcodeHash256(op *ParsedOp, t *thread) error {
	buf, err := t.dstack.PopByteArray()
	if err != nil {
		return err
	}

	t.dstack.PushByteArray(crypto.Sha256d(buf))
	return nil
}

// opcodeCodeSeparator stores the current script offset as the most recently
// seen bscript.OpCODESEPARATOR which is used during signature checking.
//
// This opcode does not change the contents of the data stack.
func opcodeCodeSeparator(op *ParsedOp, t *thread) error {
	t.lastCodeSep = t.scriptOff
	return nil
}

// opcodeCheckSig treats the top 2 items on the stack as a public key and a
// signature and replaces them with a bool which indicates if the signature was
// successfully verified.
//
// The process of verifying a signature requires calculating a signature hash in
// the same way the transaction signer did.  It involves hashing portions of the
// transaction based on the hash type byte (which is the final byte of the
// signature) and the portion of the script starting from the most recent
// bscript.OpCODESEPARATOR (or the beginning of the script if there are none) to the
// end of the script (with any other bscript.OpCODESEPARATORs removed).  Once this
// "script hash" is calculated, the signature is checked using standard
// cryptographic methods against the provided public key.
//
// Stack transformation: [... signature pubkey] -> [... bool]
func opcodeCheckSig(op *ParsedOp, t *thread) error {
	pkBytes, err := t.dstack.PopByteArray()
	if err != nil {
		return err
	}

	fullSigBytes, err := t.dstack.PopByteArray()
	if err != nil {
		return err
	}

	// The signature actually needs needs to be longer than this, but at
	// least 1 byte is needed for the hash type below.  The full length is
	// checked depending on the script flags and upon parsing the signature.
	if len(fullSigBytes) < 1 {
		t.dstack.PushBool(false)
		return nil
	}

	// Trim off hashtype from the signature string and check if the
	// signature and pubkey conform to the strict encoding requirements
	// depending on the flags.
	//
	// NOTE: When the strict encoding flags are set, any errors in the
	// signature or public encoding here result in an immediate script error
	// (and thus no result bool is pushed to the data stack).  This differs
	// from the logic below where any errors in parsing the signature is
	// treated as the signature failure resulting in false being pushed to
	// the data stack.  This is required because the more general script
	// validation consensus rules do not have the new strict encoding
	// requirements enabled by the flags.
	shf := sighash.Flag(fullSigBytes[len(fullSigBytes)-1])
	sigBytes := fullSigBytes[:len(fullSigBytes)-1]
	if err = t.checkHashTypeEncoding(shf); err != nil {
		return err
	}
	if err = t.checkSignatureEncoding(sigBytes); err != nil {
		return err
	}
	if err = t.checkPubKeyEncoding(pkBytes); err != nil {
		return err
	}

	// Get script starting from the most recent bscript.OpCODESEPARATOR.
	subScript := t.subScript()

	// Generate the signature hash based on the signature hash type.
	var hash []byte

	// Remove the signature since there is no way for a signature
	// to sign itself.
	if !t.hasFlag(ScriptEnableSighashForkID) || !shf.Has(sighash.ForkID) {
		subScript = subScript.removeOpcodeByData(fullSigBytes)
		subScript = subScript.removeOpcode(bscript.OpCODESEPARATOR)
	}

	up, err := t.scriptParser.Unparse(subScript)
	if err != nil {
		return err
	}

	txCopy := t.tx.Clone()
	txCopy.Inputs[t.inputIdx].PreviousTxScript = up

	hash, err = txCopy.CalcInputSignatureHash(uint32(t.inputIdx), shf)
	if err != nil {
		t.dstack.PushBool(false)
		return err
	}

	pubKey, err := bec.ParsePubKey(pkBytes, bec.S256())
	if err != nil {
		t.dstack.PushBool(false)
		return nil //nolint:nilerr // only need a false push in this case
	}

	var signature *bec.Signature
	if t.hasFlag(ScriptVerifyStrictEncoding) || t.hasFlag(ScriptVerifyDERSignatures) {
		signature, err = bec.ParseDERSignature(sigBytes, bec.S256())
	} else {
		signature, err = bec.ParseSignature(sigBytes, bec.S256())
	}
	if err != nil {
		t.dstack.PushBool(false)
		return nil //nolint:nilerr // only need a false push in this case
	}

	ok := signature.Verify(hash, pubKey)
	if !ok && t.hasFlag(ScriptVerifyNullFail) && len(sigBytes) > 0 {
		return scriptError(ErrNullFail, "signature not empty on failed checksig")
	}

	t.dstack.PushBool(ok)
	return nil
}

// opcodeCheckSigVerify is a combination of opcodeCheckSig and opcodeVerify.
// The opcodeCheckSig function is invoked followed by opcodeVerify.  See the
// documentation for each of those opcodes for more details.
//
// Stack transformation: signature pubkey] -> [... bool] -> [...]
func opcodeCheckSigVerify(op *ParsedOp, t *thread) error {
	if err := opcodeCheckSig(op, t); err != nil {
		return err
	}

	return abstractVerify(op, t, ErrCheckSigVerify)
}

// parsedSigInfo houses a raw signature along with its parsed form and a flag
// for whether or not it has already been parsed.  It is used to prevent parsing
// the same signature multiple times when verifying a multisig.
type parsedSigInfo struct {
	signature       []byte
	parsedSignature *bec.Signature
	parsed          bool
}

// opcodeCheckMultiSig treats the top item on the stack as an integer number of
// public keys, followed by that many entries as raw data representing the public
// keys, followed by the integer number of signatures, followed by that many
// entries as raw data representing the signatures.
//
// Due to a bug in the original Satoshi client implementation, an additional
// dummy argument is also required by the consensus rules, although it is not
// used.  The dummy value SHOULD be an bscript.Op0, although that is not required by
// the consensus rules.  When the ScriptStrictMultiSig flag is set, it must be
// bscript.Op0.
//
// All of the aforementioned stack items are replaced with a bool which
// indicates if the requisite number of signatures were successfully verified.
//
// See the opcodeCheckSigVerify documentation for more details about the process
// for verifying each signature.
//
// Stack transformation:
// [... dummy [sig ...] numsigs [pubkey ...] numpubkeys] -> [... bool]
func opcodeCheckMultiSig(op *ParsedOp, t *thread) error {
	numKeys, err := t.dstack.PopInt()
	if err != nil {
		return err
	}

	numPubKeys := int(numKeys.Int32())
	if numPubKeys < 0 {
		return scriptError(ErrInvalidPubKeyCount, "number of pubkeys %d is negative", numPubKeys)
	}
	if numPubKeys > t.cfg.MaxPubKeysPerMultiSig() {
		return scriptError(ErrInvalidPubKeyCount, "too many pubkeys: %d > %d", numPubKeys, t.cfg.MaxPubKeysPerMultiSig())
	}
	t.numOps += numPubKeys
	if t.numOps > t.cfg.MaxOps() {
		return scriptError(ErrTooManyOperations, "exceeded max operation limit of %d", t.cfg.MaxOps())
	}

	pubKeys := make([][]byte, 0, numPubKeys)
	for i := 0; i < numPubKeys; i++ {
		pubKey, err := t.dstack.PopByteArray() //nolint:govet // ignore shadowed error
		if err != nil {
			return err
		}
		pubKeys = append(pubKeys, pubKey)
	}

	numSigs, err := t.dstack.PopInt()
	if err != nil {
		return err
	}

	numSignatures := int(numSigs.Int32())
	if numSignatures < 0 {
		return scriptError(ErrInvalidSignatureCount, "number of signatures %d is negative", numSignatures)
	}
	if numSignatures > numPubKeys {
		return scriptError(ErrInvalidSignatureCount, "more signatures than pubkeys: %d > %d", numSignatures, numPubKeys)
	}

	signatures := make([]*parsedSigInfo, 0, numSignatures)
	for i := 0; i < numSignatures; i++ {
		signature, err := t.dstack.PopByteArray() //nolint:govet // ignore shadowed error
		if err != nil {
			return err
		}
		sigInfo := &parsedSigInfo{signature: signature}
		signatures = append(signatures, sigInfo)
	}

	// A bug in the original Satoshi client implementation means one more
	// stack value than should be used must be popped.  Unfortunately, this
	// buggy behaviour is now part of the consensus and a hard fork would be
	// required to fix it.
	dummy, err := t.dstack.PopByteArray()
	if err != nil {
		return err
	}

	// Since the dummy argument is otherwise not checked, it could be any
	// value which unfortunately provides a source of malleability.  Thus,
	// there is a script flag to force an error when the value is NOT 0.
	if t.hasFlag(ScriptStrictMultiSig) && len(dummy) != 0 {
		return scriptError(ErrSigNullDummy, "multisig dummy argument has length %d instead of 0", len(dummy))
	}

	// Get script starting from the most recent bscript.OpCODESEPARATOR.
	script := t.subScript()

	for _, sigInfo := range signatures {
		script = script.removeOpcodeByData(sigInfo.signature)
		script = script.removeOpcode(bscript.OpCODESEPARATOR)
	}

	success := true
	numPubKeys++
	pubKeyIdx := -1
	signatureIdx := 0
	for numSignatures > 0 {
		// When there are more signatures than public keys remaining,
		// there is no way to succeed since too many signatures are
		// invalid, so exit early.
		pubKeyIdx++
		numPubKeys--
		if numSignatures > numPubKeys {
			success = false
			break
		}

		sigInfo := signatures[signatureIdx]
		pubKey := pubKeys[pubKeyIdx]

		// The order of the signature and public key evaluation is
		// important here since it can be distinguished by an
		// bscript.OpCHECKMULTISIG NOT when the strict encoding flag is set.

		rawSig := sigInfo.signature
		if len(rawSig) == 0 {
			// Skip to the next pubkey if signature is empty.
			continue
		}

		// Split the signature into hash type and signature components.
		shf := sighash.Flag(rawSig[len(rawSig)-1])
		signature := rawSig[:len(rawSig)-1]

		// Only parse and check the signature encoding once.
		var parsedSig *bec.Signature
		if !sigInfo.parsed {
			if err := t.checkHashTypeEncoding(shf); err != nil {
				return err
			}
			if err := t.checkSignatureEncoding(signature); err != nil {
				return err
			}

			// Parse the signature.
			var err error
			if t.hasFlag(ScriptVerifyStrictEncoding) ||
				t.hasFlag(ScriptVerifyDERSignatures) {

				parsedSig, err = bec.ParseDERSignature(signature,
					bec.S256())
			} else {
				parsedSig, err = bec.ParseSignature(signature,
					bec.S256())
			}
			sigInfo.parsed = true
			if err != nil {
				continue
			}
			sigInfo.parsedSignature = parsedSig
		} else {
			// Skip to the next pubkey if the signature is invalid.
			if sigInfo.parsedSignature == nil {
				continue
			}

			// Use the already parsed signature.
			parsedSig = sigInfo.parsedSignature
		}

		if err := t.checkPubKeyEncoding(pubKey); err != nil {
			return err
		}

		// Parse the pubkey.
		parsedPubKey, err := bec.ParsePubKey(pubKey, bec.S256())
		if err != nil {
			continue
		}

		up, err := t.scriptParser.Unparse(script)
		if err != nil {
			t.dstack.PushBool(false)
			return nil //nolint:nilerr // only need a false push in this case
		}

		// Generate the signature hash based on the signature hash type.
		txCopy := t.tx.Clone()
		txCopy.Inputs[t.inputIdx].PreviousTxScript = up

		signatureHash, err := txCopy.CalcInputSignatureHash(uint32(t.inputIdx), shf)
		if err != nil {
			t.dstack.PushBool(false)
			return nil //nolint:nilerr // only need a false push in this case
		}

		if ok := parsedSig.Verify(signatureHash, parsedPubKey); ok {
			// PubKey verified, move on to the next signature.
			signatureIdx++
			numSignatures--
		}
	}

	if !success && t.hasFlag(ScriptVerifyNullFail) {
		for _, sig := range signatures {
			if len(sig.signature) > 0 {
				return scriptError(ErrNullFail, "not all signatures empty on failed checkmultisig")
			}
		}
	}

	t.dstack.PushBool(success)
	return nil
}

// opcodeCheckMultiSigVerify is a combination of opcodeCheckMultiSig and
// opcodeVerify.  The opcodeCheckMultiSig is invoked followed by opcodeVerify.
// See the documentation for each of those opcodes for more details.
//
// Stack transformation:
// [... dummy [sig ...] numsigs [pubkey ...] numpubkeys] -> [... bool] -> [...]
func opcodeCheckMultiSigVerify(op *ParsedOp, t *thread) error {
	if err := opcodeCheckMultiSig(op, t); err != nil {
		return err
	}

	return abstractVerify(op, t, ErrCheckMultiSigVerify)
}

func success() Error {
	return scriptError(ErrOK, "success")
}
