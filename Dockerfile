FROM golang:1.22 as builder
ENV GOOS=linux
ENV CGO_ENABLED=0
ENV GO111MODULE=on
COPY . /src
WORKDIR /src
RUN go mod download
RUN make test
RUN make check
RUN go build -a -installsuffix cgo -o kolide-event-handler ./cmd/kolide-event-handler

FROM alpine:3.18
LABEL org.opencontainers.image.source https://github.com/nais/kolide-event-handler
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /src/kolide-event-handler /app/kolide-event-handler
CMD ["/app/kolide-event-handler"]
