package common

import (
	"crypto/rsa"
	"crypto/x509"
)

type Cryptography interface {
	EncryptMsg(msg string) string
	DecryptMsg(msg string) string
	RandStr(n int) string
	GenCertificate(ca *x509.Certificate, caPri *rsa.PrivateKey) (*x509.Certificate, *rsa.PrivateKey, error)
	GetCertificate(bool, bool) string
	LoadCertificate(string, bool) (any, error)
	GenKeys()
}

type Crypto struct {
	certLocation string
	workingDir   string
}
