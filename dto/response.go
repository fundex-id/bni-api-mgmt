package dto

import (
	"encoding/json"
	"errors"
)

// === AUTH resp ===
type GetTokenResponse struct {
	AccessToken string `json:"access_token,omitempty"`
	TokenType   string `json:"token_type,omitempty"`
	ExpiredIn   int64  `json:"expired_in,omitempty"`
	Scope       string `json:"scope,omitempty"`
}

// === API resp ====

type ApiResponse struct {
	GetBalanceResponse          *GetBalanceResponse          `json:"getBalanceResponse,omitempty"`
	GetInHouseInquiryResponse   *GetInHouseInquiryResponse   `json:"getInHouseInquiryResponse,omitempty"`
	DoPaymentResponse           *DoPaymentResponse           `json:"doPaymentResponse,omitempty"`
	GetPaymentStatusResponse    *GetPaymentStatusResponse    `json:"getPaymentStatusResponse,omitempty"`
	GetInterBankInquiryResponse *GetInterBankInquiryResponse `json:"getInterBankInquiryResponse,omitempty"`

	BadRespResponse             *BadRespResponse             `json:"Response,omitempty"`
	BadRespGeneralErrorResponse *BadRespGeneralErrorResponse `json:"General Error Response,omitempty"`
}

type GetBalanceResponse struct {
	CommonResponse
	Parameters GetBalanceResponseParam `json:"parameters,omitempty"`
}

type GetBalanceResponseParam struct {
	CommonResponseParam
	CustomerName    string `json:"customerName,omitempty"`
	AccountCurrency string `json:"accountCurrency,omitempty"`
	AccountBalance  int64  `json:"accountBalance,omitempty"`
}

type GetInHouseInquiryResponse struct {
	CommonResponse
	Parameters GetInHouseInquiryResponseParam `json:"parameters,omitempty"`
}

type GetInHouseInquiryResponseParam struct {
	CommonResponseParam
	CustomerName    string `json:"customerName,omitempty"`
	AccountCurrency string `json:"accountCurrency,omitempty"`
	AccountNumber   string `json:"accountNumber,omitempty"`
	AccountStatus   string `json:"accountStatus,omitempty"`
	AccountType     string `json:"accountType,omitempty"`
}

type DoPaymentResponse struct {
	CommonResponse
	Parameters DoPaymentResponseParam `json:"parameters,omitempty"`
}

type DoPaymentResponseParam struct {
	CommonResponseParam
	DebitAccountNo    int64       `json:"debitAccountNo,omitempty"`
	CreditAccountNo   int64       `json:"creditAccountNo,omitempty"`
	ValueAmount       int64       `json:"valueAmount,omitempty"`
	ValueCurrency     string      `json:"valueCurrency,omitempty"`
	BankReference     int64       `json:"bankReference,omitempty"`
	CustomerReference json.Number `json:"customerReference,omitempty"`
}

type GetPaymentStatusResponse struct {
	CommonResponse
	Parameters GetPaymentStatusResponseParam `json:"parameters,omitempty"`
}

type GetPaymentStatusResponseParamPreviousResponse struct {
	TransactionStatus         string `json:"transactionStatus,omitempty"`
	PreviousResponseCode      string `json:"previousResponseCode,omitempty"`
	PreviousResponseMessage   string `json:"previousResponseMessage,omitempty"`
	PreviousResponseTimestamp string `json:"previousResponseTimestamp,omitempty"`

	DebitAccountNo  int64  `json:"debitAccountNo,omitempty"`
	CreditAccountNo int64  `json:"creditAccountNo,omitempty"`
	ValueAmount     int64  `json:"valueAmount,omitempty"`
	ValueCurrency   string `json:"valueCurrency,omitempty"`
}

type GetPaymentStatusResponseParam struct {
	CommonResponseParam

	PreviousResponse GetPaymentStatusResponseParamPreviousResponse `json:"previousResponse,omitempty"`

	BankReference     int64       `json:"bankReference,omitempty"`
	CustomerReference json.Number `json:"customerReference,omitempty"`
}

type GetInterBankInquiryResponse struct {
	CommonResponse
	Parameters GetInterBankInquiryResponseParam `json:"parameters,omitempty"`
}

type GetInterBankInquiryResponseParam struct {
	CommonResponseParam
	DestinationAccountNum  string      `json:"destinationAccountNum,omitempty"`
	DestinationAccountName string      `json:"destinationAccountName,omitempty"`
	DestinationBankName    string      `json:"destinationBankName,omitempty"`
	RetrievalReffNum       json.Number `json:"retrievalReffNum,omitempty"`
}

// === BAD resp ===

type BadRespResponse struct {
	CommonResponse
	Parameters CommonResponseParam `json:"parameters,omitempty"`
}

type BadRespGeneralErrorResponse struct {
	CommonResponse
	Parameters CommonResponseParam `json:"parameters,omitempty"`
}

// === COMMON resp ===

type CommonResponse struct {
	ClientID string `json:"clientId,omitempty"`
	// Parameters        interface{}
	BankReference     string
	CustomerReference string
}

type CommonResponseParam struct {
	ResponseCode      string `json:"responseCode,omitempty"`
	ResponseMessage   string `json:"responseMessage,omitempty"`
	ErrorMessage      string `json:"errorMessage,omitempty"`
	ResponseTimestamp string `json:"responseTimestamp,omitempty"`
}

type ParentResponse map[string]interface{}

func GetCommonResponse(parentResp ParentResponse, keyResp string) (*CommonResponse, error) {
	// https://stackoverflow.com/questions/53486878/http-json-body-response-to-map
	// err := json.NewDecoder(httpResponse.Body).Decode(&data)

	commonResp, exist := parentResp[keyResp]
	if !exist {
		commonResp, exist = parentResp["Response"]
		if !exist {
			commonResp, exist = parentResp["General Error Response"]
			if !exist {
				return nil, errors.New("can't be parsed as CommonResponse")
			}
		}
	}

	resp, ok := commonResp.(CommonResponse)
	if !ok {
		return nil, errors.New("failed to cast as CommonResponse")
	}

	return &resp, nil
}
