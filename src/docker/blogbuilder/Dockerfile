FROM node:12-slim as node-builder
WORKDIR /usr/src/app
COPY package.json package-lock.json tsconfig.json ./
COPY ../.. ./server
RUN npm ci && npm run compile

FROM node:12-slim as server
WORKDIR /usr/src/app
COPY --from=node-builder /usr/src/app/out ./out
COPY --from=node-builder package.json package-lock.json ./
RUN npm ci --production
CMD ["npm", "start"]
