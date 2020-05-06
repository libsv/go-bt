package transaction

import (
	"encoding/hex"
	"github.com/libsv/libsv/bsvsuite/bsvec"
	"github.com/libsv/libsv/utils"
)

// InternalSigner implements the Signer interface. It is used to sign a Transaction locally
// given a PrivateKey and SIGHASH type.
type InternalSigner struct {
	PrivateKey *bsvec.PrivateKey
	SigType    uint32
}

// Sign the transaction
func (is *InternalSigner) Sign(unsignedTx *Transaction) (*Transaction, error) {
	if is.SigType == 0 {
		is.SigType = SighashAllForkID
	}

	payload, err := unsignedTx.GetSighashPayload(is.SigType)
	if err != nil {
		return nil, err
	}

	// loops through signing items for each input and signs accordingly
	for _, signingItem := range *payload {
		h, err := hex.DecodeString(signingItem.SigHash)
		if err != nil {
			return nil, err
		}
		sig, err := is.PrivateKey.Sign(utils.ReverseBytes(h))
		if err != nil {
			return nil, err
		}
		pubkey := is.PrivateKey.PubKey().SerializeCompressed()
		signingItem.PublicKey = hex.EncodeToString(pubkey)
		signingItem.Signature = hex.EncodeToString(sig.Serialize())
	}

	// after applying this function we will have a signed transaction
	err = unsignedTx.ApplySignatures(payload, is.SigType)
	if err != nil {
		return nil, err
	}
	return unsignedTx, nil
}
