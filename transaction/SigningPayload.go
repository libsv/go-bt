package transaction

// SigningItem struct
type SigningItem struct {
	Address   string  `json:"address"`
	SigHash   string  `json:"sigHash"`
	PublicKey *string `json:"publicKey"`
	Signature *string `json:"signature"`
}

// SigningPayload type
type SigningPayload []*SigningItem

// NewSigningPayload comment
func NewSigningPayload() SigningPayload {
	sp := make([]*SigningItem, 0)
	p := SigningPayload(sp)
	return p
}

// AddItem function
func (sp *SigningPayload) AddItem(address string, sigHash string) {
	si := &SigningItem{
		Address: address,
		SigHash: sigHash,
	}

	*sp = append(*sp, si)
}
