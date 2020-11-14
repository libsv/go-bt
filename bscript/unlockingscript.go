package bscript

import "github.com/libsv/go-bt/sighash"

// NewP2PKHUnlockingScript creates a new unlocking script which spends
// a P2PKH locking script from a public key, a signature, and
// a SIGHASH flag.
func NewP2PKHUnlockingScript(pubKey []byte, sig []byte,
	sigHashFlag sighash.Flag) (*Script, error) {

	// append SIGHASH to DER sig
	sigBuf := []byte{}
	sigBuf = append(sigBuf, sig...)
	sigBuf = append(sigBuf, uint8(sigHashFlag))

	scriptBuf := [][]byte{sigBuf, pubKey}

	s := &Script{}
	err := s.AppendPushDataArray(scriptBuf)

	return s, err
}
