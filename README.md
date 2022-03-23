# Large Supergraph Uplink

This is a substitude uplink service that can load Supergraphs over 10MB.

## Build 

To build the container run `make build`

## Deploy to Cloud Run

 1. Copy `dot_env` to `.env` and add your GCP bucket and project name.
 2. Run `make push` to build/deploy docker image to GCR.
 3. Run `make deploy` to deploy to Cloud Run
