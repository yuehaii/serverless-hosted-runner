package common

import (
	"encoding/base64"
	"os"
	"strings"
	"testing"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/ingka-group-digital/app-monitor-agent/logrus"
	"github.com/stretchr/testify/assert"
)

func initJwtTesting() IGcpJsWebToken {
	envKey := os.Getenv("GOOGLE_PROJECT_APIKEY")
	keyDec, err := base64.StdEncoding.DecodeString(envKey)
	//parsedKey, err := jwt.ParseRSAPrivateKeyFromPEM(keyDec)
	if err != nil {
		logrus.Warnf("fail to ParseRSAPrivateKeyFromPEM: %s", err)
		return nil
	}

	key := strings.Trim(strings.ReplaceAll(string(keyDec), "\n", " "), " ")
	logrus.Warnf("EnvsBase64: key is %s", key)
	if err != nil {
		logrus.Warnf("fail to decode base64 env key: %s", err)
	}

	envCred := os.Getenv("GOOGLE_CREDENTIALS")
	credDec, err := base64.StdEncoding.DecodeString(envCred)
	//f_str := strings.ReplaceAll(string(credDec), "\\\"", "\"")
	// s_str := strings.ReplaceAll(string(credDec), "\\n", "\n")
	cred := strings.ReplaceAll(string(credDec), " ", "")
	logrus.Warnf("EnvsBase64: cred is %s", cred)
	if err != nil {
		logrus.Warnf("fail to decode base64 env cred: %s", err)
	}
	jctl := CreateGcpJsWebTokenCtl(jwt.SigningMethodRS256, string(keyDec), cred)
	if err = jctl.SetIssFromGCPCredential(); err != nil {
		logrus.Errorf("fail to get iss from gcp credential, %s", err)
	}
	return jctl
}

func TestGcpAccessToken(t *testing.T) {
	jctl := initJwtTesting()
	tk, err := jctl.ExchangeAPIKey()
	if len(tk) == 0 {
		logrus.Warnf("fail to gen api key with go ctl. ")
	} else {
		logrus.Warnf("success gen api key with go ctl: %s", tk)
	}
	assert.NotEqual(t, err, nil)
}

func TestJwt(t *testing.T) {
	jctl := initJwtTesting()
	tk := jctl.GenToken()
	if len(tk) == 0 {
		logrus.Warnf("fail to gen gcp token with go ctl. ")
	} else {
		logrus.Warnf("success gen gcp token with go ctl: %s", tk)
	}
	assert.Equal(t, tk, "")
}
