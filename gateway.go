package smgwreader

import "context"

// Gateway is the vendor-agnostic interface for smart meter gateway operations.
// Implementations (EMH CASA, Theben, PPC, etc.) must support this interface.
type Gateway interface {
	// GetReadings retrieves current meter readings with metadata.
	// Returns an Information struct containing device metadata and readings indexed by OBIS code.
	GetReadings(ctx context.Context) (*Information, error)

	// DiscoverMeterID discovers and returns the meter ID from the gateway.
	// This typically involves querying available contracts and extracting the meter identifier.
	DiscoverMeterID(ctx context.Context) (string, error)
}

// MeterProvider is an optional interface for gateways that support explicit meter ID management.
// Gateways implementing this can switch between multiple meters on the same device.
type MeterProvider interface {
	Gateway

	// SetMeterID sets the meter ID for subsequent operations.
	SetMeterID(meterID string)

	// MeterID returns the currently configured meter ID.
	MeterID() string
}

// HostConfigurer is an optional interface for gateways that support custom Host header configuration.
// This is useful for SSH tunnel scenarios where the gateway is accessed via localhost
// but expects the original hostname in the HTTP Host header.
type HostConfigurer interface {
	// SetHostHeader sets the HTTP Host header for requests to the gateway.
	// If not set, the default host from the URI is used.
	SetHostHeader(host string)
}
