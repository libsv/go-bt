# go-bt

> The go-to Bitcoin Transaction (BT) GoLang library  

[![Release](https://img.shields.io/github/release-pre/libsv/go-bt.svg?logo=github&style=flat&v=1)](https://github.com/libsv/go-bt/releases)
[![Build Status](https://img.shields.io/github/workflow/status/libsv/go-bt/run-go-tests?logo=github&v=3)](https://github.com/libsv/go-bt/actions)
[![Report](https://goreportcard.com/badge/github.com/libsv/go-bt?style=flat&v=1)](https://goreportcard.com/report/github.com/libsv/go-bt)
[![codecov](https://codecov.io/gh/libsv/go-bt/branch/master/graph/badge.svg?v=1)](https://codecov.io/gh/libsv/go-bt)
[![Go](https://img.shields.io/github/go-mod/go-version/libsv/go-bt?v=1)](https://golang.org/)
[![Sponsor](https://img.shields.io/badge/sponsor-libsv-181717.svg?logo=github&style=flat&v=3)](https://github.com/sponsors/libsv)
[![Donate](https://img.shields.io/badge/donate-bitcoin-ff9900.svg?logo=bitcoin&style=flat&v=3)](https://gobitcoinsv.com/#sponsor)

<br/>

## Table of Contents

- [Installation](#installation)
- [Documentation](#documentation)
- [Examples & Tests](#examples--tests)
- [Benchmarks](#benchmarks)
- [Code Standards](#code-standards)
- [Usage](#usage)
- [Maintainers](#maintainers)
- [Contributing](#contributing)
- [License](#license)

<br/>

## Installation

**go-bt** requires a [supported release of Go](https://golang.org/doc/devel/release.html#policy).

```shell script
go get -u github.com/libsv/go-bt
```

<br/>

## Documentation

View the generated [documentation](https://pkg.go.dev/github.com/libsv/go-bt)

[![GoDoc](https://godoc.org/github.com/libsv/go-bt?status.svg&style=flat)](https://pkg.go.dev/github.com/libsv/go-bt)

For more information around the technical aspects of Bitcoin, please see the updated [Bitcoin Wiki](https://wiki.bitcoinsv.io/index.php/Main_Page)

<br/>

### Features

- Full Featured Bitcoin Transactions
- Auto-Fee Calculations for Change Address
- Bitcoin Transaction [Script](bscript/) Functionality
  - P2PKH (base58 addresses)
  - Data (OP_RETURN)
  - [BIP276](https://github.com/moneybutton/bips/blob/master/bip-0276.mediawiki)

- Transaction Signing Extendability

#### Coming Soon! (18 months<sup>TM</sup>)

- Complete SigHash Flag Capability
- MultiSig functionality

<details>
<summary><strong><code>Library Deployment</code></strong></summary>
<br/>

[goreleaser](https://github.com/goreleaser/goreleaser) for easy binary or library deployment to Github and can be installed via: `brew install goreleaser`.

The [.goreleaser.yml](.goreleaser.yml) file is used to configure [goreleaser](https://github.com/goreleaser/goreleaser).

Use `make release-snap` to create a snapshot version of the release, and finally `make release` to ship to production.
</details>

<details>
<summary><strong><code>Makefile Commands</code></strong></summary>
<br/>

View all `makefile` commands

```shell script
make help
```

List of all current commands:

```text
all                  Runs multiple commands
clean                Remove previous builds and any test cache data
clean-mods           Remove all the Go mod cache
coverage             Shows the test coverage
godocs               Sync the latest tag with GoDocs
help                 Show this help message
install              Install the application
install-go           Install the application (Using Native Go)
lint                 Run the golangci-lint application (install if not found)
release              Full production release (creates release in Github)
release              Runs common.release then runs godocs
release-snap         Test the full release (build binaries)
release-test         Full production test release (everything except deploy)
replace-version      Replaces the version in HTML/JS (pre-deploy)
tag                  Generate a new tag and push (tag version=0.0.0)
tag-remove           Remove a tag if found (tag-remove version=0.0.0)
tag-update           Update an existing tag to current commit (tag-update version=0.0.0)
test                 Runs vet, lint and ALL tests
test-ci              Runs all tests via CI (exports coverage)
test-ci-no-race      Runs all tests via CI (no race) (exports coverage)
test-ci-short        Runs unit tests via CI (exports coverage)
test-short           Runs vet, lint and tests (excludes integration tests)
uninstall            Uninstall the application (and remove files)
update-linter        Update the golangci-lint package (macOS only)
vet                  Run the Go vet application
```

</details>

<br/>

## Examples & Tests

All unit tests and [examples](examples) run via [Github Actions](https://github.com/libsv/go-bt/actions) and
uses [Go version 1.15.x](https://golang.org/doc/go1.15). View the [configuration file](.github/workflows/run-tests.yml).

Run all tests (including integration tests)

```shell script
make test
```

Run tests (excluding integration tests)

```shell script
make test-short
```

<br/>

## Benchmarks

Run the Go [benchmarks](tx_test.go):

```shell script
make bench
```

<br/>

## Code Standards

Read more about this Go project's [code standards](CODE_STANDARDS.md).

<br/>

## Usage

View the [examples](examples)

<br/>

## Maintainers

| [<img src="https://github.com/HaddadJoe.png" height="50" alt="JH" />](https://github.com/HaddadJoe) | [<img src="https://github.com/jadwahab.png" height="50" alt="JW" />](https://github.com/jadwahab) | [<img src="https://github.com/mrz1836.png" height="50" alt="MrZ" />](https://github.com/mrz1836) |
|:---:|:---:|:---:|
|  [JH](https://github.com/HaddadJoe)| [JW](https://github.com/jadwahab) | [MrZ](https://github.com/mrz1836) |

<br/>

## Contributing

View the [contributing guidelines](CONTRIBUTING.md) and please follow the [code of conduct](CODE_OF_CONDUCT.md).

### How can I help?

All kinds of contributions are welcome :raised_hands:!
The most basic way to show your support is to star :star2: the project, or to raise issues :speech_balloon:.
You can also support this project by [becoming a sponsor on GitHub](https://github.com/sponsors/libsv) :clap:
or by making a [**bitcoin donation**](https://gobitcoinsv.com/#sponsor) to ensure this journey continues indefinitely! :rocket:

<br/>

## License

![License](https://img.shields.io/github/license/libsv/go-bt.svg?style=flat&v=1)
