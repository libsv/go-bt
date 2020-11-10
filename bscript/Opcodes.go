package bscript

// BitCoin Script constants.
// See https://wiki.bitcoinsv.io/index.php/Opcodes_used_in_Bitcoin_Script
const (
	Op0                   byte = 0x00 // 0
	OpZERO                byte = 0x00 // 0
	OpFALSE               byte = 0x00 // 0
	OpDATA1               byte = 0x01 // 1
	OpDATA2               byte = 0x02 // 2
	OpDATA3               byte = 0x03 // 3
	OpDATA4               byte = 0x04 // 4
	OpDATA5               byte = 0x05 // 5
	OpDATA6               byte = 0x06 // 6
	OpDATA7               byte = 0x07 // 7
	OpDATA8               byte = 0x08 // 8
	OpDATA9               byte = 0x09 // 9
	OpDATA10              byte = 0x0a // 10
	OpDATA11              byte = 0x0b // 11
	OpDATA12              byte = 0x0c // 12
	OpDATA13              byte = 0x0d // 13
	OpDATA14              byte = 0x0e // 14
	OpDATA15              byte = 0x0f // 15
	OpDATA16              byte = 0x10 // 16
	OpDATA17              byte = 0x11 // 17
	OpDATA18              byte = 0x12 // 18
	OpDATA19              byte = 0x13 // 19
	OpDATA20              byte = 0x14 // 20
	OpDATA21              byte = 0x15 // 21
	OpDATA22              byte = 0x16 // 22
	OpDATA23              byte = 0x17 // 23
	OpDATA24              byte = 0x18 // 24
	OpDATA25              byte = 0x19 // 25
	OpDATA26              byte = 0x1a // 26
	OpDATA27              byte = 0x1b // 27
	OpDATA28              byte = 0x1c // 28
	OpDATA29              byte = 0x1d // 29
	OpDATA30              byte = 0x1e // 30
	OpDATA31              byte = 0x1f // 31
	OpDATA32              byte = 0x20 // 32
	OpDATA33              byte = 0x21 // 33
	OpDATA34              byte = 0x22 // 34
	OpDATA35              byte = 0x23 // 35
	OpDATA36              byte = 0x24 // 36
	OpDATA37              byte = 0x25 // 37
	OpDATA38              byte = 0x26 // 38
	OpDATA39              byte = 0x27 // 39
	OpDATA40              byte = 0x28 // 40
	OpDATA41              byte = 0x29 // 41
	OpDATA42              byte = 0x2a // 42
	OpDATA43              byte = 0x2b // 43
	OpDATA44              byte = 0x2c // 44
	OpDATA45              byte = 0x2d // 45
	OpDATA46              byte = 0x2e // 46
	OpDATA47              byte = 0x2f // 47
	OpDATA48              byte = 0x30 // 48
	OpDATA49              byte = 0x31 // 49
	OpDATA50              byte = 0x32 // 50
	OpDATA51              byte = 0x33 // 51
	OpDATA52              byte = 0x34 // 52
	OpDATA53              byte = 0x35 // 53
	OpDATA54              byte = 0x36 // 54
	OpDATA55              byte = 0x37 // 55
	OpDATA56              byte = 0x38 // 56
	OpDATA57              byte = 0x39 // 57
	OpDATA58              byte = 0x3a // 58
	OpDATA59              byte = 0x3b // 59
	OpDATA60              byte = 0x3c // 60
	OpDATA61              byte = 0x3d // 61
	OpDATA62              byte = 0x3e // 62
	OpDATA63              byte = 0x3f // 63
	OpDATA64              byte = 0x40 // 64
	OpDATA65              byte = 0x41 // 65
	OpDATA66              byte = 0x42 // 66
	OpDATA67              byte = 0x43 // 67
	OpDATA68              byte = 0x44 // 68
	OpDATA69              byte = 0x45 // 69
	OpDATA70              byte = 0x46 // 70
	OpDATA71              byte = 0x47 // 71
	OpDATA72              byte = 0x48 // 72
	OpDATA73              byte = 0x49 // 73
	OpDATA74              byte = 0x4a // 74
	OpDATA75              byte = 0x4b // 75
	OpPUSHDATA1           byte = 0x4c // 76
	OpPUSHDATA2           byte = 0x4d // 77
	OpPUSHDATA4           byte = 0x4e // 78
	Op1NEGATE             byte = 0x4f // 79
	OpRESERVED            byte = 0x50 // 80
	OpBASE                byte = 0x50 // 80
	Op1                   byte = 0x51 // 81
	OpONE                 byte = 0x51 // 81
	OpTRUE                byte = 0x51 // 81
	Op2                   byte = 0x52 // 82
	Op3                   byte = 0x53 // 83
	Op4                   byte = 0x54 // 84
	Op5                   byte = 0x55 // 85
	Op6                   byte = 0x56 // 86
	Op7                   byte = 0x57 // 87
	Op8                   byte = 0x58 // 88
	Op9                   byte = 0x59 // 89
	Op10                  byte = 0x5a // 90
	Op11                  byte = 0x5b // 91
	Op12                  byte = 0x5c // 92
	Op13                  byte = 0x5d // 93
	Op14                  byte = 0x5e // 94
	Op15                  byte = 0x5f // 95
	Op16                  byte = 0x60 // 96
	OpNOP                 byte = 0x61 // 97
	OpVER                 byte = 0x62 // 98
	OpIF                  byte = 0x63 // 99
	OpNOTIF               byte = 0x64 // 100
	OpVERIF               byte = 0x65 // 101
	OpVERNOTIF            byte = 0x66 // 102
	OpELSE                byte = 0x67 // 103
	OpENDIF               byte = 0x68 // 104
	OpVERIFY              byte = 0x69 // 105
	OpRETURN              byte = 0x6a // 106
	OpTOALTSTACK          byte = 0x6b // 107
	OpFROMALTSTACK        byte = 0x6c // 108
	Op2DROP               byte = 0x6d // 109
	Op2DUP                byte = 0x6e // 110
	Op3DUP                byte = 0x6f // 111
	Op2OVER               byte = 0x70 // 112
	Op2ROT                byte = 0x71 // 113
	Op2SWAP               byte = 0x72 // 114
	OpIFDUP               byte = 0x73 // 115
	OpDEPTH               byte = 0x74 // 116
	OpDROP                byte = 0x75 // 117
	OpDUP                 byte = 0x76 // 118
	OpNIP                 byte = 0x77 // 119
	OpOVER                byte = 0x78 // 120
	OpPICK                byte = 0x79 // 121
	OpROLL                byte = 0x7a // 122
	OpROT                 byte = 0x7b // 123
	OpSWAP                byte = 0x7c // 124
	OpTUCK                byte = 0x7d // 125
	OpCAT                 byte = 0x7e // 126
	OpSPLIT               byte = 0x7f // 127
	OpNUM2BIN             byte = 0x80 // 128
	OpBIN2NUM             byte = 0x81 // 129
	OpSIZE                byte = 0x82 // 130
	OpINVERT              byte = 0x83 // 131
	OpAND                 byte = 0x84 // 132
	OpOR                  byte = 0x85 // 133
	OpXOR                 byte = 0x86 // 134
	OpEQUAL               byte = 0x87 // 135
	OpEQUALVERIFY         byte = 0x88 // 136
	OpRESERVED1           byte = 0x89 // 137
	OpRESERVED2           byte = 0x8a // 138
	Op1ADD                byte = 0x8b // 139
	Op1SUB                byte = 0x8c // 140
	Op2MUL                byte = 0x8d // 141
	Op2DIV                byte = 0x8e // 142
	OpNEGATE              byte = 0x8f // 143
	OpABS                 byte = 0x90 // 144
	OpNOT                 byte = 0x91 // 145
	Op0NOTEQUAL           byte = 0x92 // 146
	OpADD                 byte = 0x93 // 147
	OpSUB                 byte = 0x94 // 148
	OpMUL                 byte = 0x95 // 149
	OpDIV                 byte = 0x96 // 150
	OpMOD                 byte = 0x97 // 151
	OpLSHIFT              byte = 0x98 // 152
	OpRSHIFT              byte = 0x99 // 153
	OpBOOLAND             byte = 0x9a // 154
	OpBOOLOR              byte = 0x9b // 155
	OpNUMEQUAL            byte = 0x9c // 156
	OpNUMEQUALVERIFY      byte = 0x9d // 157
	OpNUMNOTEQUAL         byte = 0x9e // 158
	OpLESSTHAN            byte = 0x9f // 159
	OpGREATERTHAN         byte = 0xa0 // 160
	OpLESSTHANOREQUAL     byte = 0xa1 // 161
	OpGREATERTHANOREQUAL  byte = 0xa2 // 162
	OpMIN                 byte = 0xa3 // 163
	OpMAX                 byte = 0xa4 // 164
	OpWITHIN              byte = 0xa5 // 165
	OpRIPEMD160           byte = 0xa6 // 166
	OpSHA1                byte = 0xa7 // 167
	OpSHA256              byte = 0xa8 // 168
	OpHASH160             byte = 0xa9 // 169
	OpHASH256             byte = 0xaa // 170
	OpCODESEPARATOR       byte = 0xab // 171
	OpCHECKSIG            byte = 0xac // 172
	OpCHECKSIGVERIFY      byte = 0xad // 173
	OpCHECKMULTISIG       byte = 0xae // 174
	OpCHECKMULTISIGVERIFY byte = 0xaf // 175
	OpNOP1                byte = 0xb0 // 176
	OpNOP2                byte = 0xb1 // 177
	OpNOP3                byte = 0xb2 // 178
	OpNOP4                byte = 0xb3 // 179
	OpNOP5                byte = 0xb4 // 180
	OpNOP6                byte = 0xb5 // 181
	OpNOP7                byte = 0xb6 // 182
	OpNOP8                byte = 0xb7 // 183
	OpNOP9                byte = 0xb8 // 184
	OpNOP10               byte = 0xb9 // 185
	OpUNKNOWN186          byte = 0xba // 186
	OpUNKNOWN187          byte = 0xbb // 187
	OpUNKNOWN188          byte = 0xbc // 188
	OpUNKNOWN189          byte = 0xbd // 189
	OpUNKNOWN190          byte = 0xbe // 190
	OpUNKNOWN191          byte = 0xbf // 191
	OpUNKNOWN192          byte = 0xc0 // 192
	OpUNKNOWN193          byte = 0xc1 // 193
	OpUNKNOWN194          byte = 0xc2 // 194
	OpUNKNOWN195          byte = 0xc3 // 195
	OpUNKNOWN196          byte = 0xc4 // 196
	OpUNKNOWN197          byte = 0xc5 // 197
	OpUNKNOWN198          byte = 0xc6 // 198
	OpUNKNOWN199          byte = 0xc7 // 199
	OpUNKNOWN200          byte = 0xc8 // 200
	OpUNKNOWN201          byte = 0xc9 // 201
	OpUNKNOWN202          byte = 0xca // 202
	OpUNKNOWN203          byte = 0xcb // 203
	OpUNKNOWN204          byte = 0xcc // 204
	OpUNKNOWN205          byte = 0xcd // 205
	OpUNKNOWN206          byte = 0xce // 206
	OpUNKNOWN207          byte = 0xcf // 207
	OpUNKNOWN208          byte = 0xd0 // 208
	OpUNKNOWN209          byte = 0xd1 // 209
	OpUNKNOWN210          byte = 0xd2 // 210
	OpUNKNOWN211          byte = 0xd3 // 211
	OpUNKNOWN212          byte = 0xd4 // 212
	OpUNKNOWN213          byte = 0xd5 // 213
	OpUNKNOWN214          byte = 0xd6 // 214
	OpUNKNOWN215          byte = 0xd7 // 215
	OpUNKNOWN216          byte = 0xd8 // 216
	OpUNKNOWN217          byte = 0xd9 // 217
	OpUNKNOWN218          byte = 0xda // 218
	OpUNKNOWN219          byte = 0xdb // 219
	OpUNKNOWN220          byte = 0xdc // 220
	OpUNKNOWN221          byte = 0xdd // 221
	OpUNKNOWN222          byte = 0xde // 222
	OpUNKNOWN223          byte = 0xdf // 223
	OpUNKNOWN224          byte = 0xe0 // 224
	OpUNKNOWN225          byte = 0xe1 // 225
	OpUNKNOWN226          byte = 0xe2 // 226
	OpUNKNOWN227          byte = 0xe3 // 227
	OpUNKNOWN228          byte = 0xe4 // 228
	OpUNKNOWN229          byte = 0xe5 // 229
	OpUNKNOWN230          byte = 0xe6 // 230
	OpUNKNOWN231          byte = 0xe7 // 231
	OpUNKNOWN232          byte = 0xe8 // 232
	OpUNKNOWN233          byte = 0xe9 // 233
	OpUNKNOWN234          byte = 0xea // 234
	OpUNKNOWN235          byte = 0xeb // 235
	OpUNKNOWN236          byte = 0xec // 236
	OpUNKNOWN237          byte = 0xed // 237
	OpUNKNOWN238          byte = 0xee // 238
	OpUNKNOWN239          byte = 0xef // 239
	OpUNKNOWN240          byte = 0xf0 // 240
	OpUNKNOWN241          byte = 0xf1 // 241
	OpUNKNOWN242          byte = 0xf2 // 242
	OpUNKNOWN243          byte = 0xf3 // 243
	OpUNKNOWN244          byte = 0xf4 // 244
	OpUNKNOWN245          byte = 0xf5 // 245
	OpUNKNOWN246          byte = 0xf6 // 246
	OpUNKNOWN247          byte = 0xf7 // 247
	OpUNKNOWN248          byte = 0xf8 // 248
	OpUNKNOWN249          byte = 0xf9 // 249
	OpSMALLINTEGER        byte = 0xfa // 250 - bitcoin core internal
	OpPUBKEYS             byte = 0xfb // 251 - bitcoin core internal
	OpUNKNOWN252          byte = 0xfc // 252
	OpPUBKEYHASH          byte = 0xfd // 253 - bitcoin core internal
	OpPUBKEY              byte = 0xfe // 254 - bitcoin core internal
	OpINVALIDOPCODE       byte = 0xff // 255 - bitcoin core
)

