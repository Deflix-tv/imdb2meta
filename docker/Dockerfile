FROM golang:1.15-alpine as builder

WORKDIR /go/src/app/

# Fill cache with source code of dependencies
COPY go.mod .
RUN go mod download -x

COPY . .
RUN CGO_ENABLED=0 go build -v -ldflags="-s -w" ./cmd/imdb2meta-service/

FROM gcr.io/distroless/static

COPY --from=builder /go/src/app/imdb2meta-service /

# Mount the BadgerDB or bbolt files into this directory and use that when running the container.
VOLUME [ "/data" ]
# HTTP
EXPOSE 8080
# gRPC
EXPOSE 8081

# Using ENTRYPOINT instead of CMD allows the user to easily just *add* command line arguments when using `docker run`.
#
# Default bind addr is localhost, which wouldn't allow connections from outside the container.
# Should be overwritten when using `--network host` and not wanting to expose the service to other hosts.
ENTRYPOINT ["/imdb2meta-service", "-bindAddr", "0.0.0.0"]
