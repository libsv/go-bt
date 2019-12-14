package cryptolib

// PublicKeyToP2PKHScript comment
func PublicKeyToP2PKHScript(pubkey []byte) []byte {
	hash := Hash160(pubkey)

	ret := []byte{
		OpDUP,
		OpHASH160,
		0x14,
	}
	ret = append(ret, hash...)
	ret = append(ret, OpEQUALVERIFY)
	ret = append(ret, OpCHECKSIG)

	return ret
}
