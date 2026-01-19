package emhcasa

import (
	"testing"

	smgwreader "github.com/iseeberg79/emh-casa-go"
)

// TestClientImplementsGateway verifies that Client implements the Gateway interface.
func TestClientImplementsGateway(t *testing.T) {
	var _ smgwreader.Gateway = (*Client)(nil)
}

// TestClientImplementsMeterProvider verifies that Client implements the MeterProvider interface.
func TestClientImplementsMeterProvider(t *testing.T) {
	var _ smgwreader.MeterProvider = (*Client)(nil)
}

// TestClientImplementsHostConfigurer verifies that Client implements the HostConfigurer interface.
func TestClientImplementsHostConfigurer(t *testing.T) {
	var _ smgwreader.HostConfigurer = (*Client)(nil)
}
