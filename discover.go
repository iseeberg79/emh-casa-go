package emhcasa

import (
	"fmt"

	"github.com/tobima/smgw-discover-go/smgw"
)

// DiscoverGatewayURI discovers the CASA gateway via mDNS by querying for "smgw.local".
// Returns a fully-formed URI (e.g., "https://[fe80::dead:beef%eth0]") ready for use.
// Uses the smgw-discover-go module which implements a 300ms timeout.
// Returns an error if no gateway is found.
func DiscoverGatewayURI() (string, error) {
	// Use existing smgw-discover-go module
	host, err := smgw.Discover()
	if err != nil {
		return "", fmt.Errorf("failed to discover gateway: %w", err)
	}

	// The smgw.Discover() already returns the host in the correct format:
	// - IPv6: [fe80::dead:beef:cafe:babe%eth1]
	// - IPv4: 192.168.1.100
	// Just prepend the HTTPS scheme
	return fmt.Sprintf("https://%s", host), nil
}
