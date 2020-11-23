package bt

// Signer interface to allow custom implementations of different signing mechanisms.
// Implement the Sign function as shown in InternalSigner, for example.
// Sign function takes an unsigned Tx and returns a signed Tx.
type Signer interface {
	Sign(index uint32, unsignedTx *Tx) (signedTx *Tx, err error)
	SignAuto(unsignedTx *Tx) (signedTx *Tx, inputsSigned []int, err error)
}
