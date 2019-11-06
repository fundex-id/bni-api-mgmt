package bni

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_buildURL(t *testing.T) {
	baseUrl := "https://apidev.bni.co.id:8065"

	type args struct {
		path  string
		query url.Values
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{name: "auth", args: args{path: "/api/oauth/token"},
			want: baseUrl + "/api/oauth/token", wantErr: false},
		{name: "with query", args: args{path: "/H2H/getbalance", query: url.Values{"access_token": []string{"12345"}}},
			want: baseUrl + "/H2H/getbalance?access_token=12345", wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildURL(baseUrl, tt.args.path, tt.args.query)

			assert.Equal(t, tt.wantErr, (err != nil))
			assert.Equal(t, tt.want, got)
		})
	}
}
