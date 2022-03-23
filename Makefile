include .env

build-container:
	mkdir -p build
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s" -a -installsuffix cgo -o build/large-uplink.amd64 .
	docker build . -t large-uplink

deploy: 
	gcloud run deploy large-uplink --image gcr.io/${GCS_PROJECT}/large-uplink --allow-unauthenticated \
		--region us-east1 --update-env-vars GCS_BUCKET=${GCS_BUCKET}

push: build-container
	docker tag large-uplink gcr.io/${GCS_PROJECT}/large-uplink
	docker push gcr.io/${GCS_PROJECT}/large-uplink