package dto

import "errors"

type GetTokenResponse struct {
	AccessToken string `json:"access_token,omitempty"`
	TokenType   string `json:"token_type,omitempty"`
	ExpiredIn   int64  `json:"expired_in,omitempty"`
	Scope       string `json:"scope,omitempty"`
}

type ApiResponse struct {
	GetBalanceResponse          *GetBalanceResponse
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

type BadRespResponse struct {
	CommonResponse
	Parameters CommonResponseParam `json:"parameters,omitempty"`
}

type BadRespGeneralErrorResponse struct {
	CommonResponse
	Parameters CommonResponseParam `json:"parameters,omitempty"`
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

type ClientID struct {
	ClientID string `json:"clientId,omitempty"`
}

type CommonResponse struct {
	ClientID
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

type GetInHouseInquiryResponseParam struct {
	CommonResponseParam
}

type BaseRequest struct {
	ClientID  string
	Signature string
}