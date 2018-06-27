package cryptolib

// ReverseHexString comment
func ReverseHexString(hex string) string {
	res := ""
	if len(hex)%2 != 0 {
		hex = "0" + hex
	}

	for i := len(hex); i >= 2; i -= 2 {

		res += hex[i-2 : i]
	}
	return res
}
