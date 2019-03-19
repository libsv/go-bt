package cryptolib

import (
	"fmt"
	"testing"

	"github.com/btcsuite/btcutil/base58"
)

func TestVerifySignature(t *testing.T) {

	// message := []byte("Hello world")
	address := "1DnmCuL9xMnWBy9rUwWUyz1vi57LMM2AfJ"
	decoded, netID, err := base58.CheckDecode(address)

	fmt.Println(decoded, netID, err)

}
