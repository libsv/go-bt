package main

import (
	"log"

	"github.com/libsv/go-bt"
	"github.com/libsv/go-bt/bsvutil"
)

func main() {
	tx := bt.NewTx()

	_ = tx.From(
		"b7b0650a7c3a1bd4716369783876348b59f5404784970192cec1996e86950576",
		0,
		"76a9149cbe9f5e72fa286ac8a38052d1d5337aa363ea7f88ac",
		1000)

	_ = tx.PayTo("1C8bzHM8XFBHZ2ZZVvFy2NSoAZbwCXAicL", 900)

	_ = tx.AddOpReturnOutput([]byte("You are using go-bt!"))

	wif, _ := bsvutil.DecodeWIF("L3VJH2hcRGYYG6YrbWGmsxQC1zyYixA82YjgEyrEUWDs4ALgk8Vu")

	inputsSigned, err := tx.SignAuto(&bt.LocalSigner{PrivateKey: wif.PrivKey})
	if err != nil && len(inputsSigned) > 0 {
		log.Fatal(err.Error())
	}
	log.Println("tx: ", tx.ToString())
}
