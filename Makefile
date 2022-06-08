DIST_DIR := public
DEV_PORT := 8080

$(DIST_DIR):
	mkdir -p $(DIST_DIR)

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
dev: html
	go run ./cmd/server
