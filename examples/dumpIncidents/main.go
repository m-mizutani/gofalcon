package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/k0kubun/pp"
	"github.com/m-mizutani/gofalcon"
)

func main() {
	falconClientID := os.Getenv("FALCON_CLIENT_ID")
	falconSecret := os.Getenv("FALCON_SECRET")

	if len(os.Args) < 2 {
		fmt.Println("usage) dumpIncident [incidentID]")
		os.Exit(1)
	}

	// Sample: inc:62e9c3d557a5479258d9ac63a2efb118:131b500232ee4d9ca10c70d81eceb5dc
	incidentID := os.Args[1]

	client := gofalcon.NewClient()
	err := client.EnableOAuth2(falconClientID, falconSecret)
	if err != nil {
		log.Fatal("Fail oauth2", err)
	}

	var incBody, bhvBody struct {
		IDs []string `json:"ids"`
	}

	// ---------- Incident -----------
	incBody.IDs = []string{incidentID}
	raw, err := json.Marshal(incBody)
	if err != nil {
		log.Fatal("Fail to marshal body", err)
	}

	incRequest := gofalcon.Request{
		Method: "POST",
		Path:   "/incidents/entities/incidents/GET/v1",
		Body:   bytes.NewReader(raw),
	}

	var incident gofalcon.Response
	if err := client.SendRequest(incRequest, &incident); err != nil {
		log.Fatal("Fail request to incidents/entities/incidents/GET/v1: ", err)
	}

	fmt.Println("------- Incident entry ----------")
	pp.Println(incident)

	// -------- Behavior ------------
	var behaviour gofalcon.Response
	qs := url.Values{}
	qs.Add("filter", fmt.Sprintf("incident_id:\"%s\"", incidentID))
	queryRequest := gofalcon.Request{
		Method:      "GET",
		Path:        "/incidents/queries/behaviors/v1",
		QueryString: qs,
	}
	if err := client.SendRequest(queryRequest, &behaviour); err != nil {
		log.Fatal("Fail request to incidents/queries/behaviors/v1: ", err)
	}
	fmt.Println("------- Behavior ----------")
	pp.Println(behaviour)

	// -------- Behaviors ------------
	for _, v := range behaviour.Resources {
		bhvBody.IDs = append(bhvBody.IDs, v.(string))
	}
	raw, err = json.Marshal(bhvBody)
	if err != nil {
		log.Fatal("Fail to marshal body", err)
	}

	bhvsRequest := gofalcon.Request{
		Method: "POST",
		Path:   "/incidents/entities/behaviors/GET/v1",
		Body:   bytes.NewReader(raw),
	}

	var behaviours gofalcon.Response
	if err := client.SendRequest(bhvsRequest, &behaviours); err != nil {
		log.Fatal("Fail request to incidents/entities/behaviors/GET/v1: ", err)
	}

	fmt.Println("------- Behaviors ----------")
	pp.Println(behaviours)
}
