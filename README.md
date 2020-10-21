# LiBSV

The go-to Bitcoin GoLang library.  

For more information around the technical aspects of Bitcoin, please see the updated [Bitcoin Wiki](https://wiki.bitcoinsv.io/index.php/Main_Page).

## Documentation

Check the [GoDoc](https://pkg.go.dev/mod/github.com/libsv/libsv) documentation.

## Installation

**Install with [go](https://formulae.brew.sh/formula/go)**

```console
$ go get github.com/libsv/libsv
```

## Tests

Run all tests
```console
$ make test
```

## Examples

### Create a transaction

#### Regular P2PKH
```go
	tx := bt.New()

	tx.From(
		"11b476ad8e0a48fcd40807a111a050af51114877e09283bfa7f3505081a1819d",
		0,
		"76a914eb0bd5edba389198e73f8efabddfc61666969ff788ac6a0568656c6c6f",
		1500)

	tx.PayTo("1NRoySJ9Lvby6DuE2UQYnyT67AASwNZxGb", 1000)

	wif, _ := bsvutil.DecodeWIF("KznvCNc6Yf4iztSThoMH6oHWzH9EgjfodKxmeuUGPq5DEX5maspS")

	signer := sig.InternalSigner{PrivateKey: wif.PrivKey, SigHashFlag: 0}
  err := tx.SignAuto(&signer)
  if err != nil {
		fmt.Errorf(err.Error())
	}

	fmt.Println(tx.ToString())
```

prints:
```console
01000000019d81a1815050f3a7bf8392e077481151af50a011a10708d4fc480a8ead76b411000000006b483045022100dda18196d5217ecfe01390a7ec9c0bd577e7d97ed88f92b7c4a2bf8cb94a493b0220465f9ab035ae584d45c0fbb41363c1cd862b8439619b3b42decb1e9f556dd142412102798913bc057b344de675dac34faafe3dc2f312c758cd9068209f810877306d66ffffffff01e8030000000000001976a914eb0bd5edba389198e73f8efabddfc61666969ff788ac00000000
```

#### Regular P2PKH + OP_RETURN output
```go
	tx := bt.New()

	err := tx.From(
		"b7b0650a7c3a1bd4716369783876348b59f5404784970192cec1996e86950576",
		0,
		"76a9149cbe9f5e72fa286ac8a38052d1d5337aa363ea7f88ac",
		1000)

	tx.PayTo("1C8bzHM8XFBHZ2ZZVvFy2NSoAZbwCXAicL", 900)

	o, err := output.NewOpReturn([]byte("You are using LiBSV!"))
	if err != nil {
		fmt.Println(err.Error())
	}

	tx.AddOutput(o)

	wif, _ := bsvutil.DecodeWIF("L3VJH2hcRGYYG6YrbWGmsxQC1zyYixA82YjgEyrEUWDs4ALgk8Vu")

	signer := sig.InternalSigner{PrivateKey: wif.PrivKey, SigHashFlag: 0}
	err = tx.SignAuto(&signer)
	if err != nil {
		fmt.Println(err.Error())
	}

	fmt.Println(tx.ToString())
```

prints:
```console
0100000001760595866e99c1ce920197844740f5598b34763878696371d41b3a7c0a65b0b7000000006b48304502210095087fccf657f236ffc844d97d5a3a0c43c96972ff00a842b31cb1905e11de4a022074a41d90c548bde1fff9de3c85dd9f773ba64de26b4de2dfe2bef812ab8de23b412102ea87d1fd77d169bd56a71e700628113d0f8dfe57faa0ba0e55a36f9ce8e10be3ffffffff0284030000000000001976a9147a1980655efbfec416b2b0c663a7b3ac0b6a25d288ac000000000000000017006a14596f7520617265207573696e67204c694253562100000000
```

## Contributing
View the [contributing guidelines](CONTRIBUTING.md) and please follow the [code of conduct](CODE_OF_CONDUCT.md).
