DIST := dist

.PHONY: build
build: clean
	mkdir -p dist/circle-ci-fast-git
	pandoc --standalone --read=markdown --write=html5 --katex --bibliography=ref.bib posts/circle-ci-fast-git.md > dist/circle-ci-fast-git/index.html
	pandoc --standalone --read=markdown --write=html5 --katex --bibliography=ref.bib posts/index.md > dist/index.html

.PHONY: clean
clean:
	rm -rf $(DIST)
