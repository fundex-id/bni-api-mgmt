package dto

type CommonRequest struct {
	ClientID  string `json:"clientId,omitempty"`
	Signature string `json:"signature,omitempty"`
}

type GetBalanceRequest struct {
	CommonRequest
	AccountNo string `json:"accountNo,omitempty"`
}

type GetInHouseInquiryRequest struct {
	CommonRequest
	AccountNo string `json:"accountNo,omitempty"`
}

type DoPaymentRequest struct {
	CommonRequest
	CustomerReferenceNumber string
	PaymentMethod           string
	DebitAccountNo          string
	CreditAccountNo         string
	ValueDate               string
	ValueCurrency           string
	ValueAmount             string
	Remark                  string
	BeneficiaryEmailAddress string
	DestinationBankCode     string
	BeneficiaryName         string
	BeneficiaryAddress1     string
	BeneficiaryAddress2     string
	ChargingModelId         string
}

type GetPaymentStatusRequest struct {
	CommonRequest
	CustomerReferenceNumber string
}
