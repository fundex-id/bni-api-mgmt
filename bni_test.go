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

	bniCtx "github.com/fundex-id/bni-api-mgmt/context"
	"github.com/fundex-id/bni-api-mgmt/dto"
	"github.com/fundex-id/bni-api-mgmt/util"
	"github.com/pborman/uuid"
	"github.com/stretchr/testify/assert"
)

func TestBNI_DoAuthentication(t *testing.T) {
	t.Run("good case", func(t *testing.T) {
		givenConfig := Config{
			AuthPath: "/oauth",
			Username: "dummyusername",
			Password: "dummypassword",
		}

		testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			assert.Equal(t, req.Method, http.MethodPost)
			assert.Equal(t, req.URL.String(), givenConfig.AuthPath)

			assert.Equal(t, req.Header.Get("content-type"), "application/x-www-form-urlencoded")
			assert.Equal(t, req.Header.Get("authorization"), "Basic "+basicAuth(givenConfig.Username, givenConfig.Password))

			err := req.ParseForm()
			util.AssertErrNil(t, err)

			assert.Equal(t, req.Form.Get("grant_type"), "client_credentials")

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

		ctx := bniCtx.WithReqId(context.Background(), uuid.NewRandom().String())
		dtoResp, err := bni.DoAuthentication(ctx)
		util.AssertErrNil(t, err)
		if assert.NotNil(t, dtoResp) {
			assert.NotEmpty(t, dtoResp.Session.ID)
		}

	})

	t.Run("bad auth", func(t *testing.T) {
		givenConfig := Config{
			AuthPath: "/oauth",
			Username: "dummyusername",
			Password: "dummypassword",
		}

		testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			assert.Equal(t, req.Method, http.MethodPost)
			assert.Equal(t, req.URL.String(), givenConfig.AuthPath)

			w.WriteHeader(http.StatusUnauthorized)
		}))
		defer testServer.Close()

		givenConfig.BNIServer = testServer.URL

		bni := New(givenConfig)
		bni.api.httpClient = testServer.Client()

		ctx := bniCtx.WithReqId(context.Background(), uuid.NewRandom().String())
		dtoResp, err := bni.DoAuthentication(ctx)
		util.AssertErrNotNil(t, err)
		assert.Nil(t, dtoResp)
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
