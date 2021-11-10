package main

import (
	"encoding/json"
	"fmt"

	"github.com/libsv/go-bt/v2"
)

func main() {
	tx, err := bt.NewTxFromString("0100000001abad53d72f342dd3f338e5e3346b492440f8ea821f8b8800e318f461cc5ea5a2010000006a4730440220042edc1302c5463e8397120a56b28ea381c8f7f6d9bdc1fee5ebca00c84a76e2022077069bbdb7ed701c4977b7db0aba80d41d4e693112256660bb5d674599e390cf41210294639d6e4249ea381c2e077e95c78fc97afe47a52eb24e1b1595cd3fdd0afdf8ffffffff02000000000000000008006a0548656c6c6f7f030000000000001976a914b85524abf8202a961b847a3bd0bc89d3d4d41cc588ac00000000")
	if err != nil {
		panic(err)
	}

	bb, err := json.MarshalIndent(tx.NodeJSON(), "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(bb))
}
