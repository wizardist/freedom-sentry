FROM golang:1.19.0 AS build

WORKDIR /freedom-sentry

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN go build -ldflags="-s -w" -o /freedom-sentry-exe freedom-sentry

FROM scratch

COPY --from=build /freedom-sentry-exe /freedom-sentry
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

ENTRYPOINT ["/freedom-sentry"]
