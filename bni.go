package bni

type BNI struct {
	api       *API
	config    Config
	signature *Signature
}

func New(config Config) *BNI {
	bni := BNI{
		config:    config,
		api:       newApi(config),
		signature: newSignature(config.SignatureConfig),
	}

	return &bni
}
