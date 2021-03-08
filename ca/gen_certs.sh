#!/usr/bin/env bash

# Create CA
openssl req \
  -new \
  -x509 \
  -nodes \
  -days 3650 \
  -subj '/CN=naisdevice kolide-event-handler grpc ca' \
  -keyout ca.key \
  -out ca.crt

# Server
openssl genrsa \
  -out server.key 2048

openssl req \
  -new \
  -key server.key \
  -subj '/CN=naisdevice-kolide-event-handler.nais.io' \
  -out server.csr

openssl x509 \
  -req \
  -in server.csr \
  -CA ca.crt \
  -CAkey ca.key \
  -CAcreateserial \
  -days 3650 \
  -out server.crt

# Client
openssl genrsa \
  -out client.key 2048

openssl req \
  -new \
  -key client.key \
  -subj '/CN=naisdevice-apiserver.nais.io' \
  -out client.csr

openssl x509 \
  -req \
  -in client.csr \
  -CA ca.crt \
  -CAkey ca.key \
  -CAcreateserial \
  -days 3650 \
  -out client.crt
