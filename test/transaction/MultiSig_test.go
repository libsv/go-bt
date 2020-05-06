package transaction

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/libsv/libsv/script"

	"github.com/bitcoinsv/bsvd/chaincfg/chainhash"
	"github.com/libsv/libsv/bsvsuite/bsvec"
)

// To disable log output during tests (see https://golangcode.com/disable-log-output-during-tests/)
func TestMain(m *testing.M) {
	log.SetOutput(ioutil.Discard)
	os.Exit(m.Run())
}

func TestMultiSig(t *testing.T) {

	// Create 2of2 redeem script
	rs, err := script.NewRedeemScript(2)
	if err != nil {
		t.Error(err)
	}

	rs.AddPublicKey("xpub661MyMwAqRbcF5ivRisXcZTEoy7d9DfLF6fLqpu5GWMfeUyGHuWJHVp5uexDqXTWoySh8pNx3ELW7qymwPNg3UEYHjwh1tpdm3P9J2j4g32", []uint32{0, 0})
	rs.AddPublicKey("xpub661MyMwAqRbcFvmkwJ82wpjkNmjMWb8n4Pp9Gz3dJjPZMh4uW7z9CpSsTNjEcH3KW5Tibn77qDM9X7gJyjxySEgzBmoQ9LGxSrgHMXTMqx6", []uint32{0, 0})

	var redeemScript = rs.GetRedeemScript()

	const xprv = "xprv9s21ZrQH143K2beTKhLXFRWWFwH8jkwUssjk3SVTiApgmge7kNC3jhVc4NgHW8PhW2y7BCDErqnKpKuyQMjqSePPJooPJowAz5BVLThsv6c"
	const privHex = "5f86e4023a4e94f00463f81b70ff951f83f896a0a3e6ed89cf163c152f954f8b"

	pkBytes, err := hex.DecodeString(privHex)
	if err != nil {
		fmt.Println(err)
		return
	}
	privKey, pubKey := bsvec.PrivKeyFromBytes(bsvec.S256(), pkBytes)

	// Sign a message using the private key.
	messageHash := chainhash.DoubleHashB(redeemScript)
	signature, err := privKey.Sign(messageHash)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Serialize and display the signature.
	t.Logf("Serialized Signature: %x\n", signature.Serialize())

	// Verify the signature for the message using the public key.
	verified := signature.Verify(messageHash, pubKey)
	t.Logf("Signature Verified? %v\n", verified)

}
