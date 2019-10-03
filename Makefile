DIST_DIR := dist
DEV_PORT := 8080
BUILD_DIR := build
DOCKER_DIR := docker/blog
IMAGE := jschaf/blog-builder:latest

$(DIST_DIR):
	mkdir -p $(DIST_DIR)
$(BUILD_DIR):
	mkdir -p $(BUILD_DIR)

.PHONY: js
js: $(DIST_DIR)
	cp scripts/instant.page.v1.2.2.js $(DIST_DIR)

.PHONY: html
html: clean $(DIST_DIR)
	pandoc --standalone --read=markdown --write=html5 posts/index.md > $(DIST_DIR)/index.html

	mkdir -p dist/circle-ci-fast-git
	pandoc --template=templates/post.html --standalone --read=markdown --write=html5 posts/circle-ci-fast-git.md > $(DIST_DIR)/circle-ci-fast-git/index.html

	mkdir -p dist/go-server-with-syscalls
	pandoc --template=templates/post.html --standalone --read=markdown --write=html5 posts/go-server-with-syscalls.md > $(DIST_DIR)/go-server-with-syscalls/index.html

# Builds the docker image used to build the blog.
.PHONY: docker-image
docker-image: $(BUILD_DIR)/.docker-blog

# An empty target so Make will only run docker build if the contents change.
$(BUILD_DIR)/.docker-blog: $(BUILD_DIR) $(wildcard $(DOCKER_DIR)/*)
	docker build $(DOCKER_DIR) -t jschaf/blog-builder && touch $(BUILD_DIR)/.docker-blog

.PHONY: push-docker-image
push-docker-image: docker-image
	docker push $(IMAGE)

.PHONY: site
site: docker-image
	docker run -it --rm \
            --mount "type=bind,source=${HOME}/prog/b2,target=/home/blog-builder/b2" \
            --env "FIREBASE_TOKEN=${FIREBASE_TOKEN}" \
            --publish $(DEV_PORT):$(DEV_PORT) \
            $(IMAGE) \
            make html

.PHONY: run-docker
run-docker: docker-image
	docker run -it --rm \
            --mount "type=bind,source=${HOME}/prog/b2,target=/home/blog-builder/b2" \
            --env "FIREBASE_TOKEN=${FIREBASE_TOKEN}" \
            --publish $(DEV_PORT):$(DEV_PORT) \
            $(IMAGE)

.PHONY: dev
dev: docker-image
	docker run -it --rm \
            --mount "type=bind,source=${HOME}/prog/b2,target=/home/blog-builder/b2" \
            --env "FIREBASE_TOKEN=${FIREBASE_TOKEN}" \
            --publish $(DEV_PORT):$(DEV_PORT) \
            $(IMAGE) \
            watch_and_rebuild.sh $(DEV_PORT)

.PHONY: clean
clean:
	rm -rf $(DIST_DIR)/*

.PHONY: deploy
deploy: html
	firebase deploy

.PHONY: watch
watch: html
	build/rebuild_on_change.sh &
	firebase serve --host 0.0.0.0 --port $(DEV_PORT)
