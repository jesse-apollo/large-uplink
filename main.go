package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
)

func uplinkHandler(w http.ResponseWriter, req *http.Request) {

	bucket := os.Getenv("GCS_BUCKET")

	var requestQuery GQLQuery

	// Try to decode the request body into the struct. If there is an error,
	// respond to the client with the error message and a 400 status code.
	err := json.NewDecoder(req.Body).Decode(&requestQuery)
	if err != nil {
		log.Errorf("Could not decode uplink query: %s", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	apiKey := requestQuery.Variables["apiKey"].(string)
	graphRef := requestQuery.Variables["ref"].(string)
	graphRefParts := strings.Split(graphRef, "@")

	log.Debug("Uplink ref is: ", graphRef)

	if len(graphRefParts) != 2 {
		log.Errorf("Could not decode graph ref: %s", graphRef)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Object name for GCS
	objectName := fmt.Sprintf("%s.graphql", graphRef)
	signature := generateMAC([]byte(objectName), []byte(apiKey))

	var supergraphResult *SupergraphResult

	// check GCS cache first
	supergraphData, compositionID, err := downloadGCSFile(bucket, objectName, signature)
	if err != nil {
		log.Debugf("Could not load schema from GCS: %s", err)

		supergraphResult, err = downloadSupergraph(graphRefParts[0], graphRefParts[1], apiKey)

		if err == nil {
			// Cache supergraph to GCS
			reader := strings.NewReader(supergraphResult.Data.Service.SchemaTag.CompositionResult.SupergraphSDL)
			err := uploadGCSFile(
				reader,
				bucket,
				objectName,
				supergraphResult.Data.Service.SchemaTag.CompositionResult.GraphCompositionID,
				signature,
			)
			if err != nil {
				log.Errorf("Could not cache supergraph to GCS: %s", err)
			}
		} else {
			log.Errorf("Cannot download supergraph from Apollo: %s", err)
			http.Error(w, "Cannot download supergraph.", http.StatusNotFound)
			return
		}
	} else {
		supergraphResult = &SupergraphResult{
			Data: SupergraphFetch{
				Service: ServiceResult{
					SchemaTag: SchemaTag{
						CompositionResult: CompositionResult{
							SupergraphSDL:      string(supergraphData),
							GraphCompositionID: compositionID,
						},
					},
				},
			},
		}
	}

	uplinkResponse := UplinkResult{
		Data: UplinkRouterConfigWrapper{
			RouterConfig: UplinkRouterConfig{
				TypeName:      "RouterConfigResult",
				ID:            supergraphResult.Data.Service.SchemaTag.CompositionResult.GraphCompositionID,
				SupergraphSDL: supergraphResult.Data.Service.SchemaTag.CompositionResult.SupergraphSDL,
			},
		},
	}

	// write http response
	json.NewEncoder(w).Encode(uplinkResponse)
}

// webhookHandler - write incoming supergraph to GCS
func webhookHandler(w http.ResponseWriter, req *http.Request) {

	bucket := os.Getenv("GCS_BUCKET")

	log.Debug("Webhook handler")

	var buildStatus BuildStatusWebhook

	// Try to decode the request body into the struct. If there is an error,
	// respond to the client with the error message and a 400 status code.
	err := json.NewDecoder(req.Body).Decode(&buildStatus)
	if err != nil {
		log.Errorf("Could not decode build status webhook: %s", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if !buildStatus.BuildSucceeded {
		log.Errorf("Build failure, not updating supergraph.")
		return
	}

	// Object name for GCS
	objectName := fmt.Sprintf("%s.graphql", buildStatus.VariantID)

	log.Debugf("Deleting GCS file: %s (%s)", buildStatus.GraphID, buildStatus.VariantID)

	// clear cache (delete GCS storage copy of supergraph)
	err = deleteGCSFile(bucket, objectName)
	if err != nil {
		log.Errorf("Cannot delete GCS Supergraph cache: %s", err)
	}

	fmt.Fprintf(w, "OK")
}

func init() {
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)
}

func main() {
	// proxy testing only
	//http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/uplink", uplinkHandler)
	http.HandleFunc("/apollo_webhook", webhookHandler)

	fmt.Printf("Large Uplink server starting...\n")
	http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
}
