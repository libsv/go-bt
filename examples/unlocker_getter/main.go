package main

import (
	"context"

	"github.com/libsv/go-bk/wif"
	"github.com/libsv/go-bt/v2"
	"github.com/libsv/go-bt/v2/bscript"
	"github.com/libsv/go-bt/v2/unlocker"
)

// This example gives a simple in-memory based example of how to implement and use a `bt.UnlockerGetter`
// using derivated public/private keys.
//
// The basic idea is, we have accounts, each with a master private key. If someone would like to send money
// to an account, they request "destinations" from the account. These destinations are added to the
// tx, which is then ultimately broadcast.
//
// A destination in this example is simply a P2PKH locking script, however, the PK Hash will be unique
// on each call as under the hood, the account is deriving a new private/public key pair from its master
// key, creating a P2PKH script from that this pair, and storing the value used to derive this private/public
// key pair against the P2PKH script that it produced.
//
// When an account wishes to spend a fund it received in this manner, after adding the funds to the tx it is
// currently building, it calls `tx.FillAllInputs`. This function iterates all inputs on the tx, passing
// their `PreviousTxScript` to the `bt.UnlockerGetter` provided to the `bt.FillAllInputs` call.
// This allows an account when receiving a locking script to refer to its own script=>derivation mapping,
// and ultimately derive the private key used to create the public key that used to create the locking script.
// Finally allowing for an `unlocker.Simple` to be returned, with this derived private key.
func main() {
	// Create two accounts. The first is our account, which we will pretend to fund to begin with.
	// The second is the merchant, which we will pretend to send money to.
	myAccount := newAccount()
	merchantAccount := newAccount()

	// Base Tx to "fund" our account.
	baseTx := bt.NewTx()
	if err := baseTx.From(
		"11b476ad8e0a48fcd40807a111a050af51114877e09283bfa7f3505081a1819d",
		0,
		"76a914b48b288c48e6cd7246876e19f848a60f46ab4a6188ac",
		1500,
	); err != nil {
		panic(err)
	}

	// decocdedWif just for signing the base tx. It isn't relevant to myAccount or merchantAccount,
	// and can be ignored.
	decodedWif, err := wif.DecodeWIF("KznvCNc6Yf4iztSThoMH6oHWzH9EgjfodKxmeuUGPq5DEX5maspS")
	if err != nil {
		panic(err)
	}

	// Get three destinations (p2pkh scripts) from our account and add them to the baseTx.
	// We must get these from the account we're sending to, to allow the account to build its
	// own internal mapping for later spending.
	for i := 0; i < 3; i++ {
		destination := myAccount.createDestination()
		if err := baseTx.AddP2PKHOutputFromScript(destination, 400); err != nil {
			panic(err)
		}
	}

	changeScript, err := bscript.NewP2PKHFromPubKeyEC(decodedWif.PrivKey.PubKey())
	if err != nil {
		panic(err)
	}

	if err = baseTx.Change(changeScript, bt.NewFeeQuote()); err != nil {
		panic(err)
	}

	if err = baseTx.FillInput(
		context.Background(),
		&unlocker.Simple{PrivateKey: decodedWif.PrivKey},
		bt.UnlockerParams{},
	); err != nil {
		panic(err)
	}

	// Here we would broadcast the transaction, to the account. Assume this has been done
	// and that we are starting anew, only this time it's the account who is building and
	// broadcasting a transaction to elsewhere.

	// Create the tx which we are going to send to the merchant.
	tx := bt.NewTx()
	// Add the three UTXOs from the baseTx for funding.
	for i := 0; i < 3; i++ {
		if err = tx.From(baseTx.TxID(), uint32(i), baseTx.Outputs[i].LockingScript.String(), 400); err != nil {
			panic(err)
		}
	}

	if err := tx.AddP2PKHOutputFromScript(merchantAccount.createDestination(), 1000); err != nil {
		panic(err)
	}

	if err := tx.Change(myAccount.createDestination(), bt.NewFeeQuote()); err != nil {
		panic(err)
	}

	// Call fill all inputs and pass in the signing account as the UnlockerGetter. The account
	// struct implements `bt.UnlockerGetter`.
	if err := tx.FillAllInputs(context.Background(), myAccount); err != nil {
		panic(err)
	}
}
