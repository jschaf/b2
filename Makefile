DIST_DIR := public

$(DIST_DIR):
	mkdir -p $(DIST_DIR)

.PHONY: all
all: dev

.PHONY: html
html: clean $(DIST_DIR)
	go run ./cmd/compile

.PHONY: clean
clean:
	rm -rf $(DIST_DIR)

.PHONY: publish
publish: html
	go run ./cmd/publish

# Run the dev server.
.PHONY: dev
dev:
	go run ./cmd/server --log-level=debug

.PHONY: update-katex
update-katex:
	./script/update-katex.sh

# Install certs for localhost.
.PHONY: cert
cert:
	mkdir -p private/cert
	mkcert -cert-file=private/cert/localhost_cert.pem -key-file=private/cert/localhost_key.pem localhost
	mkcert -install

# Build the track server and docker image.
# https://console.cloud.google.com/artifacts/docker/jschaf/us-west2/jsc-art-uswe2-docker?authuser=2&inv=1&invt=AblqCA&project=jschaf
.PHONY: track
track:
	GOOS=linux GOARCH=amd64 go build -o cmd/track/server -ldflags '-s -w' -trimpath ./cmd/track
	docker build --platform linux/arm64 -t track -f ./cmd/track/Dockerfile ./cmd/track
	docker tag track:latest us-west2-docker.pkg.dev/jschaf/jsc-art-uswe2-docker/track_server:latest
	docker push us-west2-docker.pkg.dev/jschaf/jsc-art-uswe2-docker/track_server:latest
	gcloud run deploy track-server --image=us-west2-docker.pkg.dev/jschaf/jsc-art-uswe2-docker/track_server:latest --region=us-west2 --platform=managed --allow-unauthenticated --port=3355

# Run Terraform apply.
.PHONY: terraform
terraform:
	terraform -chdir=infra/gcp apply
