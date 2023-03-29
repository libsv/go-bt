package bt

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"mime"
	"testing"

	"github.com/libsv/go-bk/wif"
	"github.com/libsv/go-bt/v2/bscript"
	"github.com/stretchr/testify/assert"
)

func TestInscribe(t *testing.T) {
	t.Parallel()

	s, _ := bscript.NewP2PKHFromAddress("mxAoAyZFXX6LZBWhoam3vjm6xt9NxPQ15f")

	tests := map[string]struct {
		ia *bscript.InscriptionArgs
	}{
		"text/plain with 'Hello, world!'": {
			ia: &bscript.InscriptionArgs{
				LockingScriptPrefix: s,
				Data:                []byte("Hello, world!"),
				ContentType:         "text/plain;charset=utf-8",
			},
		}, "model/gltf-binary 3d model'": {
			ia: &bscript.InscriptionArgs{
				LockingScriptPrefix: s,
				Data: func() []byte {
					b, _ := base64.StdEncoding.DecodeString("Z2xURgIAAABobgwAbAcAAEpTT057ImFzc2V0Ijp7ImdlbmVyYXRvciI6Ik1pY3Jvc29mdCBHTFRGIEV4cG9ydGVyIDIuOC4zLjQwIiwidmVyc2lvbiI6IjIuMCJ9LCJhY2Nlc3NvcnMiOlt7ImJ1ZmZlclZpZXciOjAsImNvbXBvbmVudFR5cGUiOjUxMjUsImNvdW50IjozNDgsInR5cGUiOiJTQ0FMQVIifSx7ImJ1ZmZlclZpZXciOjEsImNvbXBvbmVudFR5cGUiOjUxMjYsImNvdW50IjozNDgsInR5cGUiOiJWRUMzIiwibWF4IjpbMC4zNjIyNjQ5OTA4MDY1Nzk2LDEuMTU1ODU4OTkzNTMwMjczNSwwLjQwOTIwNjAwMjk1MDY2ODM2XSwibWluIjpbLTAuNDUyMjI4OTkzMTc3NDEzOTYsLTAuNjI1MDc5OTg5NDMzMjg4NiwtMC40MDY1NDk5OTAxNzcxNTQ1Nl19LHsiYnVmZmVyVmlldyI6MiwiY29tcG9uZW50VHlwZSI6NTEyNiwiY291bnQiOjM0OCwidHlwZSI6IlZFQzMifSx7ImJ1ZmZlclZpZXciOjMsImNvbXBvbmVudFR5cGUiOjUxMjYsImNvdW50IjozNDgsInR5cGUiOiJWRUM0In0seyJidWZmZXJWaWV3Ijo0LCJjb21wb25lbnRUeXBlIjo1MTI2LCJjb3VudCI6MzQ4LCJ0eXBlIjoiVkVDMiJ9XSwiYnVmZmVyVmlld3MiOlt7ImJ1ZmZlciI6MCwiYnl0ZU9mZnNldCI6MCwiYnl0ZUxlbmd0aCI6MTM5MiwidGFyZ2V0IjozNDk2M30seyJidWZmZXIiOjAsImJ5dGVPZmZzZXQiOjEzOTIsImJ5dGVMZW5ndGgiOjQxNzYsInRhcmdldCI6MzQ5NjJ9LHsiYnVmZmVyIjowLCJieXRlT2Zmc2V0Ijo1NTY4LCJieXRlTGVuZ3RoIjo0MTc2LCJ0YXJnZXQiOjM0OTYyfSx7ImJ1ZmZlciI6MCwiYnl0ZU9mZnNldCI6OTc0NCwiYnl0ZUxlbmd0aCI6NTU2OCwidGFyZ2V0IjozNDk2Mn0seyJidWZmZXIiOjAsImJ5dGVPZmZzZXQiOjE1MzEyLCJieXRlTGVuZ3RoIjoyNzg0LCJ0YXJnZXQiOjM0OTYyfSx7ImJ1ZmZlciI6MCwiYnl0ZU9mZnNldCI6MTgwOTYsImJ5dGVMZW5ndGgiOjc5NDY2OX1dLCJidWZmZXJzIjpbeyJieXRlTGVuZ3RoIjo4MTI3NjV9XSwiaW1hZ2VzIjpbeyJidWZmZXJWaWV3Ijo1LCJtaW1lVHlwZSI6ImltYWdlL3BuZyJ9XSwibWF0ZXJpYWxzIjpbeyJwYnJNZXRhbGxpY1JvdWdobmVzcyI6eyJiYXNlQ29sb3JUZXh0dXJlIjp7ImluZGV4IjowfSwibWV0YWxsaWNGYWN0b3IiOjAuMCwicm91Z2huZXNzRmFjdG9yIjowLjEwOTgwMzkyOTkyNDk2NDl9LCJkb3VibGVTaWRlZCI6dHJ1ZX1dLCJtZXNoZXMiOlt7InByaW1pdGl2ZXMiOlt7ImF0dHJpYnV0ZXMiOnsiVEFOR0VOVCI6MywiTk9STUFMIjoyLCJQT1NJVElPTiI6MSwiVEVYQ09PUkRfMCI6NH0sImluZGljZXMiOjAsIm1hdGVyaWFsIjowfV19XSwibm9kZXMiOlt7ImNoaWxkcmVuIjpbMV0sInJvdGF0aW9uIjpbLTAuNzA3MTA2NzA5NDgwMjg1NiwwLjAsLTAuMCwwLjcwNzEwNjgyODY4OTU3NTJdLCJzY2FsZSI6WzEuMCwwLjk5OTk5OTk0MDM5NTM1NTIsMC45OTk5OTk5NDAzOTUzNTUyXSwibmFtZSI6IlJvb3ROb2RlIChnbHRmIG9yaWVudGF0aW9uIG1hdHJpeCkifSx7ImNoaWxkcmVuIjpbMl0sIm5hbWUiOiJSb290Tm9kZSAobW9kZWwgY29ycmVjdGlvbiBtYXRyaXgpIn0seyJjaGlsZHJlbiI6WzNdLCJyb3RhdGlvbiI6WzAuNzA3MTA2NzY5MDg0OTMwNCwwLjAsMC4wLDAuNzA3MTA2NzY5MDg0OTMwNF0sIm5hbWUiOiI5MzcwMjFkNDkyYjM0ZjBlOWM4NDU3YjBmMTNhYTBmZS5mYngifSx7ImNoaWxkcmVuIjpbNF0sIm5hbWUiOiJSb290Tm9kZSJ9LHsiY2hpbGRyZW4iOls1XSwibmFtZSI6ImNyeXN0YWxMb3c6TWVzaCJ9LHsibWVzaCI6MCwibmFtZSI6ImNyeXN0YWxMb3c6TWVzaF9sYW1iZXJ0NF8wIn1dLCJzYW1wbGVycyI6W3sibWFnRmlsdGVyIjo5NzI5LCJtaW5GaWx0ZXIiOjk5ODd9XSwic2NlbmVzIjpbeyJub2RlcyI6WzBdfV0sInRleHR1cmVzIjpbeyJzYW1wbGVyIjowLCJzb3VyY2UiOjB9XSwic2NlbmUiOjB94GYMAEJJTgAAAAAAAQAAAAIAAAADAAAABAAAAAUAAAAGAAAABwAAAAgAAAAJAAAACgAAAAsAAAAMAAAADQAAAA4AAAAPAAAAEAAAABEAAAASAAAAEwAAABQAAAAVAAAAFgAAABcAAAAYAAAAGQAAABoAAAAbAAAAHAAAAB0AAAAeAAAAHwAAACAAAAAhAAAAIgAAACMAAAAkAAAAJQAAACYAAAAnAAAAKAAAACkAAAAqAAAAKwAAACwAAAAtAAAALgAAAC8AAAAwAAAAMQAAADIAAAAzAAAANAAAADUAAAA2AAAANwAAADgAAAA5AAAAOgAAADsAAAA8AAAAPQAAAD4AAAA")
					return b
				}(),
				ContentType: "model/gltf-binary",
			},
		}}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {

			// Create a new transaction
			tx := NewTx()

			// Call the Inscribe function with the test data
			err := tx.Inscribe(test.ia)
			assert.NoError(t, err)

			// Check if the transaction has the expected number of outputs
			if len(tx.Outputs) != 1 {
				t.Fatalf("Inscribe failed: expected 1 output, got %d", len(tx.Outputs))
			}

			// Check if the output has the expected Satoshis value
			if tx.Outputs[0].Satoshis != 1 {
				t.Fatalf("Inscribe failed: expected Satoshis value 1, got %d", tx.Outputs[0].Satoshis)
			}

			// Check if the output has the expected LockingScript value
			expectedLockingScript := *s
			_ = expectedLockingScript.AppendOpcodes(bscript.OpFALSE, bscript.OpIF)
			_ = expectedLockingScript.AppendPushDataString(OrdinalsPrefix)
			_ = expectedLockingScript.AppendOpcodes(bscript.Op1)
			_ = expectedLockingScript.AppendPushData([]byte(test.ia.ContentType))
			_ = expectedLockingScript.AppendOpcodes(bscript.Op0)
			_ = expectedLockingScript.AppendPushData(test.ia.Data)
			_ = expectedLockingScript.AppendOpcodes(bscript.OpENDIF)

			if !tx.Outputs[0].LockingScript.Equals(&expectedLockingScript) {
				t.Fatalf("Inscribe failed: expected LockingScript value %s, got %s", expectedLockingScript.String(), tx.Outputs[0].LockingScript.String())
			}
		})
	}

}

