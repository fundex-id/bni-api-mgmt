package config

type Config struct {
	Username              string
	Password              string
	ClientID              string
	BNIServer             string
	AuthPath              string
	BalancePath           string
	InHouseInquiryPath    string
	InterBankInquiryPath  string
	InHouseTransferPath   string
	InterBankTransferPath string
	LogPath               string
	SignatureConfig
}

type SignatureConfig struct {
	PrivateKeyPath string
}
