package theben

import (
	"testing"

	smgwreader "github.com/iseeberg79/emh-casa-go"
)

// TestClientImplementsGateway verifies that Theben Client implements the Gateway interface.
func TestClientImplementsGateway(t *testing.T) {
	var _ smgwreader.Gateway = (*Client)(nil)
}

// TestClientImplementsMeterProvider verifies that Theben Client implements the MeterProvider interface.
func TestClientImplementsMeterProvider(t *testing.T) {
	var _ smgwreader.MeterProvider = (*Client)(nil)
}
