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
	"github.com/sirupsen/logrus"
)

// Logger is exported to be controllable from external.
var Logger = logrus.New()

func init() {
	Logger.SetLevel(logrus.FatalLevel)
}

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
		Endpoint: "https://api.us-2.crowdstrike.com",
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

// Request is data set of API request. Method is HTTP method. Path should be set like "devices/queries/devices/v1". QueryString and Body are optional. Headers can also be modified, but basically no need to modify.
type Request struct {
	Method      string
	Path        string
	QueryString url.Values
	Body        io.Reader
	Headers     []httpHeader
}

// Response is generic falcon API response
type Response struct {
	BaseResponse
	Resources []interface{} `json:"resources"`
}

type BaseResponse struct {
	Errors []ServerError `json:"errors"`
	Meta   MetaData      `json:"meta"`
}

type Pagenation struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
	Total  int `json:"total"`
}

type MetaData struct {
	PoweredBy  string      `json:"powered_by"`
	QueryTime  float64     `json:"query_time"`
	TraceID    string      `json:"trace_id"`
	Pagenation *Pagenation `json:"pagination"`
}

type ServerError struct {
	Code    int    `json:"code"`
	ID      string `json:"id"`
	Message string `json:"message"`
}

// SendRequest sends any request to API endpoint and set results to v. This function retry the request if OAuth2 token is expired.
func (x *Client) SendRequest(req Request, resp interface{}) error {
	if err := x.sendHTTPRequest(req, resp); err != nil {
		if _, ok := err.(*authError); !ok {
			return err // General error
		}

		if err := x.refreshOAuth2Token(); err != nil {
			return err // Can not refresh token
		}

		// Retry
		return x.sendHTTPRequest(req, resp)
	}

	return nil
}

type authError struct {
	err error
}

func (x *authError) Error() string {
	return x.err.Error()
}

func (x *Client) sendHTTPRequest(req Request, resp interface{}) error {
	client := &http.Client{}
	endpoint := x.Endpoint
	if strings.HasSuffix(endpoint, "/") {
		endpoint = endpoint[:len(endpoint)-1]
	}
	path := req.Path
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	url := endpoint + path
	if len(req.QueryString) > 0 {
		url = url + "?" + req.QueryString.Encode()
	}

	httpReq, err := http.NewRequest(req.Method, url, req.Body)
	if err != nil {
		return errors.Wrap(err, "fail to create a graylog http request")
	}

	switch {
	case x.OAuth2Token != "" && x.OAuth2Type != "":
		auth := x.OAuth2Type + " " + x.OAuth2Token
		httpReq.Header.Add("authorization", auth)
	case x.User != "" && x.Token != "":
		httpReq.SetBasicAuth(x.User, x.Token)
	}

	if req.Body != nil {
		// Set default content type
		httpReq.Header.Add("content-type", "application/json")
	}

	httpReq.Header.Add("accept", "application/json")
	for _, hdr := range req.Headers {
		httpReq.Header.Set(hdr.Name, hdr.Value)
	}

	httpResp, err := client.Do(httpReq)
	if err != nil {
		return errors.Wrap(err, "fail to send request to server")
	}

	defer httpResp.Body.Close()
	rawData, err := ioutil.ReadAll(httpResp.Body)
	if err != nil {
		return errors.Wrap(err, "Fail to read httpResponse from server")
	}

	// Error handling
	if httpResp.StatusCode == 403 {
		return &authError{err: fmt.Errorf("Authentication Error (HTTP 403): %s", string(rawData))}
	} else if httpResp.StatusCode >= 400 {
		return fmt.Errorf("Fail HTTP request %d: %s", httpResp.StatusCode, string(rawData))
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

	if err := json.Unmarshal(rawData, resp); err != nil {
		return errors.Wrapf(err, "Fail to parse reponse of Falcon: %v", string(rawData))
	}

	return nil
}

// Int converts int to pointer
func Int(v int) *int { return &v }

// IntValue returns int from *int and returns 0 if nil
func IntValue(v *int) int {
	if v == nil {
		return 0
	}
	return *v
}

// String converts string to pointer
func String(v string) *string { return &v }

// StringValue returns string from *string and returns "" if nil
func StringValue(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}
