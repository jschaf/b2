DIST := dist
DEV_PORT := 8080

.PHONY: make-dist
make-dist:
	mkdir -p $(DIST)

.PHONY: copy-scripts
copy-scripts: make-dist
	cp scripts/instant.page.v1.2.2.js $(DIST)

.PHONY: build
build: clean make-dist
	pandoc --standalone --read=markdown --write=html5 posts/index.md > $(DIST)/index.html

	mkdir -p dist/circle-ci-fast-git
	pandoc --template=templates/post.html --standalone --read=markdown --write=html5 posts/circle-ci-fast-git.md > $(DIST)/circle-ci-fast-git/index.html

	mkdir -p dist/go-server-with-syscalls
	pandoc --template=templates/post.html --standalone --read=markdown --write=html5 posts/go-server-with-syscalls.md > $(DIST)/go-server-with-syscalls/index.html

.PHONY: build-docker
build-docker:
	docker run -it --rm \
            --mount "type=bind,source=${HOME}/prog/b2,target=/home/blog-builder/b2" \
            --env "FIREBASE_TOKEN=${FIREBASE_TOKEN}" \
            --publish $(DEV_PORT):$(DEV_PORT) \
            jschaf/blog-builder:latest \
            make build

.PHONY: run-docker
run-docker:
	docker run -it --rm \
            --mount "type=bind,source=${HOME}/prog/b2,target=/home/blog-builder/b2" \
            --env "FIREBASE_TOKEN=${FIREBASE_TOKEN}" \
            --publish $(DEV_PORT):$(DEV_PORT) \
            jschaf/blog-builder:latest

.PHONY: dev-docker
dev-docker:
	docker run -it --rm \
            --mount "type=bind,source=${HOME}/prog/b2,target=/home/blog-builder/b2" \
            --env "FIREBASE_TOKEN=${FIREBASE_TOKEN}" \
            --publish $(DEV_PORT):$(DEV_PORT) \
            jschaf/blog-builder:latest \
            make dev

.PHONY: clean
clean:
	rm -rf $(DIST)/*

.PHONY: deploy
deploy: build
	firebase deploy

.PHONY: dev
dev: build
  inotifywait -e close_write,moved_to,create -m posts |
    while read -r directory events filename; do
      echo "Detected change in ${directory}, events=${events}, filename=${filename}"
      make build
    done &
	firebase serve --host 0.0.0.0 --port $(DEV_PORT)
