package main

import (
	"log"

	"github.com/bitcoinsv/bsvutil"
	"github.com/libsv/go-bt"
)

func main() {
	tx := bt.NewTx()

	_ = tx.From(
		"b7b0650a7c3a1bd4716369783876348b59f5404784970192cec1996e86950576",
		0,
		"76a9149cbe9f5e72fa286ac8a38052d1d5337aa363ea7f88ac",
		1000)

	_ = tx.PayTo("1C8bzHM8XFBHZ2ZZVvFy2NSoAZbwCXAicL", 900)

	o, err := bt.NewOpReturnOutput([]byte("You are using LiBSV!"))
	if err != nil {
		log.Fatal(err.Error())
	}

	tx.AddOutput(o)

	wif, _ := bsvutil.DecodeWIF("L3VJH2hcRGYYG6YrbWGmsxQC1zyYixA82YjgEyrEUWDs4ALgk8Vu")

	err = tx.SignAuto(&bt.InternalSigner{PrivateKey: wif.PrivKey, SigHashFlag: 0})
	if err != nil {
		log.Fatal(err.Error())
	}
	log.Println("tx: ", tx.ToString())
}
