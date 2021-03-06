package bni

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strings"
	"sync"

	"github.com/avast/retry-go"
	"github.com/fundex-id/bni-api-mgmt/config"
	bniCtx "github.com/fundex-id/bni-api-mgmt/context"
	"github.com/fundex-id/bni-api-mgmt/dto"
	"github.com/fundex-id/bni-api-mgmt/logger"
	"github.com/hashicorp/go-cleanhttp"
	"github.com/juju/errors"
	"github.com/lithammer/shortuuid"
	"go.uber.org/zap"
)

var ErrUnauthorized = errors.New("Err StatusUnauthorized")

const (
	AuthPath              string = "/api/oauth/token"
	BalancePath           string = "/H2H/getbalance"
	InHouseInquiryPath    string = "/H2H/getinhouseinquiry"
	InterBankInquiryPath  string = "/H2H/getinterbankinquiry"
	PaymentStatusPath     string = "/H2H/getpaymentstatus"
	InHouseTransferPath   string = "/H2H/dopayment"
	InterBankTransferPath string = "/H2H/getinterbankpayment"
)

type API struct {
	config     config.Config
	httpClient *http.Client // for postGetToken only

	mutex       sync.Mutex
	accessToken string
	bniSessID   string
}

func newApi(config config.Config) *API {
	httpClient := cleanhttp.DefaultPooledClient()
	api := API{config: config,
		httpClient: httpClient,
	}

	return &api
}

func (api *API) setAccessToken(accessToken string) {
	api.mutex.Lock()
	defer api.mutex.Unlock()

	api.accessToken = accessToken
	api.bniSessID = shortuuid.New()
}

func (api *API) postGetToken(ctx context.Context) (*dto.GetTokenResponse, error) {
	urlTarget, err := buildURL(api.config.BNIServer, AuthPath, url.Values{})
	if err != nil {
		return nil, errors.Trace(err)
	}

	form := url.Values{"grant_type": []string{"client_credentials"}}
	bodyReq := strings.NewReader(form.Encode())

	req, err := http.NewRequest(http.MethodPost, urlTarget, bodyReq)
	if err != nil {
		return nil, errors.Trace(err)
	}
	req = req.WithContext(ctx)

	req.Header.Set("content-type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(api.config.Username, api.config.Password)

	resp, err := api.httpClient.Do(req)
	if err != nil {
		return nil, errors.Trace(err)
	}
	defer resp.Body.Close()

	bodyRespBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Trace(err)
	}

	api.log(ctx).Info(resp.StatusCode)
	api.log(ctx).Info(string(bodyRespBytes))

	resp.Body = ioutil.NopCloser(bytes.NewBuffer(bodyRespBytes))

	var dtoResp dto.GetTokenResponse
	err = json.NewDecoder(resp.Body).Decode(&dtoResp)

	if err != nil {
		return nil, errors.Trace(err)
	}

	return &dtoResp, nil
}

func (api *API) doAuthentication(ctx context.Context) (*dto.GetTokenResponse, error) {
	dtoResp, err := api.postGetToken(ctx)
	if err != nil {
		return nil, errors.Trace(err)
	}
	api.setAccessToken(dtoResp.AccessToken)
	return dtoResp, nil
}

func (api *API) postGetBalance(ctx context.Context, dtoReq *dto.GetBalanceRequest) (*dto.ApiResponse, error) {
	jsonReq, err := json.Marshal(dtoReq)
	if err != nil {
		return nil, errors.Trace(err)
	}

	return api.postToAPIWithRetry(ctx, BalancePath, jsonReq)
}

func (api *API) postGetInHouseInquiry(ctx context.Context, dtoReq *dto.GetInHouseInquiryRequest) (*dto.ApiResponse, error) {
	jsonReq, err := json.Marshal(dtoReq)
	if err != nil {
		return nil, errors.Trace(err)
	}

	return api.postToAPIWithRetry(ctx, InHouseInquiryPath, jsonReq)
}

func (api *API) postDoPayment(ctx context.Context, dtoReq *dto.DoPaymentRequest) (*dto.ApiResponse, error) {
	jsonReq, err := json.Marshal(dtoReq)
	if err != nil {
		return nil, errors.Trace(err)
	}

	return api.postToAPIWithRetry(ctx, InHouseTransferPath, jsonReq)
}

