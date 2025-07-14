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
	env_key := os.Getenv("GOOGLE_PROJECT_APIKEY")
	key_dec, err := base64.StdEncoding.DecodeString(env_key)
	//parsedKey, err := jwt.ParseRSAPrivateKeyFromPEM(key_dec)
	if err != nil {
		logrus.Warnf("fail to ParseRSAPrivateKeyFromPEM: %s", err)
		return nil
	}

	key := strings.Trim(strings.ReplaceAll(string(key_dec), "\n", " "), " ")
	logrus.Warnf("EnvsBase64: key is %s", key)
	if err != nil {
		logrus.Warnf("fail to decode base64 env key: %s", err)
	}

	env_cred := os.Getenv("GOOGLE_CREDENTIALS")
	cred_dec, err := base64.StdEncoding.DecodeString(env_cred)
	//f_str := strings.ReplaceAll(string(cred_dec), "\\\"", "\"")
	// s_str := strings.ReplaceAll(string(cred_dec), "\\n", "\n")
	cred := strings.ReplaceAll(string(cred_dec), " ", "")
	logrus.Warnf("EnvsBase64: cred is %s", cred)
	if err != nil {
		logrus.Warnf("fail to decode base64 env cred: %s", err)
	}
	jctl := CreateGcpJsWebTokenCtl(jwt.SigningMethodRS256, string(key_dec), cred)
	if err = jctl.SetIssFromGCPCredential(); err != nil {
		logrus.Errorf("fail to get iss from gcp credential, %s", err)
	}
	return jctl
}

func TestGcpAccessToken(t *testing.T) {
	jctl := initJwtTesting()
	tk, err := jctl.ExchangeApiKey()
	if len(tk) == 0 {
		logrus.Warnf("fail to gen api key with go ctl. ")
	} else {
		logrus.Warnf("success gen api key with go ctl: " + tk)
	}
	assert.NotEqual(t, err, nil)
}

func TestJwt(t *testing.T) {
	jctl := initJwtTesting()
	tk := jctl.GenToken()
	if len(tk) == 0 {
		logrus.Warnf("fail to gen gcp token with go ctl. ")
	} else {
		logrus.Warnf("success gen gcp token with go ctl: " + tk)
	}
	assert.Equal(t, tk, "")
}
