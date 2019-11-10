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

	b.api.setAccessToken(accessToken)
	b.bniSessID = shortuuid.New()
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
	funcLog := logger.Logger(bniCtx.WithBniSessId(ctx, b.bniSessID))

	funcLog.Info("=== DO_AUTH ===")

	dtoResp, err := b.api.postGetToken(ctx)
	if err != nil {
		funcLog.Error(errors.Details(err))
		return nil, errors.Trace(err)
	}

	b.setAccessToken(dtoResp.AccessToken)

	funcLog = logger.Logger(bniCtx.WithBniSessId(ctx, b.bniSessID))
	funcLog.Info("=== END DO_AUTH ===")

	return dtoResp, nil
}

func (b *BNI) GetBalance(ctx context.Context, dtoReq *dto.GetBalanceRequest) (*dto.GetBalanceResponse, error) {
	funcLog := logger.Logger(bniCtx.WithBniSessId(ctx, b.bniSessID))

	funcLog.Info("=== GET_BALANCE ===")

	dtoReq.ClientID = b.config.ClientID
	if err := b.buildSignatureGetBalance(dtoReq); err != nil {
		funcLog.Error(errors.Details(err))
		return nil, errors.Trace(err)
	}

	logReq := dto.BuildLogRequest(BalanceRequest, "", "", dtoReq)
	funcLog.Infof("%+v", logReq)

	dtoResp, err := b.api.postGetBalance(ctx, dtoReq)
	if err != nil {
		funcLog.Error(errors.Details(err))
		return nil, errors.Trace(err)
	}

	logResp := dto.BuildLogResponse(BalanceResponse, "", "", dtoResp)
	funcLog.Infof("%+v", logResp)

	funcLog.Info("=== END GET_BALANCE ===")

	if dtoResp.GetBalanceResponse == nil {
		return nil, errors.New("GetBalance: bad response")
	}

	return dtoResp.GetBalanceResponse, nil
}

// === Signature of each request ===

func (b *BNI) buildSignatureGetBalance(dtoReq *dto.GetBalanceRequest) error {
	sign, err := b.signature.Sha256WithRSA(dtoReq.ClientID + dtoReq.AccountNo)
	if err != nil {
		return errors.Trace(err)
	}
	dtoReq.Signature = sign

	return nil
}
