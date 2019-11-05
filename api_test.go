package bni

import (
	"encoding/base64"
	"testing"

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

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}
