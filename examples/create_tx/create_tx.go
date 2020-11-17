package main

import (
	"log"

	"github.com/bitcoinsv/bsvutil"
	"github.com/libsv/go-bt"
)

func main() {
	tx := bt.NewTx()

	_ = tx.From(
		"11b476ad8e0a48fcd40807a111a050af51114877e09283bfa7f3505081a1819d",
		0,
		"76a914eb0bd5edba389198e73f8efabddfc61666969ff788ac6a0568656c6c6f",
		1500)

	_ = tx.PayTo("1NRoySJ9Lvby6DuE2UQYnyT67AASwNZxGb", 1000)

	wif, _ := bsvutil.DecodeWIF("KznvCNc6Yf4iztSThoMH6oHWzH9EgjfodKxmeuUGPq5DEX5maspS")

	err := tx.SignAuto(&bt.InternalSigner{PrivateKey: wif.PrivKey, SigHashFlag: 0})
	if err != nil {
		log.Fatalln(err.Error())
	}
	log.Println("tx: ", tx.ToString())
}
