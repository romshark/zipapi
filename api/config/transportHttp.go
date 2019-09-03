package config

import (
	"crypto/tls"
	"time"

	"github.com/pkg/errors"
)

// TransportHTTPTLS represents the TLS configurations
type TransportHTTPTLS struct {
	Config              *tls.Config
	CertificateFilePath string
	PrivateKeyFilePath  string
}

// Clone creates an exact detached copy of the server TLS configurations
func (stls *TransportHTTPTLS) Clone() *TransportHTTPTLS {
	if stls == nil {
		return nil
	}
	var config *tls.Config
	if stls.Config != nil {
		config = stls.Config.Clone()
	}
	return &TransportHTTPTLS{
		Config:              config,
		CertificateFilePath: stls.CertificateFilePath,
		PrivateKeyFilePath:  stls.PrivateKeyFilePath,
	}
}

// TransportHTTP defines the HTTP server transport layer configurations
type TransportHTTP struct {
	Host              string
	KeepAliveDuration time.Duration
	TLS               *TransportHTTPTLS
}

// Init sets defaults and validates the configurations
func (conf *TransportHTTP) Init() error {
	if conf.KeepAliveDuration == time.Duration(0) {
		conf.KeepAliveDuration = 3 * time.Minute
	}

	if conf.TLS != nil {
		if conf.TLS.CertificateFilePath == "" {
			return errors.New("missing TLS certificate file path")
		}
		if conf.TLS.PrivateKeyFilePath == "" {
			return errors.New("missing TLS private key file path")
		}
	}

	return nil
}
