DIST := dist

.PHONY: make-dist
make-dist:
	mkdir -p dist

.PHONY: copy-scripts
copy-scripts: make-dist
	cp scripts/instant.page.v1.2.2.js dist/

.PHONY: build
build: clean make-dist
	pandoc --standalone --read=markdown --write=html5 posts/index.md > dist/index.html

	mkdir -p dist/circle-ci-fast-git
	pandoc --template=templates/post.html --standalone --read=markdown --write=html5 posts/circle-ci-fast-git.md > dist/circle-ci-fast-git/index.html

	mkdir -p dist/go-server-with-syscalls
	pandoc --template=templates/post.html --standalone --read=markdown --write=html5 posts/go-server-with-syscalls.md > dist/go-server-with-syscalls/index.html

.PHONY: clean
clean:
	rm -rf $(DIST)

.PHONY: deploy
deploy: build
	firebase deploy

.PHONY: dev
dev: build
	firebase serve
