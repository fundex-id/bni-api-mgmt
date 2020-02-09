package config

type Config struct {
	Username  string
	Password  string
	ClientID  string
	BNIServer string
	LogPath   string
	SignatureConfig
}

type SignatureConfig struct {
	PrivateKeyPath string
}
