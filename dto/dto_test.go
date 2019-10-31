package dto

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/davecgh/go-spew/spew"
)

func TestGetCommonResponse(t *testing.T) {
	t.Run("get_balance_response.json", func(t *testing.T) {

		file, err := os.Open("./testdata/get_balance_response.json")
		defer file.Close()

		if err != nil {
			log.Fatal(err)
		}

		byteValue, err := ioutil.ReadAll(file)
		if err != nil {
			log.Fatal(err)
		}

		// var parentResp ParentResponse

		// err = json.Unmarshal(byteValue, &parentResp)
		// if err != nil {
		// 	log.Fatal(err)
		// }

		// commonResp, err := GetCommonResponse(parentResp, "getBalanceResponse")

		// t.Logf("COMMONRESP: %+v", commonResp)
		// t.Logf("error: %+v", err)

		var apiResponse ApiResponse

		err = json.Unmarshal(byteValue, &apiResponse)
		if err != nil {
			log.Fatal(err)
		}

		// commonResp, err := GetCommonResponse(apiResponse, "getBalanceResponse")

		t.Logf("APIRESPONSE: %+v", spew.Sdump(apiResponse))
		t.Logf("error: %+v", err)

	})
	// type args struct {
	// 	parentResp ParentResponse
	// 	keyResp    string
	// }
	// tests := []struct {
	// 	name    string
	// 	args    args
	// 	want    *CommonResponse
	// 	wantErr bool
	// }{
	// 	// TODO: Add test cases.
	// }
	// for _, tt := range tests {
	// 	t.Run(tt.name, func(t *testing.T) {
	// 		got, err := GetCommonResponse(tt.args.parentResp, tt.args.keyResp)
	// 		if (err != nil) != tt.wantErr {
	// 			t.Errorf("GetCommonResponse() error = %v, wantErr %v", err, tt.wantErr)
	// 			return
	// 		}
	// 		if !reflect.DeepEqual(got, tt.want) {
	// 			t.Errorf("GetCommonResponse() = %v, want %v", got, tt.want)
	// 		}
	// 	})
	// }
}
