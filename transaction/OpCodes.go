package transaction

const (
	opZERO          = 0x00
	opONE           = 0x51
	opSIXTEEN       = 0x60
	opBASE          = 0x50
	opDUP           = 0x76 // Duplicate the top item in the stack
	opHASH160       = 0xa9 // Return RIPEMD160(SHA256(x)) hash of top item
	opEQUAL         = 0x87 //	Returns 1 if the inputs are exactly equal, 0 otherwise.
	opEQUALVERIFY   = 0x88 // Same as OP_EQUAL, but run OP_VERIFY after to halt if not TRUE
	opCHECKSIG      = 0xac // Pop a public key and signature and validate the signature for the transaction's hashed data, return TRUE if matching
	opCHECKMULTISIG = 0xae
)
