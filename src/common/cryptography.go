package common

import (
	"crypto/aes"
	"crypto/cipher"
	crypto_rand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"io/ioutil"
	"math/big"
	"math/rand"
	"net"
	"os"
	"strings"
	"time"

	"github.com/ingka-group-digital/app-monitor-agent/logrus"
)

type Cryptography interface {
	EncryptMsg(msg string) string
	DecryptMsg(msg string) string
	RandStr(n int) string
	GenCertificate(ca *x509.Certificate, ca_pri *rsa.PrivateKey) (error, *x509.Certificate, *rsa.PrivateKey)
	GetCertificate(bool, bool) string
	LoadCertificate(string, bool) (error, any)
}

func DefaultCryptography(key string) Cryptography {
	return AESCryptography{strings.Trim(key, " "), "aescfb", []byte{10, 31, 41, 22, 21, 20, 11, 65, 76, 34, 99, 02, 47, 36, 11, 32}}
}

func DESCryptography(key string) Cryptography {
	return nil
}

func RSACryptography(host string) Cryptography {
	if !strings.Contains(host, "127.0.0.1") && len(host) > 0 {
		host += ",127.0.0.1"
	} else if len(host) == 0 {
		host = "127.0.0.1"
	}
	return RSACrypto{host, "", 365 * 24 * time.Hour, 2048, "ccoecn", "./certs/", "/go/bin/"}
}

type AESCryptography struct {
	key       string
	algorithm string
	iv        []byte
}

type RSACrypto struct {
	host          string
	valid_from    string
	valid_to      time.Duration
	bits          int
	org           string
	cert_location string
	working_dir   string
}

func (crypt RSACrypto) GetCertificate(is_key bool, is_ca bool) string {
	cert_f := crypt.working_dir + crypt.cert_location + Ternary(is_ca, "ca.", "leaf.").(string) +
		Ternary(is_key, "key", "cert").(string) + ".pem"
	_, err := os.Stat(cert_f)
	if err != nil {
		logrus.Warnf("cert file %s dose not exists." + cert_f)
		return ""
	}
	return cert_f
}

func (crypt RSACrypto) LoadCertificate(filepath string, is_key bool) (error, any) {
	cert_bytes, err := ioutil.ReadFile(filepath)
	if err != nil {
		logrus.Errorf("Fail to read cert file: %s, err: %v", filepath, err)
		return err, nil
	}
	der_bytes, _ := pem.Decode(cert_bytes)
	if der_bytes == nil {
		logrus.Errorf("Fail to decode cert: %v", err)
		return err, nil
	}

	if is_key {
		key, err := x509.ParsePKCS8PrivateKey(der_bytes.Bytes)
		if err != nil {
			logrus.Errorf("Fail to parse key: %v", err)
		}
		logrus.Printf("Success load key")
		return nil, key
	} else {
		cert, err := x509.ParseCertificate(der_bytes.Bytes)
		if err != nil {
			logrus.Errorf("Fail to parse cert: %v", err)
			return err, nil
		}
		logrus.Printf("Success load cert with serial number:%s", cert.SerialNumber.String())
		return nil, cert
	}
}

