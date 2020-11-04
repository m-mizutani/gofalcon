package gofalcon

import (
	"bytes"
	"net/url"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// OAuth2API provides oauth2 get token and revoke token operation.
type OAuth2API struct {
	client *Client
}

// TokenInput is arguments of Token
type TokenInput struct {
	ClientID     *string // client_id
	ClientSecret *string // client_secret
}

// TokenOutput is a result of Token
type TokenOutput struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

// Token generates an OAuth2 access token
func (x *OAuth2API) Token(input *TokenInput) (*TokenOutput, error) {
	qs := url.Values{}
	buf := bytes.Buffer{}
	if input.ClientID != nil {
		qs.Add("client_id", *input.ClientID)
	}
	if input.ClientSecret != nil {
		qs.Add("client_secret", *input.ClientSecret)
	}

	buf.Write([]byte(qs.Encode()))

	req := Request{
		Method:  "POST",
		Path:    "oauth2/token",
		Body:    bytes.NewReader(buf.Bytes()),
		Headers: []httpHeader{{"Content-Type", "application/x-www-form-urlencoded"}},
	}

	var output TokenOutput
	if err := x.client.SendRequest(req, &output); err != nil {
		return nil, errors.Wrap(err, "Fail to OAuth2 Token")
	}

	Logger.WithFields(logrus.Fields{
		"client_id": *input.ClientID,
	}).Debug("Authorized OAuth2")

	return &output, nil
}

// RevokeInput is arguments of OAuth2API.Revoke
type RevokeInput struct {
	Token *string // token
}

// RevokeOutput is a result of OAuth2API.Revoke
type RevokeOutput struct{}

// Revoke disable oauth2 token
func (x *OAuth2API) Revoke(input *RevokeInput) (*RevokeOutput, error) {
	qs := url.Values{}
	buf := bytes.Buffer{}
	qs.Add("token", *input.Token)
	buf.Write([]byte(qs.Encode()))

	req := Request{
		Method:  "POST",
		Path:    "oauth2/revoke",
		Body:    bytes.NewReader(buf.Bytes()),
		Headers: []httpHeader{{"Content-Type", "application/x-www-form-urlencoded"}},
	}

	var output RevokeOutput
	if err := x.client.SendRequest(req, &output); err != nil {
		return nil, errors.Wrap(err, "Fail to revoke OAuth2 Token")
	}

	Logger.Debug("Revoked OAuth2 Token")

	return &output, nil
}
