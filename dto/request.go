package dto

type CommonRequest struct {
	ClientID  string `json:"clientId,omitempty"`
	Signature string `json:"signature,omitempty"`
}

type GetBalanceRequest struct {
	CommonRequest
	AccountNo string `json:"accountNo,omitempty"`
}
