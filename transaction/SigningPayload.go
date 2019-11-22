package transaction

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

// AddItem function
func (sp *SigningPayload) AddItem(publicKeyHash string, sigHash string) {
	si := &SigningItem{
		PublicKeyHash: publicKeyHash,
		SigHash:       sigHash,
	}

	*sp = append(*sp, si)
}
