# gofalcon

CrowdStrike Falcon API client in Go. This library mainly supports OAuth2, Event Stream API and API in `api.crowdstrike.com`.

## Getting Started

### Rerieve API user and token

Go to https://falcon.crowdstrike.com/support/api-clients-and-keys and create a new token by `Add new API Client`. See [document](https://falcon.crowdstrike.com/support/documentation/46/crowdstrike-oauth2-based-apis#api-clients) for more detail about API client.

Then save `client ID` (e.g. `1fbcxxxxxxxxxxxxxxxxxxxxxxxxxx`) and `client secret` (e.g. `o8eC9qXxXxXXXXXxxxxxxXXXXXxxxxxxXXXXX`)

### Example to get detections

Assume `Client ID` and `Client secret` are set to `FALCON_CLIENT_ID` and `FALCON_SECRET` of environment variables.

```go
package main

import (
	"bytes"
	"encoding/json"
	"log"
	"os"

	"github.com/k0kubun/pp"
	"github.com/m-mizutani/gofalcon"
)

func main() {
	falconClientID := os.Getenv("FALCON_CLIENT_ID")
	falconSecret := os.Getenv("FALCON_SECRET")

	// New client and authentication
	client := gofalcon.NewClient()
	err := client.EnableOAuth2(falconClientID, falconSecret)
	if err != nil {
		log.Fatal("Fail oauth2: ", err)
	}

	// Get list of detection IDs
	queryrReq := gofalcon.Request{
		Method: "GET",
		Path:   "/detects/queries/detects/v1",
	}

	var queryResp gofalcon.Response
	if err := client.SendRequest(queryrReq, &queryResp); err != nil {
		log.Fatal("Fail request: ", err)
	}

	// Get summaries of detections
	var body struct {
		IDs []string `json:"ids"`
	}
	for _, resource := range queryResp.Resources {
		body.IDs = append(body.IDs, resource.(string))
	}
	rawBody, err := json.Marshal(body)
	if err != nil {
		log.Fatal("Fail marshal: ", err)
	}

	summaryReq := gofalcon.Request{
		Method: "POST",
		Path:   "/detects/entities/summaries/GET/v1",
		Body:   bytes.NewReader(rawBody),
	}

	var summaryResp gofalcon.Response
	if err := client.SendRequest(summaryReq, &summaryResp); err != nil {
		log.Fatal("Fail request: ", err)
	}

	for _, resource := range summaryResp.Resources {
		pp.Println(resource)
	}
}
```

The example code is in [examples directory](`./examples/list-detects`). Then, you can run it.

```bash
$ env FALCON_CLIENT_ID=aaaaaaaa FALCON_SECRET=bbbbbbbb go run ./examples/list-detects
map[string]interface {}{
  "cid":           "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
  "detection_id":  "ldt:xxxxxxxxxxxxxxxxxxxxxxxxxxxx:0000000000000000",
  "last_behavior": "2020-11-05T04:10:45Z",
  "status":        "new",

(snip)
```

See [swagger](https://assets.falcon.crowdstrike.com/support/api/swagger.html) page for more API details.

- [QueryDetects](https://assets.falcon.crowdstrike.com/support/api/swagger.html#/detects/QueryDetects)
- [GetDetectSummaries](https://assets.falcon.crowdstrike.com/support/api/swagger.html#/detects/GetDetectSummaries)

# License

MIT License
