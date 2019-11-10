package bni

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/fundex-id/bni-api-mgmt/config"
	bniCtx "github.com/fundex-id/bni-api-mgmt/context"
	"github.com/fundex-id/bni-api-mgmt/dto"
	"github.com/fundex-id/bni-api-mgmt/util"
	"github.com/lithammer/shortuuid"
	"github.com/stretchr/testify/assert"
)

var testLogPath string = "test.log"
var dummySignatureConfig config.SignatureConfig = config.SignatureConfig{
	PrivateKeyPath: "testdata/id_rsa.pem",
}

func TestBNI_DoAuthentication(t *testing.T) {
	t.Run("good case", func(t *testing.T) {
		givenConfig := config.Config{
			AuthPath: "/oauth",
			Username: "dummyusername",
			Password: "dummypassword",
			LogPath:  testLogPath,
		}

		testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			assert.Equal(t, http.MethodPost, req.Method)
			assert.Equal(t, givenConfig.AuthPath, req.URL.Path)

			assert.Equal(t, "application/x-www-form-urlencoded", req.Header.Get("content-type"))
			assert.Equal(t, "Basic "+basicAuth(givenConfig.Username, givenConfig.Password), req.Header.Get("authorization"))

			err := req.ParseForm()
			util.AssertErrNil(t, err)

			assert.Equal(t, "client_credentials", req.Form.Get("grant_type"))

			var dtoResp dto.GetTokenResponse
			getJSON("testdata/get_token_response.json", &dtoResp)

			w.WriteHeader(http.StatusOK)
			err = json.NewEncoder(w).Encode(dtoResp)
			util.AssertErrNil(t, err)

		}))
		defer testServer.Close()

		givenConfig.BNIServer = testServer.URL

		bni := New(givenConfig)
		bni.api.httpClient = testServer.Client()

		ctx := bniCtx.WithHTTPReqID(context.Background(), shortuuid.New())
		dtoResp, err := bni.DoAuthentication(ctx)

		assert.NotEmpty(t, dtoResp)
		if util.AssertErrNil(t, err) {
			assert.NotEmpty(t, bni.api.accessToken)
			assert.NotEmpty(t, bni.bniSessID)
		}

	})

	t.Run("bad auth", func(t *testing.T) {
		givenConfig := config.Config{
			AuthPath: "/oauth",
			Username: "dummyusername",
			Password: "dummypassword",
			LogPath:  testLogPath,
		}

		testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			assert.Equal(t, http.MethodPost, req.Method)
			assert.Equal(t, givenConfig.AuthPath, req.URL.Path)

			w.WriteHeader(http.StatusUnauthorized)
		}))
		defer testServer.Close()

		givenConfig.BNIServer = testServer.URL

		bni := New(givenConfig)
		bni.api.httpClient = testServer.Client()

		ctx := bniCtx.WithHTTPReqID(context.Background(), shortuuid.New())
		dtoResp, err := bni.DoAuthentication(ctx)

		assert.Nil(t, dtoResp)
		if util.AssertErrNotNil(t, err) {
			assert.Empty(t, bni.api.accessToken)
			assert.Empty(t, bni.bniSessID)
		}
	})
}

