ARG DEBIAN_VERSION=sid-20190812-slim
ARG NODE_VERSION=10.15.3

# Builder image for node-prune.
FROM golang:1.12.9 as builder-node-prune
RUN go get github.com/tj/node-prune/cmd/node-prune

# Builder image for NPM and node.
FROM node:${NODE_VERSION} as builder-node
COPY --from=builder-node-prune /go/bin/node-prune /usr/bin/
ARG NODE_MODULES=/usr/local/lib/node_modules
RUN set -eux \
  # Remove unnecessary cruft from NPM's node_modules.
  && rm -rf ${NODE_MODULES}/npm/{man,html,doc,changelogs} \
  && find ${NODE_MODULES} -type f -name '*.min.js' -delete \
  && node-prune ${NODE_MODULES}

FROM debian:${DEBIAN_VERSION} as blog-builder
ARG DEBIAN_FRONTEND=noninteractive
COPY --from=builder-node /usr/local/lib/node_modules /usr/local/lib/node_modules
COPY --from=builder-node /usr/local/bin/node /usr/local/bin/
# Link NPM so it's available on the path.
RUN ln -s /usr/local/lib/node_modules/npm/bin/npm-cli.js /usr/local/bin/npm
RUN set -eux \
  && apt-get update \
  && apt-get install -y --no-install-recommends \
    inotify-tools \
    make \
    pandoc \
    pandoc-citeproc
RUN set -eux \
  && npm install -g firebase-tools katex
WORKDIR /home/blog-builder/b2
