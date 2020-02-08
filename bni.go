package bni

import (
	"context"
	"net/http"
	"os"

	"github.com/fundex-id/bni-api-mgmt/config"
	bniCtx "github.com/fundex-id/bni-api-mgmt/context"
	"github.com/fundex-id/bni-api-mgmt/dto"
	"github.com/fundex-id/bni-api-mgmt/logger"
	"github.com/fundex-id/bni-api-mgmt/signature"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/juju/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var BadResponseError error = errors.New("Bad response")

type BNI struct {
	api       *API
	config    config.Config
	signature *signature.Signature
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

	b.api.setAccessToken(dtoResp.AccessToken)

	b.log(ctx).Info("=== END DO_AUTH ===")

	return dtoResp, nil
}

func (b *BNI) GetBalance(ctx context.Context, dtoReq *dto.GetBalanceRequest) (*dto.GetBalanceResponse, error) {
	ctx = bniCtx.WithBNISessID(ctx, b.api.bniSessID)

	b.log(ctx).Info("=== GET_BALANCE ===")

	dtoReq.ClientID = b.config.ClientID
	if err := b.setSignatureGetBalance(dtoReq); err != nil {
		b.log(ctx).Error(errors.Details(err))
		return nil, errors.Trace(err)
	}

	logReq := dto.BuildLogRequest(BalanceRequest, dtoReq)
	b.log(ctx).Infof("%+v", logReq)

	dtoResp, err := b.api.postGetBalance(ctx, dtoReq)
	if err != nil {
		b.log(ctx).Error(errors.Details(err))
		return nil, errors.Trace(err)
	}

	logResp := dto.BuildLogResponse(BalanceResponse, dtoResp)
	b.log(ctx).Infof("%+v", logResp)

	dtoParamResp := dtoResp.GetBalanceResponse
	if dtoParamResp == nil {
		b.log(ctx).Error(BadResponseError)
		return nil, BadResponseError
	}

	b.log(ctx).Info("=== END GET_BALANCE ===")

	return dtoParamResp, nil
}

func (b *BNI) GetInHouseInquiry(ctx context.Context, dtoReq *dto.GetInHouseInquiryRequest) (*dto.GetInHouseInquiryResponse, error) {
	ctx = bniCtx.WithBNISessID(ctx, b.api.bniSessID)

	b.log(ctx).Info("=== GET_IN_HOUSE_INQUIRY ===")

	dtoReq.ClientID = b.config.ClientID
	if err := b.setSignatureGetInHouseInquiry(dtoReq); err != nil {
		b.log(ctx).Error(errors.Details(err))
		return nil, errors.Trace(err)
	}

	logReq := dto.BuildLogRequest(InHouseInquiryRequest, dtoReq)
	b.log(ctx).Infof("%+v", logReq)

	dtoResp, err := b.api.postGetInHouseInquiry(ctx, dtoReq)
	if err != nil {
		b.log(ctx).Error(errors.Details(err))
		return nil, errors.Trace(err)
	}

	logResp := dto.BuildLogResponse(InHouseInquiryResponse, dtoResp)
	b.log(ctx).Infof("%+v", logResp)

	dtoParamResp := dtoResp.GetInHouseInquiryResponse
	if dtoParamResp == nil {
		b.log(ctx).Error(BadResponseError)
		return nil, BadResponseError
	}

	b.log(ctx).Info("=== END GET_IN_HOUSE_INQUIRY ===")

	return dtoParamResp, nil
}

func (b *BNI) DoPayment(ctx context.Context, dtoReq *dto.DoPaymentRequest) (*dto.DoPaymentResponse, error) {
	ctx = bniCtx.WithBNISessID(ctx, b.api.bniSessID)

	b.log(ctx).Info("=== DO_PAYMENT ===")

	dtoReq.ClientID = b.config.ClientID
	if err := b.setSignatureDoPayment(dtoReq); err != nil {
		b.log(ctx).Error(errors.Details(err))
		return nil, errors.Trace(err)
	}

	logReq := dto.BuildLogRequest(InHouseTransferRequest, dtoReq)
	b.log(ctx).Infof("%+v", logReq)

	dtoResp, err := b.api.postDoPayment(ctx, dtoReq)
	if err != nil {
		b.log(ctx).Error(errors.Details(err))
		return nil, errors.Trace(err)
	}

	logResp := dto.BuildLogResponse(InHouseTransferResponse, dtoResp)
	b.log(ctx).Infof("%+v", logResp)

	dtoParamResp := dtoResp.DoPaymentResponse
	if dtoParamResp == nil {
		b.log(ctx).Error(BadResponseError)
		return nil, BadResponseError
	}

	b.log(ctx).Info("=== END DO_PAYMENT ===")

	return dtoParamResp, nil
}

func (b *BNI) GetPaymentStatus(ctx context.Context, dtoReq *dto.GetPaymentStatusRequest) (*dto.GetPaymentStatusResponse, error) {
	ctx = bniCtx.WithBNISessID(ctx, b.api.bniSessID)

	b.log(ctx).Info("=== GET_PAYMENT_STATUS ===")

	dtoReq.ClientID = b.config.ClientID
	if err := b.setSignatureGetPaymentStatus(dtoReq); err != nil {
		b.log(ctx).Error(errors.Details(err))
		return nil, errors.Trace(err)
	}

	logReq := dto.BuildLogRequest(PaymentStatusRequest, dtoReq)
	b.log(ctx).Infof("%+v", logReq)

	dtoResp, err := b.api.postGetPaymentStatus(ctx, dtoReq)
	if err != nil {
		b.log(ctx).Error(errors.Details(err))
		return nil, errors.Trace(err)
	}

	logResp := dto.BuildLogResponse(PaymentStatusResponse, dtoResp)
	b.log(ctx).Infof("%+v", logResp)

	dtoParamResp := dtoResp.GetPaymentStatusResponse
	if dtoParamResp == nil {
		b.log(ctx).Error(BadResponseError)
		return nil, BadResponseError
	}

	b.log(ctx).Info("=== END GET_PAYMENT_STATUS ===")

	return dtoParamResp, nil
}

func (b *BNI) GetInterBankInquiry(ctx context.Context, dtoReq *dto.GetInterBankInquiryRequest) (*dto.GetInterBankInquiryResponse, error) {
	ctx = bniCtx.WithBNISessID(ctx, b.api.bniSessID)

	b.log(ctx).Info("=== GET_INTER_BANK_INQUIRY ===")

	dtoReq.ClientID = b.config.ClientID
	if err := b.setSignatureGetInterBankInquiry(dtoReq); err != nil {
		b.log(ctx).Error(errors.Details(err))
		return nil, errors.Trace(err)
	}

	logReq := dto.BuildLogRequest(InterBankInquiryRequest, dtoReq)
	b.log(ctx).Infof("%+v", logReq)

	dtoResp, err := b.api.postGetInterBankInquiry(ctx, dtoReq)
	if err != nil {
		b.log(ctx).Error(errors.Details(err))
		return nil, errors.Trace(err)
	}

	logResp := dto.BuildLogResponse(InterBankInquiryResponse, dtoResp)
	b.log(ctx).Infof("%+v", logResp)

	dtoParamResp := dtoResp.GetInterBankInquiryResponse
	if dtoParamResp == nil {
		b.log(ctx).Error(BadResponseError)
		return nil, BadResponseError
	}

	b.log(ctx).Info("=== END GET_INTER_BANK_INQUIRY ===")

	return dtoParamResp, nil
}

func (b *BNI) GetInterBankPayment(ctx context.Context, dtoReq *dto.GetInterBankPaymentRequest) (*dto.GetInterBankPaymentResponse, error) {
	ctx = bniCtx.WithBNISessID(ctx, b.api.bniSessID)

	b.log(ctx).Info("=== GET_INTER_BANK_PAYMENT ===")

	dtoReq.ClientID = b.config.ClientID
	if err := b.setSignatureGetInterBankPayment(dtoReq); err != nil {
		b.log(ctx).Error(errors.Details(err))
		return nil, errors.Trace(err)
	}

	logReq := dto.BuildLogRequest(InterBankTransferRequest, dtoReq)
	b.log(ctx).Infof("%+v", logReq)

	dtoResp, err := b.api.postGetInterBankPayment(ctx, dtoReq)
	if err != nil {
		b.log(ctx).Error(errors.Details(err))
		return nil, errors.Trace(err)
	}

	logResp := dto.BuildLogResponse(InterBankTransferResponse, dtoResp)
	b.log(ctx).Infof("%+v", logResp)

	dtoParamResp := dtoResp.GetInterBankPaymentResponse
	if dtoParamResp == nil {
		b.log(ctx).Error(BadResponseError)
		return nil, BadResponseError
	}

	b.log(ctx).Info("=== END GET_INTER_BANK_PAYMENT ===")

	return dtoParamResp, nil
}

// === misc func ===

func (b *BNI) log(ctx context.Context) *zap.SugaredLogger {
	return logger.Logger(bniCtx.WithBNISessID(ctx, b.api.bniSessID))
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

func (b *BNI) setSignatureGetInHouseInquiry(dtoReq *dto.GetInHouseInquiryRequest) error {
	sign, err := b.signature.Sha256WithRSA(dtoReq.ClientID + dtoReq.AccountNo)
	if err != nil {
		return errors.Trace(err)
	}
	dtoReq.Signature = sign

	return nil
}

func (b *BNI) setSignatureDoPayment(dtoReq *dto.DoPaymentRequest) error {
	sign, err := b.signature.Sha256WithRSA(
		dtoReq.ClientID +
			dtoReq.CustomerReferenceNumber +
			dtoReq.PaymentMethod +
			dtoReq.DebitAccountNo +
			dtoReq.CreditAccountNo +
			dtoReq.ValueAmount +
			dtoReq.ValueCurrency,
	)

	if err != nil {
		return errors.Trace(err)
	}
	dtoReq.Signature = sign

	return nil
}

func (b *BNI) setSignatureGetPaymentStatus(dtoReq *dto.GetPaymentStatusRequest) error {
	sign, err := b.signature.Sha256WithRSA(
		dtoReq.ClientID +
			dtoReq.CustomerReferenceNumber,
	)

	if err != nil {
		return errors.Trace(err)
	}
	dtoReq.Signature = sign

	return nil
}

func (b *BNI) setSignatureGetInterBankInquiry(dtoReq *dto.GetInterBankInquiryRequest) error {
	sign, err := b.signature.Sha256WithRSA(
		dtoReq.ClientID +
			dtoReq.DestinationBankCode +
			dtoReq.DestinationAccountNum +
			dtoReq.AccountNum,
	)

	if err != nil {
		return errors.Trace(err)
	}
	dtoReq.Signature = sign

	return nil
}

func (b *BNI) setSignatureGetInterBankPayment(dtoReq *dto.GetInterBankPaymentRequest) error {
	sign, err := b.signature.Sha256WithRSA(
		dtoReq.ClientID +
			dtoReq.DestinationAccountNum +
			dtoReq.DestinationBankCode +
			dtoReq.AccountNum +
			dtoReq.Amount +
			dtoReq.RetrievalReffNum,
	)

	if err != nil {
		return errors.Trace(err)
	}
	dtoReq.Signature = sign

	return nil
}
