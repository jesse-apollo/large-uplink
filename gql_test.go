package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"testing"

	log "github.com/sirupsen/logrus"
)

func TestGQLQuery(t *testing.T) {

	graphRef := os.Getenv("APOLLO_GRAPH_REF")
	graphRefParts := strings.Split(graphRef, "@")

	var q = GQLQuery{
		Variables: map[string]interface{}{
			"graph_id": graphRefParts[0],
			"variant":  graphRefParts[1],
		},
		Query:         SupergraphQuery,
		OperationName: "SupergraphFetchQuery",
	}

	body, _ := json.Marshal(q)

	//proxyUrl, _ := url.Parse("http://localhost:8000")
	tr := &http.Transport{
		DisableKeepAlives:  true,
		DisableCompression: false,
		//TLSClientConfig:    &tls.Config{InsecureSkipVerify: true},
		//Proxy:              http.ProxyURL(proxyUrl),
	}

	httpClient := http.Client{Transport: tr}
	postRequest, err := http.NewRequest(
		"POST",
		"https://graphql.api.apollographql.com/api/graphql",
		bytes.NewBuffer(body))

	if err != nil {
		log.Errorf("Could create request %s", err)
		return
	}

	postRequest.Close = true
	postRequest.Header.Set("Accept", "*/*")
	postRequest.Header.Set("Content-Type", "application/json")
	postRequest.Header.Set("X-API-Key", os.Getenv("APOLLO_KEY"))
	postRequest.Header.Set("apollographql-client-name", "lovelace-uplink-large")
	postRequest.Header.Set("apollographql-client-version", "0.1.0")

	resp, err := httpClient.Do(postRequest)

	if err != nil {
		log.Errorf("Could not retrieve supergraph SDL %s", err)
		return
	}
	defer resp.Body.Close()

	var supergraphResult SupergraphResult

	// Decode response
	err = json.NewDecoder(resp.Body).Decode(&supergraphResult)
	if err != nil {
		log.Errorf("Could not decode supergraph result: %s", err)
		//http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	/*if ans != -2 {

		t.Errorf("IntMin(2, -2) = %d; want -2", ans)
	}*/
}
