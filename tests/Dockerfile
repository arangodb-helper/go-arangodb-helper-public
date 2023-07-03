ARG GO_VERSION
FROM golang:${GO_VERSION} AS downloader

WORKDIR /app

COPY go.mod .
COPY go.sum .
# It is done only once unless go.mod has been changed
RUN go mod download

FROM downloader AS builder

WORKDIR /app
COPY pkg /app/pkg
COPY tests /app/tests

RUN go test -c -o /usr/bin/tests /app/tests

ENTRYPOINT ["/usr/bin/tests"]
# By default test is verbose, but it can be changed when container is launched.
CMD ["-test.v"]



