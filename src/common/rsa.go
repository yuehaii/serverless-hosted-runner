package common

import (
	crypto_rand "crypto/rand"
	"crypto/rsa"
	"crypto/sha512"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"io/ioutil"
	"math/big"
	"net"
	"os"
	"strings"
	"time"

	"github.com/ingka-group-digital/app-monitor-agent/logrus"
)

type RSACrypto struct {
	Crypto
	host      string
	validFrom string
	validTo   time.Duration
	bits      int
	org       string
}

func RSACryptography(host string) Cryptography {
	if !strings.Contains(host, "127.0.0.1") && len(host) > 0 {
		host += ",127.0.0.1"
	} else if len(host) == 0 {
		host = "127.0.0.1"
	}
	return RSACrypto{Crypto{"./certs/", "/go/bin/"}, host, "", 365 * 24 * time.Hour, 2048, "ccoecn"}
}

func (crypt RSACrypto) GetCertificate(isKey bool, isCa bool) string {
	certFile := crypt.workingDir + crypt.certLocation + Ternary(isCa, "ca.", "leaf.").(string) +
		Ternary(isKey, "key", "cert").(string) + ".pem"
	_, err := os.Stat(certFile)
	if err != nil {
		logrus.Warnf("cert file %s dose not exists", certFile)
		return ""
	}
	return certFile
}

func (crypt RSACrypto) LoadCertificate(filepath string, isKey bool) (any, error) {
	certBytes, err := os.ReadFile(filepath)
	if err != nil {
		logrus.Errorf("Fail to read cert file: %s, err: %v", filepath, err)
		return nil, err
	}
	derBytes, _ := pem.Decode(certBytes)
	if derBytes == nil {
		logrus.Errorf("Fail to decode cert: %v", err)
		return nil, err
	}

	if isKey {
		key, err := x509.ParsePKCS8PrivateKey(derBytes.Bytes)
		if err != nil {
			logrus.Errorf("Fail to parse key: %v", err)
		}
		logrus.Printf("Success load key")
		return key, nil
	} else {
		cert, err := x509.ParseCertificate(derBytes.Bytes)
		if err != nil {
			logrus.Errorf("Fail to parse cert: %v", err)
			return nil, err
		}
		logrus.Printf("Success load cert with serial number:%s", cert.SerialNumber.String())
		return cert, nil
	}
}

