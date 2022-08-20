package http

import (
	"crypto/x509"
	"fmt"
	"net"
	"net/http"
	"time"

	corev1 "k8s.io/api/core/v1"
)

const CACertKey = "certCA"

// DefaultTransport is a copy of the standard http DefaultTransport.
var DefaultTransport = &http.Transport{
	Proxy: http.ProxyFromEnvironment,
	DialContext: (&net.Dialer{
		// By default we wrap the transport in retries, so reduce the
		// default dial timeout to 5s to avoid 5x 30s of connection
		// timeouts when doing the "ping" on certain http registries.
		Timeout:   5 * time.Second,
		KeepAlive: 30 * time.Second,
	}).DialContext,
	ForceAttemptHTTP2:     true,
	MaxIdleConns:          100,
	IdleConnTimeout:       90 * time.Second,
	TLSHandshakeTimeout:   10 * time.Second,
	ExpectContinueTimeout: 1 * time.Second,
}

// TransportFromSecret configures an http.Transport with the TLS Config in a
// secret, if provided.
func TransportFromSecret(secret *corev1.Secret) (*http.Transport, error) {
	transport := DefaultTransport.Clone()
	clientConfig := transport.TLSClientConfig

	if caCert, ok := secret.Data[CACertKey]; ok {
		syscerts, err := x509.SystemCertPool()
		if err != nil {
			return nil, fmt.Errorf("setting up the cert pool: %w", err)
		}
		syscerts.AppendCertsFromPEM(caCert)
		clientConfig.RootCAs = syscerts
	}

	return transport, nil
}