func (api *API) postGetPaymentStatus(ctx context.Context, dtoReq *dto.GetPaymentStatusRequest) (*dto.ApiResponse, error) {
	jsonReq, err := json.Marshal(dtoReq)
	if err != nil {
		return nil, errors.Trace(err)
	}

	return api.postToAPIWithRetry(ctx, PaymentStatusPath, jsonReq)
}

func (api *API) postGetInterBankInquiry(ctx context.Context, dtoReq *dto.GetInterBankInquiryRequest) (*dto.ApiResponse, error) {
	jsonReq, err := json.Marshal(dtoReq)
	if err != nil {
		return nil, errors.Trace(err)
	}

	return api.postToAPIWithRetry(ctx, InterBankInquiryPath, jsonReq)
}

func (api *API) postGetInterBankPayment(ctx context.Context, dtoReq *dto.GetInterBankPaymentRequest) (*dto.ApiResponse, error) {
	jsonReq, err := json.Marshal(dtoReq)
	if err != nil {
		return nil, errors.Trace(err)
	}

	return api.postToAPIWithRetry(ctx, InterBankTransferPath, jsonReq)
}

// Generic POST request to API
func (api *API) postToAPI(ctx context.Context, path string, bodyReqPayload []byte) (dtoResp dto.ApiResponse, err error) {
	urlQuery := url.Values{"access_token": []string{api.accessToken}}
	urlTarget, err := buildURL(api.config.BNIServer, path, urlQuery)
	if err != nil {
		return dtoResp, errors.Trace(err)
	}

	req, err := http.NewRequest(http.MethodPost, urlTarget, bytes.NewBuffer(bodyReqPayload))
	if err != nil {
		return dtoResp, errors.Trace(err)
	}
	req = req.WithContext(ctx)

	req.Header.Set("content-type", "application/json")

	resp, err := api.httpClient.Do(req)
	if err != nil {
		return dtoResp, errors.Trace(err)
	}
	defer resp.Body.Close()

	dtoResp.StatusCode = resp.StatusCode

	bodyRespBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return dtoResp, errors.Trace(err)
	}

	api.log(ctx).Info(resp.StatusCode)
	api.log(ctx).Info(string(bodyRespBytes))

	resp.Body = ioutil.NopCloser(bytes.NewBuffer(bodyRespBytes))

	err = json.NewDecoder(resp.Body).Decode(&dtoResp)

	if err != nil {
		return dtoResp, errors.Trace(err)
	}

	return dtoResp, nil
}

func (api *API) postToAPIWithRetry(ctx context.Context, path string, bodyReqPayload []byte) (*dto.ApiResponse, error) {
	var dtoResp dto.ApiResponse
	var err error

	retryOpts := api.retryOptions(ctx)
	err = retry.Do(func() error {
		dtoResp, err = api.postToAPI(ctx, path, bodyReqPayload)
		if dtoResp.StatusCode == http.StatusUnauthorized {
			return ErrUnauthorized
		}
		if err != nil {
			return errors.Trace(err)
		}
		return nil
	}, retryOpts...)

	if err != nil {
		return nil, errors.Trace(err)
	}

	return &dtoResp, nil
}

func (api *API) retryDecision(ctx context.Context) func(err error) bool {
	return func(err error) bool {
		return err == ErrUnauthorized
	}
}

func (api *API) retryOptions(ctx context.Context) []retry.Option {
	return []retry.Option{
		retry.Attempts(2),
		retry.RetryIf(api.retryDecision(ctx)),
		retry.OnRetry(func(n uint, err error) {
			api.log(ctx).Infof("[Retry] === START AUTH === [Attempts: %d Err: %+v]", n, err)
			api.doAuthentication(ctx)
			api.log(ctx).Infof("[Retry] === END AUTH ===")
		}),
	}
}

// === misc func ===
func (api *API) log(ctx context.Context) *zap.SugaredLogger {
	return logger.Logger(bniCtx.WithBNISessID(ctx, api.bniSessID))
}

func buildURL(baseUrl, paths string, query url.Values) (string, error) {
	u, err := url.Parse(baseUrl)
	if err != nil {
		return "", errors.Trace(err)
	}

	u.Path = path.Join(u.Path, paths)
	u.RawQuery = query.Encode()

	return u.String(), nil
}
