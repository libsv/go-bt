package cryptolib

// ReverseBytes reverses the bytes (little endian/big endian).
// This is used when computing merkle trees in Bitcoin, for example.
func ReverseBytes(a []byte) []byte {
	tmp := make([]byte, len(a))
	copy(tmp, a)

	for i, j := 0, len(a)-1; i < j; i, j = i+1, j-1 {
		tmp[i], tmp[j] = tmp[j], tmp[i]
	}
	return tmp
}
