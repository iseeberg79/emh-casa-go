# emh-casa-go

A Go client library for EMH CASA 1.1 Smart Meter Gateways.

This library provides a clean, reusable interface for querying meter data from EMH CASA 1.1 smart meter gateways, handling HTTP digest authentication, custom host headers, and OBIS value parsing.

## Features

- **HTTP Digest Authentication**: Secure communication with CASA gateways
- **Auto-discovery**: Automatically discovers meter IDs from TAF-1 contracts
- **OBIS Conversion**: Converts CASA logical names to standard OBIS C.D.E format
- **Unit Handling**: Automatic scaling and unit conversion (W, Wh, A, V, Hz)
- **Self-signed Certificates**: Works with typical CASA gateway configurations
- **HTTP/1.1 Support**: Enforces HTTP/1.1 (required for CASA gateways)

## Installation

```bash
go get github.com/iseeberg79/emh-casa-go
```

## Quick Start

```go
package main

import (
	"fmt"
	"log"

	"github.com/iseeberg79/emh-casa-go"
)

func main() {
	// Create a client
	client, err := emhcasa.NewClient(
		"https://192.168.33.2",    // CASA gateway URI
		"admin",                     // Username
		"password",                  // Password
		"",                          // Meter ID (empty to auto-discover)
		"192.168.33.2",             // Host header (required for most CASA gateways)
	)
	if err != nil {
		log.Fatal(err)
	}

	// Fetch meter values
	values, err := client.GetMeterValues()
	if err != nil {
		log.Fatal(err)
	}

	// Use OBIS codes to access specific values
	if power, ok := values["16.7.0"]; ok {
		fmt.Printf("Current Power: %.2f W\n", power)
	}

	if energy, ok := values["1.8.0"]; ok {
		fmt.Printf("Total Energy: %.2f kWh\n", energy)
	}

	// Phase currents
	fmt.Printf("Phase 1 Current: %.2f A\n", values["31.7.0"])
	fmt.Printf("Phase 2 Current: %.2f A\n", values["51.7.0"])
	fmt.Printf("Phase 3 Current: %.2f A\n", values["71.7.0"])

	// Phase voltages
	fmt.Printf("Phase 1 Voltage: %.2f V\n", values["32.7.0"])
	fmt.Printf("Phase 2 Voltage: %.2f V\n", values["52.7.0"])
	fmt.Printf("Phase 3 Voltage: %.2f V\n", values["72.7.0"])
}
```

## API Overview

### Client

```go
// Create a new CASA client
client, err := emhcasa.NewClient(
	uri,      // Gateway URI (http/https)
	user,     // Username for digest auth
	password, // Password for digest auth
	meterID,  // Meter ID (empty to auto-discover)
	hostHeader, // Host header for custom routing
)

// Fetch all meter values (returns OBIS code -> value map)
values, err := client.GetMeterValues()

// Get the configured meter ID
meterID := client.MeterID()

// Auto-discover meter ID from TAF-1 contracts
err := client.DiscoverMeterID()
```

## Common OBIS Codes

| OBIS Code | Description | Unit |
|-----------|-------------|------|
| 1.8.0 | Total Energy Import | kWh |
| 2.8.0 | Total Energy Export | kWh |
| 16.7.0 | Current Power (Active) | W |
| 31.7.0 | Phase 1 Current | A |
| 32.7.0 | Phase 1 Voltage | V |
| 36.7.0 | Phase 1 Power | W |
| 51.7.0 | Phase 2 Current | A |
| 52.7.0 | Phase 2 Voltage | V |
| 56.7.0 | Phase 2 Power | W |
| 71.7.0 | Phase 3 Current | A |
| 72.7.0 | Phase 3 Voltage | V |
| 76.7.0 | Phase 3 Power | W |

## Configuration

### Host Header

Most CASA gateways require a specific host header for routing. If not provided, the library attempts to derive it from the URI. For best results, explicitly specify the gateway's IP address:

```go
client, err := emhcasa.NewClient(
	"https://casa.example.com",
	"user",
	"pass",
	"",
	"192.168.33.2", // Required for most setups
)
```

### Meter ID Auto-discovery

If no meter ID is provided, the library automatically discovers the first TAF-1 contract:

```go
// Meter ID auto-discovered
client, err := emhcasa.NewClient(uri, user, pass, "", host)

// Or explicitly provide it if known
client, err := emhcasa.NewClient(uri, user, pass, "ABC123...", host)
```

## evcc Integration

This library is used by [evcc](https://evcc.io) for CASA gateway meter support:

```go
import "github.com/iseeberg79/emh-casa-go"

// Create evcc meter wrapper
meter := &EMHCasa{
	client: casaClient,
	// ... logging and caching
}

// Implements evcc meter interfaces
power, _ := meter.CurrentPower()     // api.Meter
energy, _ := meter.TotalEnergy()     // api.MeterEnergy
l1, l2, l3, _ := meter.Currents()   // api.PhaseCurrents
```

## Attribution

Based on work by [gosanman](https://github.com/gosanman/smartmetergateway)

Original implementation: https://github.com/gosanman/smartmetergateway

## License

MIT License - See [LICENSE](LICENSE) file for details

## Troubleshooting

### Connection Issues

1. **Verify host header**: Most CASA gateways need the IP address as host header
2. **Check credentials**: Verify username and password are correct
3. **Self-signed certificates**: The library automatically trusts self-signed certs

### Meter Discovery Fails

- Ensure the gateway has at least one TAF-1 contract configured
- Try providing the meter ID explicitly if known

### No Values Returned

- Confirm the meter ID is correct
- Check gateway API is responding with `/json/metering/origin/{meterID}/extended`
