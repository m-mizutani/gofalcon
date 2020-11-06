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