func TestMultipleInscriptionsIn1Tx(t *testing.T) {
	decodedWif, _ := wif.DecodeWIF("KznpA63DPFrmHecASyL6sFmcRgrNT9oM8Ebso8mwq1dfJF3ZgZ3V")

	// get public key bytes and address
	pubkey := decodedWif.SerialisePubKey()
	addr, _ := bscript.NewAddressFromPublicKeyString(hex.EncodeToString(pubkey), true)
	s, _ := bscript.NewP2PKHFromAddress(addr.AddressString)

	tx := NewTx()

	_ = tx.From(
		"39e5954ee335fdb5a1368ab9e851a954ed513f73f6e8e85eff5e31adbb5837e7",
		0,
		"76a9144bca0c466925b875875a8e1355698bdcc0b2d45d88ac",
		500,
	)

	tx.Inscribe(&bscript.InscriptionArgs{
		LockingScriptPrefix: s,
		Data:                []byte("Hello, world!"),
		ContentType:         "text/plain;charset=utf-8",
	})
	tx.Inscribe(&bscript.InscriptionArgs{
		LockingScriptPrefix: s,
		Data:                []byte("Hello, world!"),
		ContentType:         "text/plain;charset=utf-8",
	})

	tx.Inscribe(&bscript.InscriptionArgs{
		LockingScriptPrefix: s,
		Data:                []byte("Hello, world!"),
		ContentType:         "text/plain;charset=utf-8",
	})

	// Check if the output has the expected LockingScript value
	expectedLockingScript := s
	_ = expectedLockingScript.AppendOpcodes(bscript.OpFALSE, bscript.OpIF)
	_ = expectedLockingScript.AppendPushDataString(OrdinalsPrefix)
	_ = expectedLockingScript.AppendOpcodes(bscript.Op1)
	_ = expectedLockingScript.AppendPushData([]byte("text/plain;charset=utf-8"))
	_ = expectedLockingScript.AppendOpcodes(bscript.Op0)
	_ = expectedLockingScript.AppendPushData([]byte("Hello, world!"))
	_ = expectedLockingScript.AppendOpcodes(bscript.OpENDIF)

	for i := range tx.Outputs {
		if !tx.Outputs[i].LockingScript.Equals(expectedLockingScript) {
			t.Fatalf("Inscribe failed: expected LockingScript value %s, got %s at index %d", expectedLockingScript.String(), tx.Outputs[i].LockingScript.String(), i)
		}
	}

	log.Println("tx: ", tx.String())
}

