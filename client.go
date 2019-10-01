package gofalcon

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/pkg/errors"
)

// Client is Falcon API client
type Client struct {
	User        string
	Token       string
	Endpoint    string
	OAuth2Token string
	OAuth2Type  string
	ClientID    string
	Secret      string

	Device    *DeviceAPI
	OAuth2    *OAuth2API
	Detection *DetectionAPI
	Sensor    *SensorAPI
}

// NewClient is constructor of Client
func NewClient() *Client {
	client := Client{
		Endpoint: "https://api.crowdstrike.com",
	}
	client.Device = &DeviceAPI{client: &client}
	client.OAuth2 = &OAuth2API{client: &client}
	client.Detection = &DetectionAPI{client: &client}
	client.Sensor = &SensorAPI{client: &client}

	return &client
}

// EnableOAuth2 retrieves OAuth2 token and set it to the client
func (x *Client) EnableOAuth2(clientID, secret string) error {
	x.ClientID = clientID
	x.Secret = secret

	return x.refreshOAuth2Token()
}

func (x *Client) refreshOAuth2Token() error {
	resp, err := x.OAuth2.Token(&TokenInput{
		ClientID:     &x.ClientID,
		ClientSecret: &x.Secret,
	})
	if err != nil {
		return errors.Wrap(err, "Fail to OAuth2 authentication")
	}

	x.SetOAuth2Token(resp.AccessToken, resp.TokenType)
	return nil
}

// SetOAuth2Token sets OAuth2Token already generated
func (x *Client) SetOAuth2Token(token, tokenType string) {
	x.OAuth2Token = token
	x.OAuth2Type = tokenType
}

// SetUserToken sets user and token for authorization
func (x *Client) SetUserToken(user, token string) {
	x.User = user
	x.Token = token
}

type httpHeader struct {
	Name  string
	Value string
}

type request struct {
	Method      string
	Path        string
	QueryString url.Values
	Body        io.Reader
	Headers     []httpHeader
}

func (x *Client) sendRequest(req request, v interface{}) error {
	if err := x.sendHTTPRequest(req, v); err != nil {
		if _, ok := err.(*authError); !ok {
			return err // General error
		}

		if err := x.refreshOAuth2Token(); err != nil {
			return err // Can not refresh token
		}

		// Retry
		return x.sendHTTPRequest(req, v)
	}

	return nil
}

type authError struct {
	err error
}

func (x *authError) Error() string {
	return x.err.Error()
}

func (x *Client) sendHTTPRequest(req request, v interface{}) error {
	client := &http.Client{}
	url := fmt.Sprintf("%s/%s", x.Endpoint, req.Path)
	if len(req.QueryString) > 0 {
		url = url + "?" + req.QueryString.Encode()
	}

	r, err := http.NewRequest(req.Method, url, req.Body)
	if err != nil {
		return errors.Wrap(err, "fail to create a graylog http request")
	}

	switch {
	case x.OAuth2Token != "" && x.OAuth2Type != "":
		auth := x.OAuth2Type + " " + x.OAuth2Token
		r.Header.Add("authorization", auth)
	case x.User != "" && x.Token != "":
		r.SetBasicAuth(x.User, x.Token)
	}

	r.Header.Add("accept", "application/json")
	for _, hdr := range req.Headers {
		r.Header.Add(hdr.Name, hdr.Value)
	}

	resp, err := client.Do(r)
	if err != nil {
		return errors.Wrap(err, "fail to send request to server")
	}

	defer resp.Body.Close()
	rawData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "Fail to read response from server")
	}

	// Error handling
	if resp.StatusCode == 403 {
		return &authError{err: fmt.Errorf("Authentication Error (HTTP 403): %s", string(rawData))}
	} else if resp.StatusCode >= 400 {
		return fmt.Errorf("Fail HTTP request %d: %s", resp.StatusCode, string(rawData))
	}

	var base BaseResponse
	if err := json.Unmarshal(rawData, &base); err != nil {
		return errors.Wrapf(err, "Fail to parse base reponse of Falcon: %v", string(rawData))
	}
	if len(base.Errors) > 0 {
		var messages []string
		for _, e := range base.Errors {
			messages = append(messages, fmt.Sprintf("%d: %s", e.Code, e.Message))
		}
		return fmt.Errorf("Fail to request: %s", strings.Join(messages, ", "))
	}

	if err := json.Unmarshal(rawData, v); err != nil {
		return errors.Wrapf(err, "Fail to parse reponse of Falcon: %v", string(rawData))
	}

	return nil
}

// Int converts int to pointer
func Int(v int) *int { return &v }

// String converts string to pointer
func String(v string) *string { return &v }
