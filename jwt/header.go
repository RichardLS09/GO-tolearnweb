package jwt

// Header is jwt's header
type Header struct {
	Type  string `json:"typ,omitempty"`
	Alg   string `json:"alg,omitempty"`
	KeyId string `json:"kid,omitempty"`
}

// NewHeader can build a Header entity pointer with algorithm
func NewHeader(alg Alg) *Header {
	return &Header{
		Type: "JWT",
		Alg:  alg.Name(),
	}
}
