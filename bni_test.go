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

		dtoReq := dto.GetBalanceRequest{
			AccountNo: "115471119",
		}

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

		dtoReq := dto.GetBalanceRequest{
			AccountNo: "115471119",
		}

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

		dtoReq := dto.GetBalanceRequest{
			AccountNo: "115471119",
		}

		firstReqID := shortuuid.New()
		ctx := bniCtx.WithHTTPReqID(context.Background(), firstReqID)
		dtoResp, err := bni.GetBalance(ctx, &dtoReq)

		util.AssertErrNil(t, err)
		assert.NotEmpty(t, dtoResp)

		dtoReq = dto.GetBalanceRequest{
			AccountNo: "225471120",
		}

		secondReqID := shortuuid.New()
		ctx = bniCtx.WithHTTPReqID(context.Background(), secondReqID)
		dtoResp, err = bni.GetBalance(ctx, &dtoReq)

		util.AssertErrNil(t, err)
		assert.NotEmpty(t, dtoResp)

		t.Logf("firstReq: %s secondReq: %s", firstReqID, secondReqID)
	})
}

func TestBNI_GetInHouseInquiry(t *testing.T) {
	t.Run("good case", func(t *testing.T) {
		givenConfig := config.Config{
			InHouseInquiryPath: "/H2H/getinhouseinquiry",
			LogPath:            testLogPath,
			SignatureConfig:    dummySignatureConfig,
		}

		bni, testServer := buildBNIAndMockServerGoodResponse(t, givenConfig,
			givenConfig.InHouseInquiryPath,
			"testdata/get_inhouseinquiry_response.json",
		)

		dtoReq := dto.GetInHouseInquiryRequest{
			AccountNo: "115471119",
		}

		ctx := bniCtx.WithHTTPReqID(context.Background(), shortuuid.New())
		dtoResp, err := bni.GetInHouseInquiry(ctx, &dtoReq)

		util.AssertErrNil(t, err)
		assert.NotEmpty(t, dtoResp)

		testServer.Close()
	})
}

func TestBNI_DoPayment(t *testing.T) {
	t.Run("good case", func(t *testing.T) {
		givenConfig := config.Config{
			InHouseTransferPath: "/H2H/dopayment",
			LogPath:             testLogPath,
			SignatureConfig:     dummySignatureConfig,
		}

		bni, testServer := buildBNIAndMockServerGoodResponse(t, givenConfig,
			givenConfig.InHouseTransferPath,
			"testdata/get_dopayment_response.json",
		)

		dtoReq := dto.DoPaymentRequest{
			CustomerReferenceNumber: "20170227000000000020",
			PaymentMethod:           "0",
			DebitAccountNo:          "113183203",
			CreditAccountNo:         "115471119",
			ValueDate:               "20170227000000000",
			ValueCurrency:           "IDR",
			ValueAmount:             "100500",
			Remark:                  "?",
			BeneficiaryEmailAddress: "",
			BeneficiaryName:         "Mr.X",
			BeneficiaryAddress1:     "Jakarta",
			BeneficiaryAddress2:     "",
			DestinationBankCode:     "CENAIDJAXXX",
			ChargingModelId:         "NONE",
		}

		ctx := bniCtx.WithHTTPReqID(context.Background(), shortuuid.New())
		dtoResp, err := bni.DoPayment(ctx, &dtoReq)

		util.AssertErrNil(t, err)
		assert.NotEmpty(t, dtoResp)

		testServer.Close()
	})
}

func TestBNI_GetPaymentStatus(t *testing.T) {
	t.Run("good case", func(t *testing.T) {
		givenConfig := config.Config{
			PaymentStatusPath: "/H2H/getpaymentstatus",
			LogPath:           testLogPath,
			SignatureConfig:   dummySignatureConfig,
		}

		bni, testServer := buildBNIAndMockServerGoodResponse(t, givenConfig,
			givenConfig.PaymentStatusPath,
			"testdata/get_getpaymentstatus_response.json",
		)

		dtoReq := dto.GetPaymentStatusRequest{
			CustomerReferenceNumber: "20170227000000000020",
		}

		ctx := bniCtx.WithHTTPReqID(context.Background(), shortuuid.New())
		dtoResp, err := bni.GetPaymentStatus(ctx, &dtoReq)

		util.AssertErrNil(t, err)
		assert.NotEmpty(t, dtoResp)

		testServer.Close()
	})
}

func TestBNI_GetInterBankInquiry(t *testing.T) {
	t.Run("good case", func(t *testing.T) {
		givenConfig := config.Config{
			InterBankInquiryPath: "/H2H/getinterbankinquiry",
			LogPath:              testLogPath,
			SignatureConfig:      dummySignatureConfig,
		}

		bni, testServer := buildBNIAndMockServerGoodResponse(t, givenConfig,
			givenConfig.InterBankInquiryPath,
			"testdata/get_getinterbankinquiry_response.json",
		)

		dtoReq := dto.GetInterBankInquiryRequest{
			CustomerReferenceNumber: "20170227000000000021",
			AccountNum:              "113183203",
			DestinationBankCode:     "014",
			DestinationAccountNum:   "3333333333",
		}

		ctx := bniCtx.WithHTTPReqID(context.Background(), shortuuid.New())
		dtoResp, err := bni.GetInterBankInquiry(ctx, &dtoReq)

		util.AssertErrNil(t, err)
		assert.NotEmpty(t, dtoResp)

		testServer.Close()
	})
}

func TestBNI_GetInterBankPayment(t *testing.T) {
	t.Run("good case", func(t *testing.T) {
		givenConfig := config.Config{
			InterBankTransferPath: "/H2H/getinterbankpayment",
			LogPath:               testLogPath,
			SignatureConfig:       dummySignatureConfig,
		}

		bni, testServer := buildBNIAndMockServerGoodResponse(t, givenConfig,
			givenConfig.InterBankTransferPath,
			"testdata/get_getinterbankpayment_response.json",
		)

		dtoReq := dto.GetInterBankPaymentRequest{
			CustomerReferenceNumber: "20170227000000000021",
			Amount:                  "10000",
			DestinationAccountNum:   "3333333333",
			DestinationAccountName:  "BENEFICIARY NAME 1 2(OPT) UNTIL HERE2",
			DestinationBankCode:     "014",
			DestinationBankName:     "BCA",
			AccountNum:              "115471119",
			RetrievalReffNum:        "100000000024",
		}

		ctx := bniCtx.WithHTTPReqID(context.Background(), shortuuid.New())
		dtoResp, err := bni.GetInterBankPayment(ctx, &dtoReq)

		util.AssertErrNil(t, err)
		assert.NotEmpty(t, dtoResp)

		testServer.Close()
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

	err = file.Close()
	if err != nil {
		log.Fatal(err)
	}
}

func buildBNIAndMockServerGoodResponse(t *testing.T, givenConfig config.Config, assertPath string, jsonPathTestData string) (bni *BNI, testServer *httptest.Server) {
	t.Helper()

	testServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		assert.Equal(t, http.MethodPost, req.Method)
		assert.Equal(t, assertPath, req.URL.Path)

		assert.Equal(t, "application/json", req.Header.Get("content-type"))

		var dtoResp dto.ApiResponse
		getJSON(jsonPathTestData, &dtoResp)

		w.WriteHeader(http.StatusOK)
		err := json.NewEncoder(w).Encode(dtoResp)
		util.AssertErrNil(t, err)

	}))

	givenConfig.BNIServer = testServer.URL

	bni = New(givenConfig)
	bni.api.retryablehttpClient.HTTPClient = testServer.Client()

	return bni, testServer
}
