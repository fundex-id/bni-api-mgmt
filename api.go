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

	"github.com/fundex-id/bni-api-mgmt/config"
	bniCtx "github.com/fundex-id/bni-api-mgmt/context"
	"github.com/fundex-id/bni-api-mgmt/dto"
	"github.com/fundex-id/bni-api-mgmt/logger"
	"github.com/hashicorp/go-cleanhttp"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/juju/errors"
	"go.uber.org/zap"
)

type API struct {
	config              config.Config
	httpClient          *http.Client // for postGetToken only
	retryablehttpClient *retryablehttp.Client

	mutex       sync.Mutex
	accessToken string
	bniSessID   string
}

func newApi(config config.Config) *API {

	httpClient := cleanhttp.DefaultPooledClient()
	retryablehttpClient := retryablehttp.NewClient()

	api := API{config: config,
		httpClient:          httpClient,
		retryablehttpClient: retryablehttpClient,
	}

	return &api
}

func (api *API) setAccessTokenAndSessID(accessToken, bniSessID string) {
	api.mutex.Lock()
	defer api.mutex.Unlock()

	api.accessToken = accessToken
	api.bniSessID = bniSessID
}

func (api *API) postGetToken(ctx context.Context) (*dto.GetTokenResponse, error) {
	urlTarget, err := buildURL(api.config.BNIServer, api.config.AuthPath, url.Values{})
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

func (api *API) postGetBalance(ctx context.Context, dtoReq *dto.GetBalanceRequest) (*dto.ApiResponse, error) {
	jsonReq, err := json.Marshal(dtoReq)
	if err != nil {
		return nil, errors.Trace(err)
	}

	return api.postToAPI(ctx, api.config.BalancePath, jsonReq)
}

func (api *API) postGetInHouseInquiry(ctx context.Context, dtoReq *dto.GetInHouseInquiryRequest) (*dto.ApiResponse, error) {
	jsonReq, err := json.Marshal(dtoReq)
	if err != nil {
		return nil, errors.Trace(err)
	}

	return api.postToAPI(ctx, api.config.InHouseInquiryPath, jsonReq)
}

func (api *API) postDoPayment(ctx context.Context, dtoReq *dto.DoPaymentRequest) (*dto.ApiResponse, error) {
	jsonReq, err := json.Marshal(dtoReq)
	if err != nil {
		return nil, errors.Trace(err)
	}

	return api.postToAPI(ctx, api.config.InHouseTransferPath, jsonReq)
}

// Generic POST request to API
func (api *API) postToAPI(ctx context.Context, path string, bodyReqPayload []byte) (*dto.ApiResponse, error) {
	urlQuery := url.Values{"access_token": []string{api.accessToken}}
	urlTarget, err := buildURL(api.config.BNIServer, path, urlQuery)
	if err != nil {
		return nil, err
	}

	req, err := retryablehttp.NewRequest(http.MethodPost, urlTarget, bytes.NewBuffer(bodyReqPayload))
	if err != nil {
		return nil, errors.Trace(err)
	}
	req = req.WithContext(ctx)

	req.Header.Set("content-type", "application/json")

	resp, err := api.retryablehttpClient.Do(req)
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

	var jsonResp dto.ApiResponse
	err = json.NewDecoder(resp.Body).Decode(&jsonResp)

	if err != nil {
		return nil, errors.Trace(err)
	}

	return &jsonResp, nil
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
