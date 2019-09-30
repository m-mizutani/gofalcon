package gofalcon

import (
	"bytes"
	"net/url"

	"github.com/pkg/errors"
)

// OAuth2API provides oauth2 get token and revoke token operation.
type OAuth2API struct {
	client *Client
}

type TokenInput struct {
	ClientID     string // client_id
	ClientSecret string // client_secret
}

type TokenOutput struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

// Token generates an OAuth2 access token
func (x *OAuth2API) Token(input *TokenInput) (*TokenOutput, error) {
	qs := url.Values{}
	buf := bytes.Buffer{}
	qs.Add("client_id", input.ClientID)
	qs.Add("client_secret", input.ClientSecret)
	buf.Write([]byte(qs.Encode()))

	req := request{
		Method:  "POST",
		Path:    "oauth2/token",
		Body:    bytes.NewReader(buf.Bytes()),
		Headers: []httpHeader{{"Content-Type", "application/x-www-form-urlencoded"}},
	}

	var output TokenOutput
	if err := x.client.sendRequest(req, &output); err != nil {
		return nil, errors.Wrap(err, "Fail to OAuth2 Token")
	}

	return &output, nil
}

type RevokeInput struct {
	Token string // token
}

type RevokeOutput struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

// Token generates an OAuth2 access token
func (x *OAuth2API) Revoke(input *TokenInput) (*TokenOutput, error) {
	qs := url.Values{}
	buf := bytes.Buffer{}
	qs.Add("client_id", input.ClientID)
	qs.Add("client_secret", input.ClientSecret)
	buf.Write([]byte(qs.Encode()))

	req := request{
		Method:  "POST",
		Path:    "oauth2/token",
		Body:    bytes.NewReader(buf.Bytes()),
		Headers: []httpHeader{{"Content-Type", "application/x-www-form-urlencoded"}},
	}

	var output TokenOutput
	if err := x.client.sendRequest(req, &output); err != nil {
		return nil, errors.Wrap(err, "Fail to OAuth2 Token")
	}

	return &output, nil
}
