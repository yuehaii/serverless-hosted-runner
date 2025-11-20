package common

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"math/rand"
	"strings"

	"github.com/ingka-group-digital/app-monitor-agent/logrus"
)

func DefaultCryptography(key string) Cryptography {
	return AESCryptography{Crypto{"./certs/", "/go/bin/"}, strings.Trim(key, " "), "aescfb",
		[]byte{10, 31, 41, 22, 21, 20, 11, 65, 76, 34, 99, 02, 47, 36, 11, 32}}
}

func DESCryptography(key string) Cryptography {
	return nil
}

type AESCryptography struct {
	Crypto
	key       string
	algorithm string
	iv        []byte
}

func (crypt AESCryptography) RandStr(n int) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
func (crypt AESCryptography) encode(byteMsg []byte) string {
	return base64.StdEncoding.EncodeToString(byteMsg)
}
func (crypt AESCryptography) decode(byteMsg string) []byte {
	decodeData, err := base64.StdEncoding.DecodeString(byteMsg)
	if err != nil {
		logrus.Errorln(err)
		return nil
	}
	return decodeData
}

func (crypt AESCryptography) aesCfbEncryption(msg string) (string, error) {
	block, err := aes.NewCipher([]byte(crypt.key))
	if err != nil {
		return "", err
	}
	plain := []byte(msg)
	cfb := cipher.NewCFBEncrypter(block, crypt.iv)
	cp := make([]byte, len(plain))
	cfb.XORKeyStream(cp, plain)
	return crypt.encode(cp), nil
}
func (crypt AESCryptography) aesCfbDecryption(msg string) (string, error) {
	block, err := aes.NewCipher([]byte(crypt.key))
	if err != nil {
		return "", err
	}
	cp := crypt.decode(msg)
	cfb := cipher.NewCFBDecrypter(block, crypt.iv)
	plain := make([]byte, len(cp))
	cfb.XORKeyStream(plain, cp)
	return string(plain), nil
}
func (crypt AESCryptography) EncryptMsg(msg string) string {
	if msg == "null" || len(msg) <= 0 {
		return msg
	}
	text, err := crypt.aesCfbEncryption(msg)
	if err != nil {
		logrus.Errorln("error encrypting text: ", err)
	}
	return text
}

func (crypt AESCryptography) DecryptMsg(msg string) string {
	if msg == "null" || len(msg) <= 0 {
		return msg
	}
	text, err := crypt.aesCfbDecryption(msg)
	if err != nil {
		logrus.Errorln("error decrypting encrypted text: ", err)
	}
	return text
}

func (crypt AESCryptography) GenCertificate(ca *x509.Certificate, caPri *rsa.PrivateKey) (*x509.Certificate, *rsa.PrivateKey, error) {
	return nil, nil, nil
}

func (crypt AESCryptography) GetCertificate(isKey bool, isCa bool) string {
	return ""
}

func (crypt AESCryptography) LoadCertificate(filepath string, isKey bool) (any, error) {
	return nil, nil
}

func (crypt AESCryptography) GenKeys() {
}
