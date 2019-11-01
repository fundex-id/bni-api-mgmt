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
	"time"

	"log"

	"github.com/fundex-id/bni-api-mgmt/dto"
	"github.com/fundex-id/bni-api-mgmt/logger"
	"github.com/juju/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Api struct {
	config     Config
	httpClient *http.Client
}

func newApi(config Config) *Api {

	httpClient := &http.Client{
		Timeout: time.Second * 10,
	}

	api := Api{config: config,
		httpClient: httpClient,
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

func (api *Api) postGetToken(ctx context.Context) (*dto.GetTokenResponse, error) {
	apiLog := logger.Logger(ctx)
	log.Print("HELLO WORLD dari LOG")
	apiLog.Info("HELLO WORLD dari APILOG")

	url, err := joinUrl(api.config.BNIServer, api.config.AuthPath)
	if err != nil {
		return nil, err
	}

	bodyReq := strings.NewReader("grant_type=client_credentials")

	req, err := http.NewRequest(http.MethodPost, url, bodyReq)
	if err != nil {
		return nil, errors.Trace(err)
	}

	req.Header.Add("content-type", "application/x-www-form-urlencoded")
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

	apiLog.Info(resp.StatusCode)
	apiLog.Info(string(bodyRespBytes))

	resp.Body = ioutil.NopCloser(bytes.NewBuffer(bodyRespBytes))

	// fmt.Println(resp)
	// fmt.Println(string(bodyRespBytes))

	// Step 3
	// oR := new(jsonResponse)
	// json.NewDecoder(resp.Body).Decode(oR)

	var jsonResp dto.GetTokenResponse
	err = json.NewDecoder(resp.Body).Decode(&jsonResp)
	if err != nil {
		return nil, errors.Trace(err)
	}

	return &jsonResp, nil
}

// func (api *Api) sendInHouseTransferRequest(accessToken string, inHouseTransferRequet InHouseTransferRequest) error {

// }

func joinUrl(baseUrl, paths string) (string, error) {
	u, err := url.Parse(baseUrl)
	if err != nil {
		return "", errors.Trace(err)
	}

	u.Path = path.Join(u.Path, paths)

	return u.String(), nil
}
