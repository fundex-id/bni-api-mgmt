package bni

import (
	"context"
	"net/http"
	"os"
	"sync"

	"github.com/fundex-id/bni-api-mgmt/config"
	bniCtx "github.com/fundex-id/bni-api-mgmt/context"
	"github.com/fundex-id/bni-api-mgmt/dto"
	"github.com/fundex-id/bni-api-mgmt/logger"
	"github.com/fundex-id/bni-api-mgmt/signature"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/juju/errors"
	"github.com/lithammer/shortuuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var BadResponseError error = errors.New("Bad response")

type BNI struct {
	api       *API
	config    config.Config
	signature *signature.Signature

	mutex     sync.Mutex
	bniSessID string
}

func New(config config.Config) *BNI {
	bni := BNI{
		config:    config,
		api:       newApi(config),
		signature: signature.New(config.SignatureConfig),
	}

	logger.SetOptions(zap.WrapCore(func(core zapcore.Core) zapcore.Core {

		fileWriteSyncer := zapcore.AddSync(&lumberjack.Logger{
			Filename: config.LogPath,
			MaxSize:  500, // megabytes
			// MaxBackups: 3,
			// MaxAge:     28, // days
		})
		stdoutWriteSyncer := zapcore.AddSync(os.Stdout)

		return zapcore.NewCore(
			zapcore.NewJSONEncoder(logger.DefaultEncoderConfig),
			zapcore.NewMultiWriteSyncer(fileWriteSyncer, stdoutWriteSyncer),
			zap.InfoLevel,
		)

		// return core
	}))

	retryablehttpClient := retryablehttp.NewClient()
	retryablehttpClient.RetryMax = 1
	retryablehttpClient.CheckRetry = bni.retryPolicy

	bni.api.retryablehttpClient = retryablehttpClient

	return &bni
}

func (b *BNI) setAccessToken(accessToken string) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	newSessID := shortuuid.New()
	b.api.setAccessTokenAndSessID(accessToken, newSessID)
	b.bniSessID = newSessID
}

func (b *BNI) retryPolicy(ctx context.Context, resp *http.Response, err error) (bool, error) {
	// do not retry on context.Canceled or context.DeadlineExceeded
	if ctx.Err() != nil {
		return false, ctx.Err()
	}

	if resp.StatusCode == http.StatusUnauthorized {
		b.log(ctx).Infof("Retry to auth. Got resp (code: %d, status: %s). Prev err: %+v", resp.StatusCode, resp.Status, err)
		_, errAuth := b.DoAuthentication(ctx)

		return true, errAuth
	}

	return false, nil
}

// === APi based on spec ===

func (b *BNI) DoAuthentication(ctx context.Context) (*dto.GetTokenResponse, error) {
	b.log(ctx).Info("=== DO_AUTH ===")

	dtoResp, err := b.api.postGetToken(ctx)
	if err != nil {
		b.log(ctx).Error(errors.Details(err))
		return nil, errors.Trace(err)
	}

	b.setAccessToken(dtoResp.AccessToken)

	b.log(ctx).Info("=== END DO_AUTH ===")

	return dtoResp, nil
}

func (b *BNI) GetBalance(ctx context.Context, dtoReq *dto.GetBalanceRequest) (*dto.GetBalanceResponse, error) {
	ctx = bniCtx.WithBNISessID(ctx, b.bniSessID)

	b.log(ctx).Info("=== GET_BALANCE ===")

	dtoReq.ClientID = b.config.ClientID
	if err := b.setSignatureGetBalance(dtoReq); err != nil {
		b.log(ctx).Error(errors.Details(err))
		return nil, errors.Trace(err)
	}

	logReq := dto.BuildLogRequest(BalanceRequest, "", "", dtoReq)
	b.log(ctx).Infof("%+v", logReq)

	dtoResp, err := b.api.postGetBalance(ctx, dtoReq)
	if err != nil {
		b.log(ctx).Error(errors.Details(err))
		return nil, errors.Trace(err)
	}

	logResp := dto.BuildLogResponse(BalanceResponse, "", "", dtoResp)
	b.log(ctx).Infof("%+v", logResp)

	if dtoResp.GetBalanceResponse == nil {
		b.log(ctx).Error(BadResponseError)
		return nil, BadResponseError
	}

	b.log(ctx).Info("=== END GET_BALANCE ===")

	return dtoResp.GetBalanceResponse, nil
}

// === misc func ===

func (b *BNI) log(ctx context.Context) *zap.SugaredLogger {
	return logger.Logger(bniCtx.WithBNISessID(ctx, b.bniSessID))
}

// === Signature of each request ===

func (b *BNI) setSignatureGetBalance(dtoReq *dto.GetBalanceRequest) error {
	sign, err := b.signature.Sha256WithRSA(dtoReq.ClientID + dtoReq.AccountNo)
	if err != nil {
		return errors.Trace(err)
	}
	dtoReq.Signature = sign

	return nil
}
