package common

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/ingka-group-digital/app-monitor-agent/logrus"
	"golang.org/x/oauth2/google"
	gjwt "golang.org/x/oauth2/jwt"
)

type IJsWebToken interface {
	GenToken() string
	VerifySignature(sig string) (*rsa.PrivateKey, error)
}

type IGcpJsWebToken interface {
	IJsWebToken
	SetIssFromGCPCredential() error
	ExchangeApiKey() (string, error)
}

type JsWebToken struct {
	scope, iss, aud   string
	exp               time.Duration
	iat               int64
	sign_method       jwt.SigningMethod
	claims            *jwt.Token
	private_key       *rsa.PrivateKey
	private_key_bytes []byte
}

type GCPJswWebToken struct {
	JsWebToken
	gcp_credential string
}

func CreateGcpJsWebTokenCtl(sign_method jwt.SigningMethod, private_key string, gcp_credential string) IGcpJsWebToken {
	parsedKey, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(private_key))
	if err != nil {
		logrus.Warnf("fail to ParseRSAPrivateKeyFromPEM: %s", err)
		return nil
	}
	return &GCPJswWebToken{JsWebToken{"https://www.googleapis.com/auth/cloud-platform", "",
		"https://www.googleapis.com/oauth2/v4/token", time.Hour,
		time.Now().Unix(), sign_method, nil, parsedKey, []byte(private_key)},
		strings.ReplaceAll(gcp_credential, " ", "")}
}

func (j *GCPJswWebToken) SetIssFromGCPCredential() error {
	// this structure should only be ref here in this func. declare it as internal.
	type GcpCredential struct {
		Type                    string `json:"type"`
		ProjectID               string `json:"project_id"`
		PrivateKeyID            string `json:"private_key_id"`
		PrivateKey              string `json:"private_key"`
		ClientEmail             string `json:"client_email"`
		ClientID                string `json:"client_id"`
		AuthURI                 string `json:"auth_uri"`
		TokenURI                string `json:"token_uri"`
		AuthProviderX509CertURL string `json:"auth_provider_x509_cert_url"`
		ClientX509CertURL       string `json:"client_x509_cert_url"`
		UniverseDomain          string `json:"universe_domain"`
	}

	cred_js := GcpCredential{}
	if err := json.Unmarshal([]byte(j.gcp_credential), &cred_js); err != nil {
		logrus.Errorf("fail to parse iss from gcp credential: %s", err)
		return err
	}
	j.iss = cred_js.ClientEmail
	return nil
}

func (j *JsWebToken) GenToken() string {
	j.initToken()
	tk, err := j.signToken()
	if err != nil {
		logrus.Errorf("fail to generate jwt token: %s", err)
		return ""
	}
	return tk
}

func (j *JsWebToken) VerifySignature(sig string) (*rsa.PrivateKey, error) {
	parsed_token, err := jwt.Parse(sig, func(token *jwt.Token) (interface{}, error) {
		return j.private_key, nil
	})
	if err != nil {
		return nil, err
	}
	if !parsed_token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return nil, nil
}

func (j *JsWebToken) signToken() (signed_tk string, err error) {
	if j.claims == nil {
		return "", fmt.Errorf("fail to sign token, pls create claims first.")
	}
	return j.claims.SignedString(j.private_key)
}

func (j *JsWebToken) initToken() {
	j.claims = jwt.NewWithClaims(j.sign_method, jwt.MapClaims{
		"scope": j.scope,
		"iss":   j.iss,
		"aud":   j.aud,
		"exp":   time.Now().Add(j.exp).Unix(),
		"iat":   j.iat,
	})
}

func (j *JsWebToken) ExchangeApiKey() (string, error) {
	conf := &gjwt.Config{
		Email:      j.iss,
		PrivateKey: j.private_key_bytes,
		Scopes: []string{
			j.scope,
		},
		TokenURL: google.JWTTokenURL,
		Expires:  j.exp, //less than 1 hour
		Audience: j.aud,
	}
	tk, err := conf.TokenSource(context.Background()).Token()
	if err != nil {
		logrus.Errorf("fail to call token source: %s", err)
		return "", err
	}
	return tk.AccessToken, err
}
