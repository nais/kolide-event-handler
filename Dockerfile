ARG GO_VERSION="1.24"
FROM golang:${GO_VERSION} AS builder
WORKDIR /src
COPY . /src
RUN go mod download
RUN go build -o kolide-event-handler ./cmd/kolide-event-handler

FROM gcr.io/distroless/base
WORKDIR /app
COPY --from=builder /src/kolide-event-handler /app/kolide-event-handler
CMD ["/app/kolide-event-handler"]
