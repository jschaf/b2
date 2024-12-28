DIST_DIR := public
DEV_PORT := 8080

$(DIST_DIR):
	mkdir -p $(DIST_DIR)

.PHONY: all
all: dev

.PHONY: js
js: $(DIST_DIR)
	cp scripts/instant.page.v1.2.2.js $(DIST_DIR)

.PHONY: html
html: clean $(DIST_DIR)
	go run github.com/jschaf/b2/cmd/compile

.PHONY: clean
clean:
	rm -rf $(DIST_DIR)

.PHONY: publish
publish:
	go run ./cmd/publish

.PHONY: dev
dev:
	go run ./cmd/server --log-level=debug

.PHONY: update-katex
update-katex:
	./script/update-katex.sh

.PHONY: cert
cert:
	mkdir -p private/cert
	mkcert -cert-file=private/cert/localhost_cert.pem -key-file=private/cert/localhost_key.pem localhost
	mkcert -install
