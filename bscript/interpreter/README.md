interpreter
========

[![ISC License](http://img.shields.io/badge/license-ISC-blue.svg)](http://copyfree.org)
[![GoDoc](https://pkg.go.dev/badge/github.com/libsv/go-bt/bscript/interpreter?utm_source=godoc)](http://godoc.org/github.com/libsv/got-bt/bscript/interpreter)

Package interpreter implements the an interpreter for the bitcoin transaction language.  There is
a comprehensive test suite.

This package has intentionally been designed so it can be used as a standalone
package for any projects needing to use or validate bitcoin transaction scripts.

## Bitcoin Scripts

Bitcoin provides a stack-based, FORTH-like language for the scripts in
the bitcoin transactions.  This language is not turing complete
although it is still fairly powerful.  A description of the language
can be found at https://wiki.bitcoinsv.io/index.php/Script

## Installation and Updating

```bash
$ go get -u github.com/libsv/go-bt/bscript/interpreter
```

## Examples

* [Standard Pay-to-pubkey-hash Script](http://github.com/libsv/go-bt/bscript/interpreter#example-PayToAddrScript)  
  Demonstrates creating a script which pays to a bitcoin address.  It also
  prints the created script hex and uses the DisasmString function to display
  the disassembled script.

## License

Package interpreter is licensed under the [copyfree](http://copyfree.org) ISC
License.
