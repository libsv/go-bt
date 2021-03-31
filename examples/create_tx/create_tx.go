package main

import (
	"log"

	"github.com/libsv/go-bt"
	"github.com/libsv/go-bt/bsvutil"
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

	inputsSigned, err := tx.SignAuto(&bt.LocalSigner{PrivateKey: wif.PrivKey})
	if err != nil && len(inputsSigned) > 0 {
		log.Fatal(err.Error())
	}
	log.Println("tx: ", tx.ToString())
}