var opCodeStrings = map[string]byte{
	"OP_0":                   Op0,
	"OP_ZERO":                OpZERO,
	"OP_FALSE":               OpFALSE,
	"OP_DATA_1":              OpDATA1,
	"OP_DATA_2":              OpDATA2,
	"OP_DATA_3":              OpDATA3,
	"OP_DATA_4":              OpDATA4,
	"OP_DATA_5":              OpDATA5,
	"OP_DATA_6":              OpDATA6,
	"OP_DATA_7":              OpDATA7,
	"OP_DATA_8":              OpDATA8,
	"OP_DATA_9":              OpDATA9,
	"OP_DATA_10":             OpDATA10,
	"OP_DATA_11":             OpDATA11,
	"OP_DATA_12":             OpDATA12,
	"OP_DATA_13":             OpDATA13,
	"OP_DATA_14":             OpDATA14,
	"OP_DATA_15":             OpDATA15,
	"OP_DATA_16":             OpDATA16,
	"OP_DATA_17":             OpDATA17,
	"OP_DATA_18":             OpDATA18,
	"OP_DATA_19":             OpDATA19,
	"OP_DATA_20":             OpDATA20,
	"OP_DATA_21":             OpDATA21,
	"OP_DATA_22":             OpDATA22,
	"OP_DATA_23":             OpDATA23,
	"OP_DATA_24":             OpDATA24,
	"OP_DATA_25":             OpDATA25,
	"OP_DATA_26":             OpDATA26,
	"OP_DATA_27":             OpDATA27,
	"OP_DATA_28":             OpDATA28,
	"OP_DATA_29":             OpDATA29,
	"OP_DATA_30":             OpDATA30,
	"OP_DATA_31":             OpDATA31,
	"OP_DATA_32":             OpDATA32,
	"OP_DATA_33":             OpDATA33,
	"OP_DATA_34":             OpDATA34,
	"OP_DATA_35":             OpDATA35,
	"OP_DATA_36":             OpDATA36,
	"OP_DATA_37":             OpDATA37,
	"OP_DATA_38":             OpDATA38,
	"OP_DATA_39":             OpDATA39,
	"OP_DATA_40":             OpDATA40,
	"OP_DATA_41":             OpDATA41,
	"OP_DATA_42":             OpDATA42,
	"OP_DATA_43":             OpDATA43,
	"OP_DATA_44":             OpDATA44,
	"OP_DATA_45":             OpDATA45,
	"OP_DATA_46":             OpDATA46,
	"OP_DATA_47":             OpDATA47,
	"OP_DATA_48":             OpDATA48,
	"OP_DATA_49":             OpDATA49,
	"OP_DATA_50":             OpDATA50,
	"OP_DATA_51":             OpDATA51,
	"OP_DATA_52":             OpDATA52,
	"OP_DATA_53":             OpDATA53,
	"OP_DATA_54":             OpDATA54,
	"OP_DATA_55":             OpDATA55,
	"OP_DATA_56":             OpDATA56,
	"OP_DATA_57":             OpDATA57,
	"OP_DATA_58":             OpDATA58,
	"OP_DATA_59":             OpDATA59,
	"OP_DATA_60":             OpDATA60,
	"OP_DATA_61":             OpDATA61,
	"OP_DATA_62":             OpDATA62,
	"OP_DATA_63":             OpDATA63,
	"OP_DATA_64":             OpDATA64,
	"OP_DATA_65":             OpDATA65,
	"OP_DATA_66":             OpDATA66,
	"OP_DATA_67":             OpDATA67,
	"OP_DATA_68":             OpDATA68,
	"OP_DATA_69":             OpDATA69,
	"OP_DATA_70":             OpDATA70,
	"OP_DATA_71":             OpDATA71,
	"OP_DATA_72":             OpDATA72,
	"OP_DATA_73":             OpDATA73,
	"OP_DATA_74":             OpDATA74,
	"OP_DATA_75":             OpDATA75,
	"OP_PUSHDATA1":           OpPUSHDATA1,
	"OP_PUSHDATA2":           OpPUSHDATA2,
	"OP_PUSHDATA4":           OpPUSHDATA4,
	"OP_1NEGATE":             Op1NEGATE,
	"OP_RESERVED":            OpRESERVED,
	"OP_BASE":                OpBASE,
	"OP_1":                   Op1,
	"OP_ONE":                 OpONE,
	"OP_TRUE":                OpTRUE,
	"OP_2":                   Op2,
	"OP_3":                   Op3,
	"OP_4":                   Op4,
	"OP_5":                   Op5,
	"OP_6":                   Op6,
	"OP_7":                   Op7,
	"OP_8":                   Op8,
	"OP_9":                   Op9,
	"OP_10":                  Op10,
	"OP_11":                  Op11,
	"OP_12":                  Op12,
	"OP_13":                  Op13,
	"OP_14":                  Op14,
	"OP_15":                  Op15,
	"OP_16":                  Op16,
	"OP_NOP":                 OpNOP,
	"OP_VER":                 OpVER,
	"OP_IF":                  OpIF,
	"OP_NOTIF":               OpNOTIF,
	"OP_VERIF":               OpVERIF,
	"OP_VERNOTIF":            OpVERNOTIF,
	"OP_ELSE":                OpELSE,
	"OP_ENDIF":               OpENDIF,
	"OP_VERIFY":              OpVERIFY,
	"OP_RETURN":              OpRETURN,
	"OP_TOALTSTACK":          OpTOALTSTACK,
	"OP_FROMALTSTACK":        OpFROMALTSTACK,
	"OP_2DROP":               Op2DROP,
	"OP_2DUP":                Op2DUP,
	"OP_3DUP":                Op3DUP,
	"OP_2OVER":               Op2OVER,
	"OP_2ROT":                Op2ROT,
	"OP_2SWAP":               Op2SWAP,
	"OP_IFDUP":               OpIFDUP,
	"OP_DEPTH":               OpDEPTH,
	"OP_DROP":                OpDROP,
	"OP_DUP":                 OpDUP,
	"OP_NIP":                 OpNIP,
	"OP_OVER":                OpOVER,
	"OP_PICK":                OpPICK,
	"OP_ROLL":                OpROLL,
	"OP_ROT":                 OpROT,
	"OP_SWAP":                OpSWAP,
	"OP_TUCK":                OpTUCK,
	"OP_CAT":                 OpCAT,
	"OP_SPLIT":               OpSPLIT,
	"OP_NUM2BIN":             OpNUM2BIN,
	"OP_BIN2NUM":             OpBIN2NUM,
	"OP_SIZE":                OpSIZE,
	"OP_INVERT":              OpINVERT,
	"OP_AND":                 OpAND,
	"OP_OR":                  OpOR,
	"OP_XOR":                 OpXOR,
	"OP_EQUAL":               OpEQUAL,
	"OP_EQUALVERIFY":         OpEQUALVERIFY,
	"OP_RESERVED1":           OpRESERVED1,
	"OP_RESERVED2":           OpRESERVED2,
	"OP_1ADD":                Op1ADD,
	"OP_1SUB":                Op1SUB,
	"OP_2MUL":                Op2MUL,
	"OP_2DIV":                Op2DIV,
	"OP_NEGATE":              OpNEGATE,
	"OP_ABS":                 OpABS,
	"OP_NOT":                 OpNOT,
	"OP_0NOTEQUAL":           Op0NOTEQUAL,
	"OP_ADD":                 OpADD,
	"OP_SUB":                 OpSUB,
	"OP_MUL":                 OpMUL,
	"OP_DIV":                 OpDIV,
	"OP_MOD":                 OpMOD,
	"OP_LSHIFT":              OpLSHIFT,
	"OP_RSHIFT":              OpRSHIFT,
	"OP_BOOLAND":             OpBOOLAND,
	"OP_BOOLOR":              OpBOOLOR,
	"OP_NUMEQUAL":            OpNUMEQUAL,
	"OP_NUMEQUALVERIFY":      OpNUMEQUALVERIFY,
	"OP_NUMNOTEQUAL":         OpNUMNOTEQUAL,
	"OP_LESSTHAN":            OpLESSTHAN,
	"OP_GREATERTHAN":         OpGREATERTHAN,
	"OP_LESSTHANOREQUAL":     OpLESSTHANOREQUAL,
	"OP_GREATERTHANOREQUAL":  OpGREATERTHANOREQUAL,
	"OP_MIN":                 OpMIN,
	"OP_MAX":                 OpMAX,
	"OP_WITHIN":              OpWITHIN,
	"OP_RIPEMD160":           OpRIPEMD160,
	"OP_SHA1":                OpSHA1,
	"OP_SHA256":              OpSHA256,
	"OP_HASH160":             OpHASH160,
	"OP_HASH256":             OpHASH256,
	"OP_CODESEPARATOR":       OpCODESEPARATOR,
	"OP_CHECKSIG":            OpCHECKSIG,
	"OP_CHECKSIGVERIFY":      OpCHECKSIGVERIFY,
	"OP_CHECKMULTISIG":       OpCHECKMULTISIG,
	"OP_CHECKMULTISIGVERIFY": OpCHECKMULTISIGVERIFY,
	"OP_NOP1":                OpNOP1,
	"OP_NOP2":                OpNOP2,
	"OP_NOP3":                OpNOP3,
	"OP_NOP4":                OpNOP4,
	"OP_NOP5":                OpNOP5,
	"OP_NOP6":                OpNOP6,
	"OP_NOP7":                OpNOP7,
	"OP_NOP8":                OpNOP8,
	"OP_NOP9":                OpNOP9,
	"OP_NOP10":               OpNOP10,
	"OP_UNKNOWN186":          OpUNKNOWN186,
	"OP_UNKNOWN187":          OpUNKNOWN187,
	"OP_UNKNOWN188":          OpUNKNOWN188,
	"OP_UNKNOWN189":          OpUNKNOWN189,
	"OP_UNKNOWN190":          OpUNKNOWN190,
	"OP_UNKNOWN191":          OpUNKNOWN191,
	"OP_UNKNOWN192":          OpUNKNOWN192,
	"OP_UNKNOWN193":          OpUNKNOWN193,
	"OP_UNKNOWN194":          OpUNKNOWN194,
	"OP_UNKNOWN195":          OpUNKNOWN195,
	"OP_UNKNOWN196":          OpUNKNOWN196,
	"OP_UNKNOWN197":          OpUNKNOWN197,
	"OP_UNKNOWN198":          OpUNKNOWN198,
	"OP_UNKNOWN199":          OpUNKNOWN199,
	"OP_UNKNOWN200":          OpUNKNOWN200,
	"OP_UNKNOWN201":          OpUNKNOWN201,
	"OP_UNKNOWN202":          OpUNKNOWN202,
	"OP_UNKNOWN203":          OpUNKNOWN203,
	"OP_UNKNOWN204":          OpUNKNOWN204,
	"OP_UNKNOWN205":          OpUNKNOWN205,
	"OP_UNKNOWN206":          OpUNKNOWN206,
	"OP_UNKNOWN207":          OpUNKNOWN207,
	"OP_UNKNOWN208":          OpUNKNOWN208,
	"OP_UNKNOWN209":          OpUNKNOWN209,
	"OP_UNKNOWN210":          OpUNKNOWN210,
	"OP_UNKNOWN211":          OpUNKNOWN211,
	"OP_UNKNOWN212":          OpUNKNOWN212,
	"OP_UNKNOWN213":          OpUNKNOWN213,
	"OP_UNKNOWN214":          OpUNKNOWN214,
	"OP_UNKNOWN215":          OpUNKNOWN215,
	"OP_UNKNOWN216":          OpUNKNOWN216,
	"OP_UNKNOWN217":          OpUNKNOWN217,
	"OP_UNKNOWN218":          OpUNKNOWN218,
	"OP_UNKNOWN219":          OpUNKNOWN219,
	"OP_UNKNOWN220":          OpUNKNOWN220,
	"OP_UNKNOWN221":          OpUNKNOWN221,
	"OP_UNKNOWN222":          OpUNKNOWN222,
	"OP_UNKNOWN223":          OpUNKNOWN223,
	"OP_UNKNOWN224":          OpUNKNOWN224,
	"OP_UNKNOWN225":          OpUNKNOWN225,
	"OP_UNKNOWN226":          OpUNKNOWN226,
	"OP_UNKNOWN227":          OpUNKNOWN227,
	"OP_UNKNOWN228":          OpUNKNOWN228,
	"OP_UNKNOWN229":          OpUNKNOWN229,
	"OP_UNKNOWN230":          OpUNKNOWN230,
	"OP_UNKNOWN231":          OpUNKNOWN231,
	"OP_UNKNOWN232":          OpUNKNOWN232,
	"OP_UNKNOWN233":          OpUNKNOWN233,
	"OP_UNKNOWN234":          OpUNKNOWN234,
	"OP_UNKNOWN235":          OpUNKNOWN235,
	"OP_UNKNOWN236":          OpUNKNOWN236,
	"OP_UNKNOWN237":          OpUNKNOWN237,
	"OP_UNKNOWN238":          OpUNKNOWN238,
	"OP_UNKNOWN239":          OpUNKNOWN239,
	"OP_UNKNOWN240":          OpUNKNOWN240,
	"OP_UNKNOWN241":          OpUNKNOWN241,
	"OP_UNKNOWN242":          OpUNKNOWN242,
	"OP_UNKNOWN243":          OpUNKNOWN243,
	"OP_UNKNOWN244":          OpUNKNOWN244,
	"OP_UNKNOWN245":          OpUNKNOWN245,
	"OP_UNKNOWN246":          OpUNKNOWN246,
	"OP_UNKNOWN247":          OpUNKNOWN247,
	"OP_UNKNOWN248":          OpUNKNOWN248,
	"OP_UNKNOWN249":          OpUNKNOWN249,
	"OP_SMALLINTEGER":        OpSMALLINTEGER,
	"OP_PUBKEYS":             OpPUBKEYS,
	"OP_UNKNOWN252":          OpUNKNOWN252,
	"OP_PUBKEYHASH":          OpPUBKEYHASH,
	"OP_PUBKEY":              OpPUBKEY,
	"OP_INVALIDOPCODE":       OpINVALIDOPCODE,
}

