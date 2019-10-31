package bni

type Config struct {
	Username              string
	Password              string
	ClientID              string
	BNIServer             string
	AuthPath              string
	InHouseInquiryPath    string
	InterBankInquiryPath  string
	InHouseTransferPath   string
	InterBankTransferPath string
	LogPath               string
}
