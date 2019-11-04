package bni

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	dummyBNIServer string = "https://dummy.com:8181"
)

func Test_joinUrl(t *testing.T) {
	baseUrl := "https://apidev.bni.co.id:8065"

	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{name: "auth", args: args{path: "/api/oauth/token"}, want: baseUrl + "/api/oauth/token", wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := joinUrl(baseUrl, tt.args.path)

			assert.Equal(t, (err != nil), tt.wantErr)
			assert.Equal(t, got, tt.want)
		})
	}
}

func TestApi_postGetToken_it(t *testing.T) {
	// api := &API{
	// 	config: Config{
	// 		BNIServer: "https://bni.com:8181",
	// 		AuthPath:  "/oauth",
	// 		LogPath:   "custom.log",
	// 	},
	// 	httpClient: http.DefaultClient,
	// }

	// reqId := uuid.NewRandom()
	// ctx := ctxApp.WithReqId(context.Background(), reqId.String())
	// _, err := api.postGetToken(ctx)
	// assert.Nil(t, err)

	// // t.Log(err)
	// t.Log(errors.ErrorStack(err))

	// type fields struct {
	// 	config     Config
	// 	httpClient *http.Client
	// 	logger     *logrus.Logger
	// }
	// type args struct {
	// 	ctx context.Context
	// }
	// tests := []struct {
	// 	name    string
	// 	fields  fields
	// 	args    args
	// 	want    *dto.GetTokenResponse
	// 	wantErr bool
	// }{
	// 	// TODO: Add test cases.
	// }
	// for _, tt := range tests {
	// 	t.Run(tt.name, func(t *testing.T) {
	// 		api := &Api{
	// 			config:     tt.fields.config,
	// 			httpClient: tt.fields.httpClient,
	// 			logger:     tt.fields.logger,
	// 		}
	// 		got, err := api.postGetToken(tt.args.ctx)
	// 		if (err != nil) != tt.wantErr {
	// 			t.Errorf("Api.postGetToken() error = %v, wantErr %v", err, tt.wantErr)
	// 			return
	// 		}
	// 		if !reflect.DeepEqual(got, tt.want) {
	// 			t.Errorf("Api.postGetToken() = %v, want %v", got, tt.want)
	// 		}
	// 	})
	// }

	t.Run("good case", func(t *testing.T) {
		givenConfig := Config{
			BNIServer: dummyBNIServer,
			AuthPath:  "/oauth",
			Username:  "dummyusername",
			Password:  "dummypassword",
		}

		t.Log("MASUK TEST")

		testHandler := func(w http.ResponseWriter, req *http.Request) {
			// url, err := joinUrl(givenConfig.BNIServer, givenConfig.AuthPath)
			// assert.Nil(t, err)

			assert.Equal(t, req.Method, http.MethodPost)
			assert.Equal(t, req.URL.String(), givenConfig.AuthPath)

			// t.Log("RECEIVED URL: ", req.URL)

			assert.Equal(t, req.Header.Get("content-type"), "application/x-www-form-urlencoded")
			assert.Equal(t, req.Header.Get("authorization"), "Basic "+basicAuth(givenConfig.Username, givenConfig.Password))

			w.WriteHeader(http.StatusBadGateway)
			t.Log("MASUK DUMMY SERVER")
		}

		testServer := httptest.NewServer(http.HandlerFunc(testHandler))
		defer testServer.Close()

		givenConfig.BNIServer = testServer.URL

		api := newApi(givenConfig)
		api.httpClient = testServer.Client()
		t.Logf("SERV URL: %s", testServer.URL)

		dtoResp, err := api.postGetToken(context.Background())
		// if err != nil {
		// 	t.Errorf("expecte nil, got: %+v", spew.Sdump(err))
		// }
		assertErrNil(t, err)
		// assert.Nil(t, err)
		assert.NotNil(t, dtoResp)
		// testAssertNil(t, dtoResp)
	})

}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func assertErrNil(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Errorf("expect nil, but got: %+v", err)
	}
}

func assertErrNotNil(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Errorf("expect not nil, but got: %+v", err)
	}
}
