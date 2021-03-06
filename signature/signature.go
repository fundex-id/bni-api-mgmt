package signature

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"io/ioutil"

	"github.com/fundex-id/bni-api-mgmt/config"
	"github.com/juju/errors"
)

type Signature struct {
	config     config.SignatureConfig
	privateKey *rsa.PrivateKey
}

func New(config config.SignatureConfig) *Signature {
	return &Signature{config: config}
}

func (s *Signature) Sha256WithRSA(data string) (string, error) {

	if s.privateKey == nil {
		privateKey, err := loadPrivateKeyFromPEMFile(s.config.PrivateKeyPath)
		if err != nil {
			return "", errors.Trace(err)
		}
		s.privateKey = privateKey
	}

	h := sha256.New()
	_, err := h.Write([]byte(data))
	if err != nil {
		return "", errors.Trace(err)
	}
	d := h.Sum(nil)

	signature, err := rsa.SignPKCS1v15(rand.Reader, s.privateKey, crypto.SHA256, d)
	if err != nil {
		return "", errors.Trace(err)
	}

	// logger.Infof("Signature in byte: %v\n\n", signature)

	encodedSig := base64.StdEncoding.EncodeToString(signature)

	// logger.Infof("Encoded signature: %v\n\n", encodedSig)

	return encodedSig, nil
}

func loadPrivateKeyFromPEMFile(privKeyFileLocation string) (*rsa.PrivateKey, error) {
	fileData, err := ioutil.ReadFile(privKeyFileLocation)
	if err != nil {
		return nil, errors.Trace(err)
	}

	block, _ := pem.Decode(fileData)
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		return nil, errors.New("Failed to load a valid private key")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, errors.Trace(err)
	}

	return privateKey, nil
}