var opCodeValues = map[byte]string{
	OpFALSE:               "OP_FALSE",
	OpDATA1:               "OP_DATA_1",
	OpDATA2:               "OP_DATA_2",
	OpDATA3:               "OP_DATA_3",
	OpDATA4:               "OP_DATA_4",
	OpDATA5:               "OP_DATA_5",
	OpDATA6:               "OP_DATA_6",
	OpDATA7:               "OP_DATA_7",
	OpDATA8:               "OP_DATA_8",
	OpDATA9:               "OP_DATA_9",
	OpDATA10:              "OP_DATA_10",
	OpDATA11:              "OP_DATA_11",
	OpDATA12:              "OP_DATA_12",
	OpDATA13:              "OP_DATA_13",
	OpDATA14:              "OP_DATA_14",
	OpDATA15:              "OP_DATA_15",
	OpDATA16:              "OP_DATA_16",
	OpDATA17:              "OP_DATA_17",
	OpDATA18:              "OP_DATA_18",
	OpDATA19:              "OP_DATA_19",
	OpDATA20:              "OP_DATA_20",
	OpDATA21:              "OP_DATA_21",
	OpDATA22:              "OP_DATA_22",
	OpDATA23:              "OP_DATA_23",
	OpDATA24:              "OP_DATA_24",
	OpDATA25:              "OP_DATA_25",
	OpDATA26:              "OP_DATA_26",
	OpDATA27:              "OP_DATA_27",
	OpDATA28:              "OP_DATA_28",
	OpDATA29:              "OP_DATA_29",
	OpDATA30:              "OP_DATA_30",
	OpDATA31:              "OP_DATA_31",
	OpDATA32:              "OP_DATA_32",
	OpDATA33:              "OP_DATA_33",
	OpDATA34:              "OP_DATA_34",
	OpDATA35:              "OP_DATA_35",
	OpDATA36:              "OP_DATA_36",
	OpDATA37:              "OP_DATA_37",
	OpDATA38:              "OP_DATA_38",
	OpDATA39:              "OP_DATA_39",
	OpDATA40:              "OP_DATA_40",
	OpDATA41:              "OP_DATA_41",
	OpDATA42:              "OP_DATA_42",
	OpDATA43:              "OP_DATA_43",
	OpDATA44:              "OP_DATA_44",
	OpDATA45:              "OP_DATA_45",
	OpDATA46:              "OP_DATA_46",
	OpDATA47:              "OP_DATA_47",
	OpDATA48:              "OP_DATA_48",
	OpDATA49:              "OP_DATA_49",
	OpDATA50:              "OP_DATA_50",
	OpDATA51:              "OP_DATA_51",
	OpDATA52:              "OP_DATA_52",
	OpDATA53:              "OP_DATA_53",
	OpDATA54:              "OP_DATA_54",
	OpDATA55:              "OP_DATA_55",
	OpDATA56:              "OP_DATA_56",
	OpDATA57:              "OP_DATA_57",
	OpDATA58:              "OP_DATA_58",
	OpDATA59:              "OP_DATA_59",
	OpDATA60:              "OP_DATA_60",
	OpDATA61:              "OP_DATA_61",
	OpDATA62:              "OP_DATA_62",
	OpDATA63:              "OP_DATA_63",
	OpDATA64:              "OP_DATA_64",
	OpDATA65:              "OP_DATA_65",
	OpDATA66:              "OP_DATA_66",
	OpDATA67:              "OP_DATA_67",
	OpDATA68:              "OP_DATA_68",
	OpDATA69:              "OP_DATA_69",
	OpDATA70:              "OP_DATA_70",
	OpDATA71:              "OP_DATA_71",
	OpDATA72:              "OP_DATA_72",
	OpDATA73:              "OP_DATA_73",
	OpDATA74:              "OP_DATA_74",
	OpDATA75:              "OP_DATA_75",
	OpPUSHDATA1:           "OP_PUSHDATA1",
	OpPUSHDATA2:           "OP_PUSHDATA2",
	OpPUSHDATA4:           "OP_PUSHDATA4",
	Op1NEGATE:             "OP_1NEGATE",
	OpBASE:                "OP_BASE",
	OpTRUE:                "OP_TRUE",
	Op2:                   "OP_2",
	Op3:                   "OP_3",
	Op4:                   "OP_4",
	Op5:                   "OP_5",
	Op6:                   "OP_6",
	Op7:                   "OP_7",
	Op8:                   "OP_8",
	Op9:                   "OP_9",
	Op10:                  "OP_10",
	Op11:                  "OP_11",
	Op12:                  "OP_12",
	Op13:                  "OP_13",
	Op14:                  "OP_14",
	Op15:                  "OP_15",
	Op16:                  "OP_16",
	OpNOP:                 "OP_NOP",
	OpVER:                 "OP_VER",
	OpIF:                  "OP_IF",
	OpNOTIF:               "OP_NOTIF",
	OpVERIF:               "OP_VERIF",
	OpVERNOTIF:            "OP_VERNOTIF",
	OpELSE:                "OP_ELSE",
	OpENDIF:               "OP_ENDIF",
	OpVERIFY:              "OP_VERIFY",
	OpRETURN:              "OP_RETURN",
	OpTOALTSTACK:          "OP_TOALTSTACK",
	OpFROMALTSTACK:        "OP_FROMALTSTACK",
	Op2DROP:               "OP_2DROP",
	Op2DUP:                "OP_2DUP",
	Op3DUP:                "OP_3DUP",
	Op2OVER:               "OP_2OVER",
	Op2ROT:                "OP_2ROT",
	Op2SWAP:               "OP_2SWAP",
	OpIFDUP:               "OP_IFDUP",
	OpDEPTH:               "OP_DEPTH",
	OpDROP:                "OP_DROP",
	OpDUP:                 "OP_DUP",
	OpNIP:                 "OP_NIP",
	OpOVER:                "OP_OVER",
	OpPICK:                "OP_PICK",
	OpROLL:                "OP_ROLL",
	OpROT:                 "OP_ROT",
	OpSWAP:                "OP_SWAP",
	OpTUCK:                "OP_TUCK",
	OpCAT:                 "OP_CAT",
	OpSPLIT:               "OP_SPLIT",
	OpNUM2BIN:             "OP_NUM2BIN",
	OpBIN2NUM:             "OP_BIN2NUM",
	OpSIZE:                "OP_SIZE",
	OpINVERT:              "OP_INVERT",
	OpAND:                 "OP_AND",
	OpOR:                  "OP_OR",
	OpXOR:                 "OP_XOR",
	OpEQUAL:               "OP_EQUAL",
	OpEQUALVERIFY:         "OP_EQUALVERIFY",
	OpRESERVED1:           "OP_RESERVED1",
	OpRESERVED2:           "OP_RESERVED2",
	Op1ADD:                "OP_1ADD",
	Op1SUB:                "OP_1SUB",
	Op2MUL:                "OP_2MUL",
	Op2DIV:                "OP_2DIV",
	OpNEGATE:              "OP_NEGATE",
	OpABS:                 "OP_ABS",
	OpNOT:                 "OP_NOT",
	Op0NOTEQUAL:           "OP_0NOTEQUAL",
	OpADD:                 "OP_ADD",
	OpSUB:                 "OP_SUB",
	OpMUL:                 "OP_MUL",
	OpDIV:                 "OP_DIV",
	OpMOD:                 "OP_MOD",
	OpLSHIFT:              "OP_LSHIFT",
	OpRSHIFT:              "OP_RSHIFT",
	OpBOOLAND:             "OP_BOOLAND",
	OpBOOLOR:              "OP_BOOLOR",
	OpNUMEQUAL:            "OP_NUMEQUAL",
	OpNUMEQUALVERIFY:      "OP_NUMEQUALVERIFY",
	OpNUMNOTEQUAL:         "OP_NUMNOTEQUAL",
	OpLESSTHAN:            "OP_LESSTHAN",
	OpGREATERTHAN:         "OP_GREATERTHAN",
	OpLESSTHANOREQUAL:     "OP_LESSTHANOREQUAL",
	OpGREATERTHANOREQUAL:  "OP_GREATERTHANOREQUAL",
	OpMIN:                 "OP_MIN",
	OpMAX:                 "OP_MAX",
	OpWITHIN:              "OP_WITHIN",
	OpRIPEMD160:           "OP_RIPEMD160",
	OpSHA1:                "OP_SHA1",
	OpSHA256:              "OP_SHA256",
	OpHASH160:             "OP_HASH160",
	OpHASH256:             "OP_HASH256",
	OpCODESEPARATOR:       "OP_CODESEPARATOR",
	OpCHECKSIG:            "OP_CHECKSIG",
	OpCHECKSIGVERIFY:      "OP_CHECKSIGVERIFY",
	OpCHECKMULTISIG:       "OP_CHECKMULTISIG",
	OpCHECKMULTISIGVERIFY: "OP_CHECKMULTISIGVERIFY",
	OpNOP1:                "OP_NOP1",
	OpNOP2:                "OP_NOP2",
	OpNOP3:                "OP_NOP3",
	OpNOP4:                "OP_NOP4",
	OpNOP5:                "OP_NOP5",
	OpNOP6:                "OP_NOP6",
	OpNOP7:                "OP_NOP7",
	OpNOP8:                "OP_NOP8",
	OpNOP9:                "OP_NOP9",
	OpNOP10:               "OP_NOP10",
	OpUNKNOWN186:          "OP_UNKNOWN186",
	OpUNKNOWN187:          "OP_UNKNOWN187",
	OpUNKNOWN188:          "OP_UNKNOWN188",
	OpUNKNOWN189:          "OP_UNKNOWN189",
	OpUNKNOWN190:          "OP_UNKNOWN190",
	OpUNKNOWN191:          "OP_UNKNOWN191",
	OpUNKNOWN192:          "OP_UNKNOWN192",
	OpUNKNOWN193:          "OP_UNKNOWN193",
	OpUNKNOWN194:          "OP_UNKNOWN194",
	OpUNKNOWN195:          "OP_UNKNOWN195",
	OpUNKNOWN196:          "OP_UNKNOWN196",
	OpUNKNOWN197:          "OP_UNKNOWN197",
	OpUNKNOWN198:          "OP_UNKNOWN198",
	OpUNKNOWN199:          "OP_UNKNOWN199",
	OpUNKNOWN200:          "OP_UNKNOWN200",
	OpUNKNOWN201:          "OP_UNKNOWN201",
	OpUNKNOWN202:          "OP_UNKNOWN202",
	OpUNKNOWN203:          "OP_UNKNOWN203",
	OpUNKNOWN204:          "OP_UNKNOWN204",
	OpUNKNOWN205:          "OP_UNKNOWN205",
	OpUNKNOWN206:          "OP_UNKNOWN206",
	OpUNKNOWN207:          "OP_UNKNOWN207",
	OpUNKNOWN208:          "OP_UNKNOWN208",
	OpUNKNOWN209:          "OP_UNKNOWN209",
	OpUNKNOWN210:          "OP_UNKNOWN210",
	OpUNKNOWN211:          "OP_UNKNOWN211",
	OpUNKNOWN212:          "OP_UNKNOWN212",
	OpUNKNOWN213:          "OP_UNKNOWN213",
	OpUNKNOWN214:          "OP_UNKNOWN214",
	OpUNKNOWN215:          "OP_UNKNOWN215",
	OpUNKNOWN216:          "OP_UNKNOWN216",
	OpUNKNOWN217:          "OP_UNKNOWN217",
	OpUNKNOWN218:          "OP_UNKNOWN218",
	OpUNKNOWN219:          "OP_UNKNOWN219",
	OpUNKNOWN220:          "OP_UNKNOWN220",
	OpUNKNOWN221:          "OP_UNKNOWN221",
	OpUNKNOWN222:          "OP_UNKNOWN222",
	OpUNKNOWN223:          "OP_UNKNOWN223",
	OpUNKNOWN224:          "OP_UNKNOWN224",
	OpUNKNOWN225:          "OP_UNKNOWN225",
	OpUNKNOWN226:          "OP_UNKNOWN226",
	OpUNKNOWN227:          "OP_UNKNOWN227",
	OpUNKNOWN228:          "OP_UNKNOWN228",
	OpUNKNOWN229:          "OP_UNKNOWN229",
	OpUNKNOWN230:          "OP_UNKNOWN230",
	OpUNKNOWN231:          "OP_UNKNOWN231",
	OpUNKNOWN232:          "OP_UNKNOWN232",
	OpUNKNOWN233:          "OP_UNKNOWN233",
	OpUNKNOWN234:          "OP_UNKNOWN234",
	OpUNKNOWN235:          "OP_UNKNOWN235",
	OpUNKNOWN236:          "OP_UNKNOWN236",
	OpUNKNOWN237:          "OP_UNKNOWN237",
	OpUNKNOWN238:          "OP_UNKNOWN238",
	OpUNKNOWN239:          "OP_UNKNOWN239",
	OpUNKNOWN240:          "OP_UNKNOWN240",
	OpUNKNOWN241:          "OP_UNKNOWN241",
	OpUNKNOWN242:          "OP_UNKNOWN242",
	OpUNKNOWN243:          "OP_UNKNOWN243",
	OpUNKNOWN244:          "OP_UNKNOWN244",
	OpUNKNOWN245:          "OP_UNKNOWN245",
	OpUNKNOWN246:          "OP_UNKNOWN246",
	OpUNKNOWN247:          "OP_UNKNOWN247",
	OpUNKNOWN248:          "OP_UNKNOWN248",
	OpUNKNOWN249:          "OP_UNKNOWN249",
	OpSMALLINTEGER:        "OP_SMALLINTEGER",
	OpPUBKEYS:             "OP_PUBKEYS",
	OpUNKNOWN252:          "OP_UNKNOWN252",
	OpPUBKEYHASH:          "OP_PUBKEYHASH",
	OpPUBKEY:              "OP_PUBKEY",
	OpINVALIDOPCODE:       "OP_INVALIDOPCODE",
}
