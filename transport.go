package emhcasa

import (
	"net/http"

	"github.com/jpfielding/go-http-digest/pkg/digest"
)

// hostHeaderTransport wraps a RoundTripper and enforces a custom Host header.
// This is necessary for CASA gateways that require specific host header values
// for proper routing and validation.
type hostHeaderTransport struct {
	base http.RoundTripper
	host string
}

// RoundTrip implements http.RoundTripper, enforcing the custom host header on each request.
func (t *hostHeaderTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	// Only override host if explicitly set
	if t.host != "" {
		req.Host = t.host
		req.Header.Set("Host", t.host)
	}
	return t.base.RoundTrip(req)
}

// NewDigestTransport creates an HTTP digest authentication transport.
// It wraps the base RoundTripper with digest authentication credentials.
func NewDigestTransport(user, password string, base http.RoundTripper) http.RoundTripper {
	return digest.NewTransport(user, password, base)
}