func (crypt RSACrypto) GenCertificate(ca *x509.Certificate, ca_pri *rsa.PrivateKey) (error, *x509.Certificate, *rsa.PrivateKey) {
	var priv any
	var err error
	priv, err = rsa.GenerateKey(crypto_rand.Reader, *&crypt.bits)

	key_usage := x509.KeyUsageDigitalSignature
	key_usage |= x509.KeyUsageKeyEncipherment
	var not_before time.Time
	if len(crypt.valid_from) == 0 {
		not_before = time.Now()
	} else {
		not_before, err = time.Parse("Jan 1 20:04:05 2025", crypt.valid_from)
		if err != nil {
			logrus.Errorf("Fail to parse creation date: %v", err)
			return err, nil, nil
		}
	}
	not_after := not_before.Add(crypt.valid_to)

	serial_num_limit := new(big.Int).Lsh(big.NewInt(1), 128)
	serial_num, err := crypto_rand.Int(crypto_rand.Reader, serial_num_limit)
	if err != nil {
		logrus.Errorf("Fail to generate serial number: %v", err)
		return err, nil, nil
	}

	cert_temp := x509.Certificate{
		SerialNumber: serial_num,
		Subject: pkix.Name{
			Organization: []string{crypt.org},
			Country:      []string{"CN"},
			Province:     []string{"Shanghai"},
		},
		NotBefore:             not_before,
		NotAfter:              not_after,
		KeyUsage:              key_usage,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
	}
	hosts := strings.Split(crypt.host, ",")
	for _, h := range hosts {
		if ip := net.ParseIP(h); ip != nil {
			cert_temp.IPAddresses = append(cert_temp.IPAddresses, ip)
		} else {
			cert_temp.DNSNames = append(cert_temp.DNSNames, h)
		}
	}
	if ca == nil {
		cert_temp.IsCA = true
		cert_temp.KeyUsage |= x509.KeyUsageCertSign
	}

	publicKey := func(priv any) any {
		switch k := priv.(type) {
		case *rsa.PrivateKey:
			return &k.PublicKey
		default:
			return nil
		}
	}
	var der_bytes []byte
	var cert_prefix string
	if ca != nil && ca_pri != nil {
		cert_prefix = crypt.cert_location + "leaf."
		der_bytes, err = x509.CreateCertificate(crypto_rand.Reader, &cert_temp, ca, publicKey(priv), ca_pri)
	} else {
		cert_prefix = crypt.cert_location + "ca."
		der_bytes, err = x509.CreateCertificate(crypto_rand.Reader, &cert_temp, &cert_temp, publicKey(priv), priv)
	}
	if err != nil {
		logrus.Errorf("Fail to create certificate: %v", err)
		return err, nil, nil
	}
	cert_out, err := os.Create(cert_prefix + "cert.pem")
	if err != nil {
		logrus.Errorf("Fail to open cert.pem for writing: %v", err)
		return err, nil, nil
	}
	if err := pem.Encode(cert_out, &pem.Block{Type: "CERTIFICATE", Bytes: der_bytes}); err != nil {
		logrus.Errorf("Fail to write data to cert.pem: %v", err)
	}
	if err := cert_out.Close(); err != nil {
		logrus.Errorf("Error closing cert.pem: %v", err)
	}
	logrus.Infof("Finish %s generation", cert_prefix+"cert.pem")

	key_out, err := os.OpenFile(cert_prefix+"key.pem", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		logrus.Errorf("Fail to open key.pem for writing: %v", err)
		return err, nil, nil
	}
	priv_bytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		logrus.Errorf("Unable to marshal private key: %v", err)
		return err, nil, nil
	}
	if err := pem.Encode(key_out, &pem.Block{Type: "PRIVATE KEY", Bytes: priv_bytes}); err != nil {
		logrus.Errorf("Fail to write data to key.pem: %v", err)
		return err, nil, nil
	}
	if err := key_out.Close(); err != nil {
		logrus.Errorf("Error closing key.pem: %v", err)
		return err, nil, nil
	}
	logrus.Infof("Finish %s generation", cert_prefix+"key.pem")
	return nil, &cert_temp, priv.(*rsa.PrivateKey)
}

func (crypt RSACrypto) EncryptMsg(msg string) string {
	return ""
}

func (crypt RSACrypto) DecryptMsg(msg string) string {
	return ""
}

func (crypt RSACrypto) RandStr(n int) string {
	return ""
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

func (crypt AESCryptography) GenCertificate(ca *x509.Certificate, ca_pri *rsa.PrivateKey) (error, *x509.Certificate, *rsa.PrivateKey) {
	return nil, nil, nil
}

func (crypt AESCryptography) GetCertificate(is_key bool, is_ca bool) string {
	return ""
}

func (crypt AESCryptography) LoadCertificate(filepath string, is_key bool) (error, any) {
	return nil, nil
}
