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
	CustomerReferenceNumber string `json:"customerReferenceNumber,omitempty"`
	PaymentMethod           string `json:"paymentMethod,omitempty"`
	DebitAccountNo          string `json:"debitAccountNo,omitempty"`
	CreditAccountNo         string `json:"creditAccountNo,omitempty"`
	ValueDate               string `json:"valueDate,omitempty"`
	ValueCurrency           string `json:"valueCurrency,omitempty"`
	ValueAmount             string `json:"valueAmount,omitempty"`
	Remark                  string `json:"remark,omitempty"`
	BeneficiaryEmailAddress string `json:"beneficiaryEmailAddress,omitempty"`
	DestinationBankCode     string `json:"destinationBankCode,omitempty"`
	BeneficiaryName         string `json:"beneficiaryName,omitempty"`
	BeneficiaryAddress1     string `json:"beneficiaryAddress1,omitempty"`
	BeneficiaryAddress2     string `json:"beneficiaryAddress2,omitempty"`
	ChargingModelId         string `json:"chargingModelId,omitempty"`
}

type GetPaymentStatusRequest struct {
	CommonRequest
	CustomerReferenceNumber string `json:"customerReferenceNumber,omitempty"`
}

type GetInterBankInquiryRequest struct {
	CommonRequest
	CustomerReferenceNumber string `json:"customerReferenceNumber,omitempty"`
	AccountNum              string `json:"accountNum,omitempty"`
	DestinationBankCode     string `json:"destinationBankCode,omitempty"`
	DestinationAccountNum   string `json:"destinationAccountNum,omitempty"`
}

type GetInterBankPaymentRequest struct {
	CommonRequest
	CustomerReferenceNumber string `json:"customerReferenceNumber,omitempty"`
	Amount                  string `json:"amount,omitempty"`
	DestinationAccountNum   string `json:"destinationAccountNum,omitempty"`
	DestinationAccountName  string `json:"destinationAccountName,omitempty"`
	DestinationBankCode     string `json:"destinationBankCode,omitempty"`
	DestinationBankName     string `json:"destinationBankName,omitempty"`
	AccountNum              string `json:"accountNum,omitempty"`
	RetrievalReffNum        string `json:"retrievalReffNum,omitempty"`
}
