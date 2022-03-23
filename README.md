# Large Supergraph Uplink

This is a substitude uplink service that can load Supergraphs over 10MB.

## Build 

To build the container run `make build`

## Deploy to Cloud Run

 1. Edit the `cloudbuild.yaml` file to change the __<CHANGE_ME>__ items to your project name.
 2. Create a `.env` file with a `GCS_BUCKET` key set to your GCS bucket name.
 3. Run `make deploy`
