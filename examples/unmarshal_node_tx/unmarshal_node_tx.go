package main

import (
	"encoding/json"
	"fmt"

	"github.com/libsv/go-bt/v2"
)

func main() {
	j := `{
	"version": 1,
	"locktime": 0,
	"txid": "aec245f27b7640c8b1865045107731bfb848115c573f7da38166074b1c9e475d",
	"hash": "aec245f27b7640c8b1865045107731bfb848115c573f7da38166074b1c9e475d",
	"size": 208,
	"hex": "0100000001abad53d72f342dd3f338e5e3346b492440f8ea821f8b8800e318f461cc5ea5a2010000006a4730440220042edc1302c5463e8397120a56b28ea381c8f7f6d9bdc1fee5ebca00c84a76e2022077069bbdb7ed701c4977b7db0aba80d41d4e693112256660bb5d674599e390cf41210294639d6e4249ea381c2e077e95c78fc97afe47a52eb24e1b1595cd3fdd0afdf8ffffffff02000000000000000008006a0548656c6c6f7f030000000000001976a914b85524abf8202a961b847a3bd0bc89d3d4d41cc588ac00000000",
	"vin": [
		{
			"scriptSig": {
				"asm": "30440220042edc1302c5463e8397120a56b28ea381c8f7f6d9bdc1fee5ebca00c84a76e2022077069bbdb7ed701c4977b7db0aba80d41d4e693112256660bb5d674599e390cf41 0294639d6e4249ea381c2e077e95c78fc97afe47a52eb24e1b1595cd3fdd0afdf8",
				"hex": "4730440220042edc1302c5463e8397120a56b28ea381c8f7f6d9bdc1fee5ebca00c84a76e2022077069bbdb7ed701c4977b7db0aba80d41d4e693112256660bb5d674599e390cf41210294639d6e4249ea381c2e077e95c78fc97afe47a52eb24e1b1595cd3fdd0afdf8"
			},
			"txid": "a2a55ecc61f418e300888b1f82eaf84024496b34e3e538f3d32d342fd753adab",
			"vout": 1,
			"sequence": 4294967295
		}
	],
	"vout": [
		{
			"value": 0,
			"n": 0,
			"scriptPubKey": {
				"asm": "OP_FALSE OP_RETURN 48656c6c6f",
				"hex": "006a0548656c6c6f",
				"type": "nulldata"
			}
		},
		{
			"value": 0.00000895,
			"n": 1,
			"scriptPubKey": {
				"asm": "OP_DUP OP_HASH160 b85524abf8202a961b847a3bd0bc89d3d4d41cc5 OP_EQUALVERIFY OP_CHECKSIG",
				"hex": "76a914b85524abf8202a961b847a3bd0bc89d3d4d41cc588ac",
				"reqSigs": 1,
				"type": "pubkeyhash"
			}
		}
	]
}`

	tx := bt.NewTx()
	if err := json.Unmarshal([]byte(j), tx.NodeJSON()); err != nil {
		panic(err)
	}

	fmt.Println(tx.String())
}
