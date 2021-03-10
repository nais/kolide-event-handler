package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
)

func loadTLS(certPath, keyPath, caPath string) (credentials.TransportCredentials, error) {
	certificate, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return nil, fmt.Errorf("loading server certification: %w", err)
	}

	data, err := ioutil.ReadFile(caPath)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	caPool := x509.NewCertPool()
	if !caPool.AppendCertsFromPEM(data) {
		return nil, fmt.Errorf("can't add ca cert")
	}

	tlsConfig := &tls.Config{
		ClientAuth:   tls.RequireAndVerifyClientCert,
		Certificates: []tls.Certificate{certificate},
		ClientCAs:    caPool,
	}

	return credentials.NewTLS(tlsConfig), nil
}

func LogClientCertSubj(ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	// get client tls info
	if p, ok := peer.FromContext(ctx); ok {
		if mtls, ok := p.AuthInfo.(credentials.TLSInfo); ok {
			for _, item := range mtls.State.PeerCertificates {
				log.Infof("request certificate subject: %v", item.Subject)
			}
		} else {
			log.Infof("not ok")
		}
	}

	return handler(ctx, req)
}

func StreamLogClientCertSubj(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	// get client tls info
	if p, ok := peer.FromContext(ss.Context()); ok {
		if mtls, ok := p.AuthInfo.(credentials.TLSInfo); ok {
			for _, item := range mtls.State.PeerCertificates {
				log.Infof("request certificate subject: %v", item.Subject)
			}
		} else {
			log.Infof("not ok")
		}
	}


	return handler(srv, ss)
}
