package cryptolib

// Bitcoin Script constants
const (
	OpZERO          = 0x00
	OpONE           = 0x51
	OpSIXTEEN       = 0x60
	OpBASE          = 0x50
	OpDUP           = 0x76 // Duplicate the top item in the stack
	OpHASH160       = 0xa9 // Return RIPEMD160(SHA256(x)) hash of top item
	OpEQUAL         = 0x87 //	Returns 1 if the inputs are exactly equal, 0 otherwise.
	OpEQUALVERIFY   = 0x88 // Same as OP_EQUAL, but run OP_VERIFY after to halt if not TRUE
	OpCHECKSIG      = 0xac // Pop a public key and signature and validate the signature for the transaction's hashed data, return TRUE if matching
	OpCHECKMULTISIG = 0xae
	OpRETURN        = 0x6a
	OpPUSHDATA1     = 0x4c
	OpPUSHDATA2     = 0x4d
	OpPUSHDATA4     = 0x4e
	OpFALSE         = 0x00
	OpDROP          = 0x75
)
