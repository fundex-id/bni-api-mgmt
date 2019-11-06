package bni

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/fundex-id/bni-api-mgmt/dto"
	"github.com/fundex-id/bni-api-mgmt/logger"
	"github.com/hashicorp/go-cleanhttp"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/juju/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type API struct {
	config              Config
	httpClient          *http.Client // for postGetToken only
	retryablehttpClient *retryablehttp.Client
}

func newApi(config Config) *API {

	httpClient := cleanhttp.DefaultPooledClient()
	retryablehttpClient := retryablehttp.NewClient()

	api := API{config: config,
		httpClient:          httpClient,
		retryablehttpClient: retryablehttpClient,
	}

	logger.SetOptions(zap.WrapCore(func(core zapcore.Core) zapcore.Core {

		fileWriteSyncer := zapcore.AddSync(&lumberjack.Logger{
			Filename: api.config.LogPath,
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

	return &api
}

func (api *API) postGetToken(ctx context.Context) (*dto.GetTokenResponse, error) {
	funcLog := logger.Logger(ctx)

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

	funcLog.Info(resp.StatusCode)
	funcLog.Info(string(bodyRespBytes))

	resp.Body = ioutil.NopCloser(bytes.NewBuffer(bodyRespBytes))

	var jsonResp dto.GetTokenResponse
	err = json.NewDecoder(resp.Body).Decode(&jsonResp)

	if err != nil {
		return nil, errors.Trace(err)
	}

	return &jsonResp, nil
}

// func (api *Api) sendInHouseTransferRequest(accessToken string, inHouseTransferRequet InHouseTransferRequest) error {

// }

func buildURL(baseUrl, paths string, query url.Values) (string, error) {
	u, err := url.Parse(baseUrl)
	if err != nil {
		return "", errors.Trace(err)
	}

	u.Path = path.Join(u.Path, paths)
	u.RawQuery = query.Encode()

	return u.String(), nil
}