func TestInscribeFromFile(t *testing.T) {
	decodedWif, _ := wif.DecodeWIF("KznpA63DPFrmHecASyL6sFmcRgrNT9oM8Ebso8mwq1dfJF3ZgZ3V")

	// get public key bytes and address
	pubkey := decodedWif.SerialisePubKey()
	addr, _ := bscript.NewAddressFromPublicKeyString(hex.EncodeToString(pubkey), true)
	s, _ := bscript.NewP2PKHFromAddress(addr.AddressString)

	tx := NewTx()

	_ = tx.From(
		"39e5954ee335fdb5a1368ab9e851a954ed513f73f6e8e85eff5e31adbb5837e7",
		0,
		"76a9144bca0c466925b875875a8e1355698bdcc0b2d45d88ac",
		500,
	)

	// Read the image file
	data, err := ioutil.ReadFile("./examples/create_tx_with_inscription/1SatLogoLight.png")
	if err != nil {
		fmt.Println(err)
		return
	}

	// Get the content type of the image
	contentType := mime.TypeByExtension(".png")

	tx.Inscribe(&bscript.InscriptionArgs{
		LockingScriptPrefix: s,
		Data:                data,
		ContentType:         contentType,
	})

	err = tx.ChangeToAddress("17ujiveRLkf2JQiGR8Sjtwb37evX7vG3WG", NewFeeQuote())
	if err != nil {
		log.Fatal(err.Error())
	}
}
