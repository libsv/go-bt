package main

import (
	"encoding/json"
	"fmt"

	"github.com/libsv/go-bt/v2"
)

func main() {
	j := `[
	{
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
	},
	{
		"version": 2,
		"locktime": 119,
		"txid": "35d2d1db9bb3d1398faaa5addfc6aeaa6b3f1357d00098660b1554d4466d99b2",
		"hash": "35d2d1db9bb3d1398faaa5addfc6aeaa6b3f1357d00098660b1554d4466d99b2",
		"size": 191,
		"hex": "020000000117d2011c2a3b8a309d481930bae86e88017b0f55845ada17f96c464684b3af520000000048473044022014a60c3e84cf0160cb7e4ee7d87a3b78c5efb6dd3b66c76970b680affdb95e8f02207f6d9e3268a934e5e278ae513a3bc6dee3bec7bae37204574480305bfb5dea0e41feffffff0240101024010000001976a9149933e4bad50e7dd4b48c1f0be98436ca7d4392a288ac00e1f505000000001976a914abbe187ad301e4326e59587e43d602edd318364e88ac77000000",
		"vin": [
			{
				"scriptSig": {
					"asm": "3044022014a60c3e84cf0160cb7e4ee7d87a3b78c5efb6dd3b66c76970b680affdb95e8f02207f6d9e3268a934e5e278ae513a3bc6dee3bec7bae37204574480305bfb5dea0e41",
					"hex": "473044022014a60c3e84cf0160cb7e4ee7d87a3b78c5efb6dd3b66c76970b680affdb95e8f02207f6d9e3268a934e5e278ae513a3bc6dee3bec7bae37204574480305bfb5dea0e41"
				},
				"txid": "52afb38446466cf917da5a84550f7b01886ee8ba3019489d308a3b2a1c01d217",
				"vout": 0,
				"sequence": 4294967294
			}
		],
		"vout": [
			{
				"value": 48.99999808,
				"n": 0,
				"scriptPubKey": {
					"asm": "OP_DUP OP_HASH160 9933e4bad50e7dd4b48c1f0be98436ca7d4392a2 OP_EQUALVERIFY OP_CHECKSIG",
					"hex": "76a9149933e4bad50e7dd4b48c1f0be98436ca7d4392a288ac",
					"reqSigs": 1,
					"type": "pubkeyhash"
				}
			},
			{
				"value": 1,
				"n": 1,
				"scriptPubKey": {
					"asm": "OP_DUP OP_HASH160 abbe187ad301e4326e59587e43d602edd318364e OP_EQUALVERIFY OP_CHECKSIG",
					"hex": "76a914abbe187ad301e4326e59587e43d602edd318364e88ac",
					"reqSigs": 1,
					"type": "pubkeyhash"
				}
			}
		]
	}
]`

	var txs bt.Txs
	if err := json.Unmarshal([]byte(j), txs.NodeJSON()); err != nil {
		panic(err)
	}

	for _, tx := range txs {
		fmt.Println(tx)
	}
}
