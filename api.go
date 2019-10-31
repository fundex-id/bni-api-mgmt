package bni

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/fundex-id/bni-api-mgmt/dto"
	"github.com/juju/errors"
)

type Api struct {
	config     Config
	httpClient *http.Client
}

func NewApi(config Config) *Api {
	httpClient := &http.Client{
		Timeout: time.Second * 10,
	}

	api := Api{config: config, httpClient: httpClient}

	return &api
}

func (api *Api) postGetToken() (*dto.GetTokenResponse, error) {
	url, err := joinUrl(api.config.BNIServer, api.config.AuthPath)
	if err != nil {
		return nil, err
	}

	body := strings.NewReader("grant_type=client_credentials")

	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return nil, errors.Trace(err)
	}

	req.Header.Add("content-type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(api.config.Username, api.config.Password)

	res, err := api.httpClient.Do(req)
	if err != nil {
		return nil, errors.Trace(err)
	}
	defer res.Body.Close()

	body, _ := ioutil.ReadAll(res.Body)

	fmt.Println(res)
	fmt.Println(string(body))

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
