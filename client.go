package gofalcon

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
)

// Client is Falcon API client
type Client struct {
	User     string
	Token    string
	Endpoint string

	Device *Device
}

// NewClient is constructor of Client
func NewClient(user, token string) Client {
	client := Client{
		User:     user,
		Token:    token,
		Endpoint: "https://falconapi.crowdstrike.com",
	}
	client.Device = &Device{client: &client}

	return client
}

type request struct {
	Method      string
	Path        string
	QueryString url.Values
}

func (x *Client) sendRequest(req request, v interface{}) error {
	client := &http.Client{}
	url := fmt.Sprintf("%s/%s", x.Endpoint, req.Path)
	if len(req.QueryString) > 0 {
		url = url + "?" + req.QueryString.Encode()
	}

	r, err := http.NewRequest(req.Method, url, nil)
	if err != nil {
		return errors.Wrap(err, "fail to create a graylog http request")
	}
	r.SetBasicAuth(x.User, x.Token)
	resp, err := client.Do(r)
	if err != nil {
		return errors.Wrap(err, "fail to send request to server")
	}

	defer resp.Body.Close()
	rawData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "Fail to read response from server")
	}

	if err := json.Unmarshal(rawData, v); err != nil {
		// logger.WithField("reponse", string(rawData)).Error("Unexpected response")
		return errors.Wrap(err, "Fail to parse reponse of Falcon")
	}

	return nil
}

func Int(v int) *int          { return &v }
func String(v string) *string { return &v }
