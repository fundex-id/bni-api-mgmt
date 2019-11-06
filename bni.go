package bni

import (
	"context"
	"net/http"
	"sync"

	"github.com/fundex-id/bni-api-mgmt/config"
	"github.com/fundex-id/bni-api-mgmt/dto"
	"github.com/fundex-id/bni-api-mgmt/logger"
	"github.com/fundex-id/bni-api-mgmt/signature"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/juju/errors"
	"github.com/pborman/uuid"
)

type BNI struct {
	api       *API
	config    config.Config
	signature *signature.Signature

	mutex       sync.Mutex
	accessToken string
	bniSessID   string
}

func New(config config.Config) *BNI {
	bni := BNI{
		config:    config,
		api:       newApi(config),
		signature: signature.New(config.SignatureConfig),
	}

	retryablehttpClient := retryablehttp.NewClient()
	retryablehttpClient.RetryMax = 2
	retryablehttpClient.CheckRetry = bni.retryPolicy

	bni.api.retryablehttpClient = retryablehttpClient

	return &bni
}

func (b *BNI) setAccessToken(accessToken string) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	b.accessToken = accessToken
	b.bniSessID = uuid.NewRandom().String()
}

func (b *BNI) retryPolicy(ctx context.Context, resp *http.Response, err error) (bool, error) {
	// do not retry on context.Canceled or context.DeadlineExceeded
	if ctx.Err() != nil {
		return false, ctx.Err()
	}

	funcLog := logger.Logger(ctx)

	if resp.StatusCode == 401 {
		funcLog.Infof("Retry to auth. Prev err: %+v", err)
		_, errAuth := b.DoAuthentication(ctx)

		return true, errAuth
	}

	return false, nil
}

// === APi based on spec ===

func (b *BNI) DoAuthentication(ctx context.Context) (*dto.GetTokenResponse, error) {
	resp, err := b.api.postGetToken(ctx)
	if err != nil {
		return nil, errors.Trace(err)
	}

	b.setAccessToken(resp.AccessToken)

	return resp, nil
}
