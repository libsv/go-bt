package script

import "github.com/libsv/libsv/transaction/signature/sighash"

// NewUnlockingScriptForP2PKHBytes creates a new unlocking script which spends a P2PKH locking script from a public key, a signature, and
// a SIGHASH flag.
func NewUnlockingScriptForP2PKHBytes(pubKey []byte, sig []byte, sigHashFlag sighash.Flag) (*Script, error) {
	// append SIGHASH to DER sig
	sigBuf := make([]byte, 0)
	sigBuf = append(sigBuf, sig...)
	sigBuf = append(sigBuf, uint8(sigHashFlag))

	scriptBuf := [][]byte{sigBuf, pubKey}

	s := &Script{}
	err := s.AppendPushDataArray(scriptBuf)

	return s, err
}
