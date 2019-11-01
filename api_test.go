package bni

import (
	"context"
	"net/http"
	"testing"

	ctxApp "github.com/fundex-id/bni-api-mgmt/context"
	"github.com/juju/errors"
	"github.com/pborman/uuid"

	"github.com/stretchr/testify/assert"
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

func TestApi_postGetToken(t *testing.T) {
	api := &Api{
		config:     Config{BNIServer: "https://bni.com:8181", AuthPath: "/oauth"},
		httpClient: http.DefaultClient,
	}

	reqId := uuid.NewRandom()
	ctx := ctxApp.WithReqId(context.Background(), reqId.String())
	_, err := api.postGetToken(ctx)
	assert.Nil(t, err)

	// t.Log(err)
	t.Log(errors.ErrorStack(err))

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
}
