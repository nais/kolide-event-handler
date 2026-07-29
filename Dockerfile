ARG GO_VERSION=1.26
FROM golang:${GO_VERSION} AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o bin/kolide-event-handler .

FROM scratch
ADD https://curl.haxx.se/ca/cacert.pem /etc/ssl/certs/ca-certificates.crt
WORKDIR /app
COPY --from=builder /src/bin/kolide-event-handler /app/kolide-event-handler
ENTRYPOINT ["/app/kolide-event-handler"]
