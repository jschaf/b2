DIST_DIR := dist
DEV_PORT := 8080

$(DIST_DIR):
	mkdir -p $(DIST_DIR)

.PHONY: js
js: $(DIST_DIR)
	cp scripts/instant.page.v1.2.2.js $(DIST_DIR)

.PHONY: html
html: clean $(DIST_DIR)
	go run github.com/jschaf/b2/cmd/compiler

.PHONY: clean
clean:
	rm -rf $(DIST_DIR)/*

.PHONY: deploy
deploy: html
	firebase deploy --only hosting:new-blog

.PHONY: dev
dev: html
	go run github.com/jschaf/b2/cmd/server
