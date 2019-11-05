package bni

import (
	"context"
	"sync"

	"github.com/fundex-id/bni-api-mgmt/dto"
	"github.com/juju/errors"
	"github.com/pborman/uuid"
)

type BNI struct {
	api       *API
	config    Config
	signature *Signature

	mutex       sync.Mutex
	accessToken string
	bniSessID   string
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

	b.setAccessToken(resp.AccessToken)

	return resp, nil
}

func (b *BNI) setAccessToken(accessToken string) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	b.accessToken = accessToken
	b.bniSessID = uuid.NewRandom().String()
}
