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

RUN go install github.com/go-delve/delve/cmd/dlv@latest

# RUN go mod download
RUN go build -gcflags="all=-N -l" -o /freedom-sentry-exe freedom-sentry

FROM alpine:3.16

COPY --from=build /freedom-sentry-exe /freedom-sentry
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /go/bin/dlv /

EXPOSE 8000 40000

CMD ["/dlv", "--listen=:40000", "--headless=true", "--api-version=2", "--accept-multiclient", "exec", "/freedom-sentry"]
