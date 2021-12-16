package main

import (
	"bufio"
	"fmt"
	"io"

	"github.com/libsv/go-bt/v2"
	"github.com/libsv/go-bt/v2/testing/data"
)

// In this example, all txs from a block are being read in via chunking, so at no point
// does the entire block have to be held in memory, and instead can be streamed.
//
// We represent the block by interatively reading a file, however it could be any data
// stream that satisfies the io.Reader interface.

func main() {
	// Open file container block data.
	f, err := data.TxBinData.Open("block.bin")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	// Create buffered reader for this file.
	r := bufio.NewReader(f)

	// Read file header. This step is specific to file reading and
	// may need omitted or modified for other implentations.
	_, err = io.ReadFull(f, make([]byte, 80))
	if err != nil {
		panic(err)
	}

	txs := bt.Txs{}
	if _, err = txs.ReadFrom(r); err != nil {
		panic(err)
	}
	for _, tx := range txs {
		fmt.Println(tx.TxID())
	}
}