func (crypt RSACrypto) GenCertificate(ca *x509.Certificate, caPri *rsa.PrivateKey) (*x509.Certificate, *rsa.PrivateKey, error) {
	var priv any
	var err error
	priv, err = rsa.GenerateKey(crypto_rand.Reader, crypt.bits)
	if err != nil {
		logrus.Errorf("Fail to generate key: %v", err)
		return nil, nil, err
	}

	keyUsage := x509.KeyUsageDigitalSignature
	keyUsage |= x509.KeyUsageKeyEncipherment
	var notBefore time.Time
	if len(crypt.validFrom) == 0 {
		notBefore = time.Now()
	} else {
		notBefore, err = time.Parse("2006-01-02 15:04:05", crypt.validFrom)
		if err != nil {
			logrus.Errorf("Fail to parse creation date: %v", err)
			return nil, nil, err
		}
	}
	notAfter := notBefore.Add(crypt.validTo)

	serialNumLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNum, err := crypto_rand.Int(crypto_rand.Reader, serialNumLimit)
	if err != nil {
		logrus.Errorf("Fail to generate serial number: %v", err)
		return nil, nil, err
	}

	certTemp := x509.Certificate{
		SerialNumber: serialNum,
		Subject: pkix.Name{
			Organization: []string{crypt.org},
			Country:      []string{"CN"},
			Province:     []string{"Shanghai"},
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              keyUsage,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
	}
	hosts := strings.Split(crypt.host, ",")
	for _, h := range hosts {
		if ip := net.ParseIP(h); ip != nil {
			certTemp.IPAddresses = append(certTemp.IPAddresses, ip)
		} else {
			certTemp.DNSNames = append(certTemp.DNSNames, h)
		}
	}
	if ca == nil {
		certTemp.IsCA = true
		certTemp.KeyUsage |= x509.KeyUsageCertSign
	}

	publicKey := func(priv any) any {
		switch k := priv.(type) {
		case *rsa.PrivateKey:
			return &k.PublicKey
		default:
			return nil
		}
	}
	var derBytes []byte
	var certPrefix string
	if ca != nil && caPri != nil {
		certPrefix = crypt.certLocation + "leaf."
		derBytes, err = x509.CreateCertificate(crypto_rand.Reader, &certTemp, ca, publicKey(priv), caPri)
	} else {
		certPrefix = crypt.certLocation + "ca."
		derBytes, err = x509.CreateCertificate(crypto_rand.Reader, &certTemp, &certTemp, publicKey(priv), priv)
	}
	if err != nil {
		logrus.Errorf("Fail to create certificate: %v", err)
		return nil, nil, err
	}
	certOut, err := os.Create(certPrefix + "cert.pem")
	if err != nil {
		logrus.Errorf("Fail to open cert.pem for writing: %v", err)
		return nil, nil, err
	}
	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		logrus.Errorf("Fail to write data to cert.pem: %v", err)
	}
	if err := certOut.Close(); err != nil {
		logrus.Errorf("Error closing cert.pem: %v", err)
	}
	logrus.Infof("Finish %s generation", certPrefix+"cert.pem")

	keyOut, err := os.OpenFile(certPrefix+"key.pem", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		logrus.Errorf("Fail to open key.pem for writing: %v", err)
		return nil, nil, err
	}
	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		logrus.Errorf("Unable to marshal private key: %v", err)
		return nil, nil, err
	}
	if err := pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes}); err != nil {
		logrus.Errorf("Fail to write data to key.pem: %v", err)
		return nil, nil, err
	}
	if err := keyOut.Close(); err != nil {
		logrus.Errorf("Error closing key.pem: %v", err)
		return nil, nil, err
	}
	logrus.Infof("Finish %s generation", certPrefix+"key.pem")
	return &certTemp, priv.(*rsa.PrivateKey), nil
}

func (crypt RSACrypto) GenKeys() {
	if pub, pri, err := crypt.genRSAKeys(); err == nil {
		if err := crypt.saveKeys(crypt.certLocation+"rsaPub.key", crypt.pubKeyToPEM(pub)); err != nil {
			logrus.Errorf("fail to save pub key, %v", err)
		}
		if err := crypt.saveKeys(crypt.certLocation+"rsaPri.key", crypt.priKeyToPEM(pri)); err != nil {
			logrus.Errorf("fail to save pri key, %v", err)
		}
	} else {
		logrus.Errorf("fail to gen rsa keys, %v", err)
	}
}

func (crypt RSACrypto) saveKeys(keyPath string, content string) error {
	kHandler, err := os.Create(keyPath)
	if err != nil {
		logrus.Errorf("fail to create key file: %v", err)
		return err
	}
	if _, err := kHandler.WriteString(content); err != nil {
		logrus.Errorf("fail to write key file: %v", err)
	}
	if err := kHandler.Close(); err != nil {
		logrus.Errorf("fail to close key file: %v", err)
		return err
	}
	return nil
}

func (crypt RSACrypto) genRSAKeys() (*rsa.PublicKey, *rsa.PrivateKey, error) {
	privateKey, err := rsa.GenerateKey(crypto_rand.Reader, 2048)
	if err != nil {
		logrus.Errorf("fail to genRSAKeys, %v", err)
		return nil, nil, err
	}
	return &privateKey.PublicKey, privateKey, nil
}

func (crypt RSACrypto) priKeyToPEM(priKey *rsa.PrivateKey) string {
	if priKey != nil {
		return string(
			pem.EncodeToMemory(
				&pem.Block{
					Type:  "RSA PRIVATE KEY",
					Bytes: x509.MarshalPKCS1PrivateKey(priKey),
				},
			),
		)
	} else {
		return ""
	}
}

