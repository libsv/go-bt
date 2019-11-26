package cryptolib

// PrivateKeyToWIF takes a 256 bit private key and outputs it in Wallet Interchange Format
func PrivateKeyToWIF(key []byte) string {
	// Add a 0x80 byte in front of it for mainnet addresses or 0xef for testnet addresses. Also add a 0x01 byte at the end if the private key will correspond to a compressed public key
	b := []byte{0x80}
	key = append(b, key...)

	// Perform double SHA-256 hash on the extended key...
	checksum := Sha256d(key)

	// Take the first 4 bytes of the double SHA-256 hash, this is the checksum.  Append it to the extended key.
	key = append(key, checksum[0:4]...)

	// Convert the result from a byte string into a base58 string using Base58Check encoding. This is the Wallet Import Format
	base58 := EncodeToString(key)
	// fmt.Printf("WIF: %v\n", base58)
	return base58
}
