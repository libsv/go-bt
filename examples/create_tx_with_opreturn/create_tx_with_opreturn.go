package main

import (
	"context"
	"log"

	"github.com/libsv/go-bk/wif"
	"github.com/libsv/go-bt/v2"
)

func main() {
	tx := bt.NewTx()

	_ = tx.From(
		"b7b0650a7c3a1bd4716369783876348b59f5404784970192cec1996e86950576",
		0,
		"76a9149cbe9f5e72fa286ac8a38052d1d5337aa363ea7f88ac",
		1000,
	)

	_ = tx.PayToAddress("1C8bzHM8XFBHZ2ZZVvFy2NSoAZbwCXAicL", 900)

	_ = tx.AddOpReturnOutput([]byte("You are using go-bt!"))

	decodedWif, _ := wif.DecodeWIF("L3VJH2hcRGYYG6YrbWGmsxQC1zyYixA82YjgEyrEUWDs4ALgk8Vu")

	err := tx.SignAll(context.Background(), &bt.LocalSignerGetter{PrivateKey: decodedWif.PrivKey})
	if err != nil {
		log.Fatal(err.Error())
	}
	log.Println("tx: ", tx.String())
}
