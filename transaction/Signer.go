package transaction

// Signer interface to allow custom implementations of different signing mechanisms.
// Implement the Sign function as shown in InternalSigner, for example.
// Sign function takes an unsigned Transaction and returns a signed Transaction.
type Signer interface {
	Sign(index uint32, unsignedTx *Transaction) (*Transaction, error)
	SignAuto(unsignedTx *Transaction) (*Transaction, error)
}