func TestBNI_GetBalance(t *testing.T) {
	t.Run("good case", func(t *testing.T) {
		givenConfig := config.Config{
			AuthPath:        "/oauth",
			BalancePath:     "/H2H/getbalance",
			Username:        "dummyusername",
			Password:        "dummypassword",
			LogPath:         testLogPath,
			SignatureConfig: dummySignatureConfig,
		}

		testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			assert.Equal(t, http.MethodPost, req.Method)
			assert.Equal(t, givenConfig.BalancePath, req.URL.Path)

			assert.Equal(t, "application/json", req.Header.Get("content-type"))

			var dtoResp dto.ApiResponse
			getJSON("testdata/get_balance_response.json", &dtoResp)

			w.WriteHeader(http.StatusOK)
			err := json.NewEncoder(w).Encode(dtoResp)
			util.AssertErrNil(t, err)

		}))
		defer testServer.Close()

		givenConfig.BNIServer = testServer.URL

		bni := New(givenConfig)
		bni.api.retryablehttpClient.HTTPClient = testServer.Client()

		dtoReq := dto.GetBalanceRequest{AccountNo: "115471119"}

		ctx := bniCtx.WithHTTPReqID(context.Background(), shortuuid.New())
		dtoResp, err := bni.GetBalance(ctx, &dtoReq)

		util.AssertErrNil(t, err)
		assert.NotEmpty(t, dtoResp)
	})

	t.Run("bad response", func(t *testing.T) {
		givenConfig := config.Config{
			AuthPath:        "/oauth",
			BalancePath:     "/H2H/getbalance",
			Username:        "dummyusername",
			Password:        "dummypassword",
			LogPath:         testLogPath,
			SignatureConfig: dummySignatureConfig,
		}

		testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			assert.Equal(t, http.MethodPost, req.Method)
			assert.Equal(t, givenConfig.BalancePath, req.URL.Path)

			assert.Equal(t, "application/json", req.Header.Get("content-type"))

			var dtoResp dto.ApiResponse
			getJSON("testdata/bad_response.json", &dtoResp)

			w.WriteHeader(http.StatusOK)
			err := json.NewEncoder(w).Encode(dtoResp)
			util.AssertErrNil(t, err)

		}))
		defer testServer.Close()

		givenConfig.BNIServer = testServer.URL

		bni := New(givenConfig)
		bni.api.retryablehttpClient.HTTPClient = testServer.Client()

		dtoReq := dto.GetBalanceRequest{AccountNo: "115471119"}

		ctx := bniCtx.WithHTTPReqID(context.Background(), shortuuid.New())
		dtoResp, err := bni.GetBalance(ctx, &dtoReq)

		util.AssertErrNotNil(t, err)
		assert.Empty(t, dtoResp)
	})

	t.Run("no auth then good response", func(t *testing.T) {
		givenConfig := config.Config{
			AuthPath:        "/oauth",
			BalancePath:     "/H2H/getbalance",
			Username:        "dummyusername",
			Password:        "dummypassword",
			LogPath:         testLogPath,
			SignatureConfig: dummySignatureConfig,
		}

		var hits uint64

		testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			assert.Equal(t, http.MethodPost, req.Method)

			if hits == 0 {
				assert.Equal(t, givenConfig.BalancePath, req.URL.Path)
				w.WriteHeader(http.StatusUnauthorized)
				hits++
				return
			}

			if hits == 1 {
				assert.Equal(t, givenConfig.AuthPath, req.URL.Path)

				var dtoResp dto.GetTokenResponse
				getJSON("testdata/get_token_response.json", &dtoResp)

				w.WriteHeader(http.StatusOK)
				err := json.NewEncoder(w).Encode(dtoResp)
				util.AssertErrNil(t, err)

				hits++
				return
			}

			if hits == 2 {
				var dtoResp dto.ApiResponse
				getJSON("testdata/get_balance_response.json", &dtoResp)

				w.WriteHeader(http.StatusOK)
				err := json.NewEncoder(w).Encode(dtoResp)
				util.AssertErrNil(t, err)
				hits = 0
			}

		}))
		defer testServer.Close()

		givenConfig.BNIServer = testServer.URL

		bni := New(givenConfig)
		bni.api.retryablehttpClient.HTTPClient = testServer.Client()

		dtoReq := dto.GetBalanceRequest{AccountNo: "115471119"}

		firstReqID := shortuuid.New()
		ctx := bniCtx.WithHTTPReqID(context.Background(), firstReqID)
		dtoResp, err := bni.GetBalance(ctx, &dtoReq)

		util.AssertErrNil(t, err)
		assert.NotEmpty(t, dtoResp)

		dtoReq = dto.GetBalanceRequest{AccountNo: "225471120"}

		secondReqID := shortuuid.New()
		ctx = bniCtx.WithHTTPReqID(context.Background(), secondReqID)
		dtoResp, err = bni.GetBalance(ctx, &dtoReq)

		util.AssertErrNil(t, err)
		assert.NotEmpty(t, dtoResp)

		t.Logf("firstReq: %s secondReq: %s", firstReqID, secondReqID)
	})
}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func getJSON(filePath string, obj interface{}) {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		log.Fatal(err)
	}

	file, err := os.Open(absPath)
	defer file.Close()

	if err != nil {
		log.Fatal(err)
	}

	byteValue, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(byteValue, &obj)
	if err != nil {
		log.Fatal(err)
	}

}
