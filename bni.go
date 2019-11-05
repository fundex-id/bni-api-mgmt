package bni

import (
	"context"

	"github.com/fundex-id/bni-api-mgmt/dto"
	"github.com/juju/errors"
	"github.com/pborman/uuid"
)

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

func (b *BNI) DoAuthentication(ctx context.Context) (*dto.GetTokenResponse, error) {
	resp, err := b.api.postGetToken(ctx)
	if err != nil {
		return nil, errors.Trace(err)
	}

	resp.Session.ID = uuid.NewRandom().String()

	return resp, nil
}
