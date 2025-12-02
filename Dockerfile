ARG GO_VERSION=1.25
FROM golang:${GO_VERSION} AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o bin/kolide-event-handler .

FROM gcr.io/distroless/base
WORKDIR /app
COPY --from=builder /src/bin/kolide-event-handler /app/kolide-event-handler
ENTRYPOINT ["/app/kolide-event-handler"]
