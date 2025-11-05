FROM golang:1.25 AS builder
WORKDIR /src
COPY go.mod go.sum .
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o kolide-event-handler ./cmd/kolide-event-handler

FROM scratch
WORKDIR /

ADD https://curl.haxx.se/ca/cacert.pem /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /src/kolide-event-handler .

ENTRYPOINT ["/kolide-event-handler"]
