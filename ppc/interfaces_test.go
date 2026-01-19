package ppc

import (
	"testing"

	smgwreader "github.com/iseeberg79/emh-casa-go"
)

// TestClientImplementsGateway verifies that PPC Client implements the Gateway interface.
func TestClientImplementsGateway(t *testing.T) {
	var _ smgwreader.Gateway = (*Client)(nil)
}

// TestClientImplementsMeterProvider verifies that PPC Client implements the MeterProvider interface.
func TestClientImplementsMeterProvider(t *testing.T) {
	var _ smgwreader.MeterProvider = (*Client)(nil)
}
