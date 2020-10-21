# Use the offical golang image to create a binary.
# This is based on Debian and sets the GOPATH to /go.
# https://hub.docker.com/_/golang
FROM golang:1.15-buster as builder
# Create and change to the app directory.
WORKDIR /app
# Copy local code to the container image.
COPY track.go ./
# Build the binary. Disable CGO to avoid linking glibc.
RUN CGO_ENABLED=0 go build -v -o server track.go

FROM gcr.io/distroless/static
# Copy the binary to the production image from the builder stage.
COPY --from=builder /app/server /app/server
# Run the web service on container startup.
CMD ["/app/server"]