package transaction

import (
	"encoding/hex"
	"errors"
)

// SigningItem struct
type SigningItem struct {
	PublicKeyHash string `json:"publicKeyHash"`
	SigHash       string `json:"sigHash"`
	PublicKey     string `json:"publicKey,omitempty"`
	Signature     string `json:"signature,omitempty"`
}

// SigningPayload type
type SigningPayload []*SigningItem

// NewSigningPayload comment
func NewSigningPayload() *SigningPayload {
	sp := make([]*SigningItem, 0)
	p := SigningPayload(sp)
	return &p
}

// NewSigningPayloadFromTx create a signinbg payload for a transaction.
func NewSigningPayloadFromTx(bt *BitcoinTransaction) (*SigningPayload, error) {
	p := NewSigningPayload()
	for idx, input := range bt.Inputs {
		if input.PreviousTxSatoshis == 0 {
			return nil, errors.New("Error getting sighashes - Inputs need to have a PreviousTxSatoshis set to be signable")
		}

		if input.PreviousTxScript == nil {
			return nil, errors.New("Error getting sighashes - Inputs need to have a PreviousScript to be signable")

		}

		sighash := sighashForForkID(bt, SighashAllForkID, uint32(idx), *input.PreviousTxScript, input.PreviousTxSatoshis)
		pkh, err := input.PreviousTxScript.GetPublicKeyHash()
		if err != nil {
			return nil, err
		}
		p.AddItem(hex.EncodeToString(pkh), hex.EncodeToString(sighash))
	}
	return p, nil
}

// AddItem function
func (sp *SigningPayload) AddItem(publicKeyHash string, sigHash string) {
	si := &SigningItem{
		PublicKeyHash: publicKeyHash,
		SigHash:       sigHash,
	}

	*sp = append(*sp, si)
}
