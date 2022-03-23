include .env

build-container:
	mkdir -p build
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s" -a -installsuffix cgo -o build/large-uplink.amd64 .
	docker build . -t large-uplink

deploy:
	echo "Running Google Cloud Build"
	@gcloud builds submit --substitutions=_GCS_BUCKET=${GCS_BUCKET}
