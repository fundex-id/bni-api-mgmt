package bni

type Signature struct {
	config SignatureConfig
}

func newSignature(config SignatureConfig) *Signature {
	return &Signature{config: config}
}

func (s *Signature) sha256WithRsa(data string) (string, error) {

}