func (crypt RSACrypto) bytesToPEM(data []byte) string {
	return string(
		pem.EncodeToMemory(
			&pem.Block{
				Type:  "MESSAGE",
				Bytes: data,
			},
		),
	)
}

func (crypt RSACrypto) pubKeyToPEM(pubKey *rsa.PublicKey) string {
	if pubKey != nil {
		return string(
			pem.EncodeToMemory(
				&pem.Block{
					Type:  "RSA PUBLIC KEY",
					Bytes: x509.MarshalPKCS1PublicKey(pubKey),
				},
			),
		)
	} else {
		return ""
	}
}

func (crypt RSACrypto) decPEMBlock(data []byte) (blockBytes []byte, err error) {
	decBlocks, _ := pem.Decode(data)
	blockBytes = decBlocks.Bytes
	passEncrypted := x509.IsEncryptedPEMBlock(decBlocks)
	if passEncrypted {
		blockBytes, err = x509.DecryptPEMBlock(decBlocks, nil)
		if err != nil {
			logrus.Errorf("fail to decPEMBlock, %v", err)
			return nil, err
		}
	}
	return blockBytes, nil
}

func (crypt RSACrypto) bytesToPubKey(data []byte) (pubKey *rsa.PublicKey, err error) {
	if blockBytes, err := crypt.decPEMBlock(data); err == nil {
		pubKey, err = x509.ParsePKCS1PublicKey(blockBytes)
		if err != nil {
			logrus.Errorf("fail to parse pkcs1 pub key, %v", err)
			return nil, err
		}
		return pubKey, nil
	} else {
		logrus.Errorf("fail to decPEMBlock in bytesToPubKey, %v", err)
		return nil, err
	}
}

func (crypt RSACrypto) bytesToPriKey(data []byte) (priKey *rsa.PrivateKey, err error) {
	if blockBytes, err := crypt.decPEMBlock(data); err == nil {
		priKey, err = x509.ParsePKCS1PrivateKey(blockBytes)
		if err != nil {
			logrus.Errorf("fail to parse pkcs1 pri key, %v", err)
			return nil, err
		}
		return priKey, nil
	} else {
		logrus.Errorf("fail to decPEMBlock in bytesToPriKey, %v", err)
		return nil, err
	}
}

func (crypt RSACrypto) EncryptMsg(msg string) string {
	bytes, err := ioutil.ReadFile(crypt.certLocation + "rsaPub.key")
	if err != nil {
		logrus.Errorf("fail to ReadFile for EncryptMsg, %v", err)
		return ""
	}
	publicKey, err := crypt.bytesToPubKey(bytes)
	if err != nil {
		logrus.Errorf("fail to bytesToPubKey for EncryptMsg, %v", err)
		return ""
	}
	cipher, err := rsa.EncryptOAEP(sha512.New(), crypto_rand.Reader, publicKey, []byte(msg), nil)
	if err != nil {
		logrus.Errorf("fail to EncryptOAEP for EncryptMsg, %v", err)
		return ""
	}
	return crypt.bytesToPEM(cipher)
}

func (crypt RSACrypto) DecryptMsg(msg string) string {
	bytes, err := ioutil.ReadFile(crypt.certLocation + "rsaPri.key")
	if err != nil {
		logrus.Errorf("fail to ReadFile for DecryptMsg, %v", err)
		return ""
	}
	priKey, err := crypt.bytesToPriKey(bytes)
	if err != nil {
		logrus.Errorf("fail to bytesToPriKey for DecryptMsg, %v", err)
		return ""
	}
	pemToCipher := func(encMsg string) []byte {
		decP, _ := pem.Decode([]byte(encMsg))
		return decP.Bytes
	}
	decByte, err := rsa.DecryptOAEP(
		sha512.New(),
		crypto_rand.Reader,
		priKey,
		pemToCipher(msg),
		nil,
	)
	if err != nil {
		logrus.Errorf("fail to DecryptOAEP for DecryptMsg, %v", err)
		return ""
	}
	return string(decByte)
}

func (crypt RSACrypto) RandStr(n int) string {
	return ""
}
