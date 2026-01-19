# emh-casa-go

A vendor-agnostic smart meter gateway library for Go with support for multiple SMGW vendors: **EMH CASA 1.1**, **PPC**, and **Theben Conexa**.

This library provides a clean, unified interface for querying meter data from smart meter gateways, with standardized OBIS code handling, automatic discovery, and rich data structures.

## Features (v2.0.0)

- **Vendor-agnostic Gateway interface**: Unified API for multiple SMGW vendors (EMH CASA 1.1, PPC, Theben Conexa)
- **Standardized OBIS codes**: "16.7.0" (current power) works consistently across all vendors
- **Rich data structures**: Information and Reading objects with timestamps, metadata, and quality indicators
- **Gateway Auto-discovery**: Automatically discovers gateways via mDNS ("smgw.local")
- **Meter ID Auto-discovery**: Automatically discovers meter IDs from available contracts
- **HTTP Digest Authentication**: Secure communication with gateways
- **Unit Handling**: Automatic scaling and unit conversion (W, kWh, A, V, Hz)
- **Self-signed Certificates**: Works with typical gateway configurations
- **Context support**: Full context support for cancellation and timeouts
- **Backward compatible**: Optional GetMeterValues() method for v1 compatibility

## Installation

```bash
go get github.com/iseeberg79/emh-casa-go
```

## Automatic Gateway Discovery

The library supports mDNS-based gateway discovery for networks where the gateway advertises itself as "smgw.local":

```go
import "github.com/iseeberg79/emh-casa-go"

// Full auto-discovery with just credentials
client, err := emhcasa.NewClientDiscover("admin", "password")
if err != nil {
    log.Fatal(err)
}

// Query meter values
values, err := client.GetMeterValues()
```

**Discovery behavior**:
- Uses the proven `smgw-discover-go` module (tested with EMH CASA 1.1)
- 300ms timeout for mDNS queries
- Queries for "smgw.local" hostname
- Works with IPv6 link-local addresses
- Preserves network interface zone identifiers (e.g., `%eth1`)

**Troubleshooting discovery**:
- Ensure gateway is on the same network subnet
- Verify gateway advertises "smgw.local" via mDNS (EMH CASA 1.1 does this by default)
- Check that multicast DNS is enabled on your network interface
- IGMP snooping could block mDNS. Check if IGMP snooping is disabled on the network switch or relevant VLAN. This does only applies to managed switches.
- Some networks may block mDNS traffic - in this case, provide the URI manually

**Manual configuration** (if discovery is not available):
```go
client, err := emhcasa.NewClient(
    "https://192.168.33.2",  // Gateway IP address
    "admin",
    "password",
    "",  // auto-discover meter ID
)
```

**SSH Tunneling**:

When using SSH tunnels, set the Host header after creating the client:

```go
client, err := emhcasa.NewClient(
    "https://localhost:8443",
    "admin",
    "password",
    "",  // auto-discover meter ID
)
if err != nil {
    log.Fatal(err)
}

// Set Host header for gateway routing
client.SetHostHeader("smgw.local")  // or "192.168.33.2"

values, err := client.GetMeterValues()
```

## Quick Start (v2.0.0)

```go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	smgwreader "github.com/iseeberg79/emh-casa-go"
	"github.com/iseeberg79/emh-casa-go/emhcasa"
	"github.com/iseeberg79/emh-casa-go/obis"
)

func main() {
	// Create EMH CASA client
	client, err := emhcasa.NewClient(
		"https://192.168.33.2",  // Gateway URI or empty for auto-discover
		"admin",                 // Username
		"password",              // Password
		"",                      // Meter ID (empty to auto-discover)
	)
	if err != nil {
		log.Fatal(err)
	}

	// Use Gateway interface for vendor-agnostic access
	var gw smgwreader.Gateway = client

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Fetch readings with rich metadata
	info, err := gw.GetReadings(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// Access data using OBIS constants
	fmt.Printf("Gateway: %s (%s)\n", info.Name, info.Manufacturer)
	fmt.Printf("Last Updated: %s\n", info.LastUpdate.Format(time.RFC3339))

	// Current power (OBIS 16.7.0)
	if reading, ok := info.Readings[obis.PowerActive]; ok {
		fmt.Printf("Current Power: %.2f %s\n", reading.Value, reading.Unit)
	}

	// Total energy (OBIS 1.8.0)
	if reading, ok := info.Readings[obis.EnergyImport]; ok {
		fmt.Printf("Total Energy: %.2f %s\n", reading.Value, reading.Unit)
	}

	// Phase currents
	for phase, obisCode := range map[int]string{1: obis.CurrentL1, 2: obis.CurrentL2, 3: obis.CurrentL3} {
		if reading, ok := info.Readings[obisCode]; ok {
			fmt.Printf("Phase %d Current: %.2f %s\n", phase, reading.Value, reading.Unit)
		}
	}

	// Phase voltages
	for phase, obisCode := range map[int]string{1: obis.VoltageL1, 2: obis.VoltageL2, 3: obis.VoltageL3} {
		if reading, ok := info.Readings[obisCode]; ok {
			fmt.Printf("Phase %d Voltage: %.2f %s\n", phase, reading.Value, reading.Unit)
		}
	}
}
```

## Vendor-Specific Examples

### PPC SMGW

```go
import (
	"context"
	"time"

	smgwreader "github.com/iseeberg79/emh-casa-go"
	"github.com/iseeberg79/emh-casa-go/ppc"
	"github.com/iseeberg79/emh-casa-go/obis"
)

// Create PPC client
client, err := ppc.NewClient("https://192.168.1.100", "admin", "password")
if err != nil {
	log.Fatal(err)
}

// Use Gateway interface
var gw smgwreader.Gateway = client

ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

// Fetch readings
info, err := gw.GetReadings(ctx)
if err != nil {
	log.Fatal(err)
}

// Access OBIS readings (PPC dynamically extracts all available OBIS codes)
fmt.Printf("Power: %.2f %s\n", info.Readings[obis.PowerActive].Value, info.Readings[obis.PowerActive].Unit)
fmt.Printf("Energy Import: %.2f %s\n", info.Readings[obis.EnergyImport].Value, info.Readings[obis.EnergyImport].Unit)
```

**PPC Notes:**
- Uses HTML scraping with BeautifulSoup-like parsing
- Self-signed certificates supported
- Dynamically extracts all OBIS codes from device (including 16.7.0 if available)

### Theben Conexa

```go
import (
	"context"
	"time"

	smgwreader "github.com/iseeberg79/emh-casa-go"
	"github.com/iseeberg79/emh-casa-go/theben"
	"github.com/iseeberg79/emh-casa-go/obis"
)

// Create Theben client
client, err := theben.NewClient("https://192.168.1.100", "admin", "password")
if err != nil {
	log.Fatal(err)
}

// Use Gateway interface
var gw smgwreader.Gateway = client

ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

// Fetch readings
info, err := gw.GetReadings(ctx)
if err != nil {
	log.Fatal(err)
}

// Access OBIS readings (Theben supports specific codes: 1.8.0, 2.8.0, 16.7.0)
fmt.Printf("Power: %.2f %s\n", info.Readings[obis.PowerActive].Value, info.Readings[obis.PowerActive].Unit)
fmt.Printf("Energy Import: %.2f %s\n", info.Readings[obis.EnergyImport].Value, info.Readings[obis.EnergyImport].Unit)
fmt.Printf("Energy Export: %.2f %s\n", info.Readings[obis.EnergyExport].Value, info.Readings[obis.EnergyExport].Unit)
```

**Theben Notes:**
- Uses JSON API
- Self-signed certificates supported
- OBIS 16.7.0 (current power) newly added
- Supports OBIS codes: 1.8.0, 2.8.0, 16.7.0, 31.7.0, 32.7.0, 36.7.0, 51.7.0, 52.7.0, 56.7.0, 71.7.0, 72.7.0, 76.7.0, 14.7.0

## Migration Guide (v1 → v2.0.0)

### Breaking Changes

v2.0.0 is a **major version** with breaking API changes. v1.x code will not work with v2.0.0.

| v1.x | v2.0.0 | Notes |
|------|--------|-------|
| `GetMeterValues()` | `GetReadings(ctx)` | Returns `*Information` instead of `map[string]float64`. `GetMeterValues()` is still available for backward compatibility. |
| `DiscoverMeterID()` | `DiscoverMeterID(ctx)` | Now requires context parameter and returns `(string, error)` |
| `MeterID()` | `MeterID()` or `MeterProvider.SetMeterID()` | Behavior unchanged, optional. Use MeterProvider interface for vendor-agnostic access. |
| `SetHostHeader()` | `HostConfigurer.SetHostHeader()` | No change, available via optional interface |
| Map access: `values["16.7.0"]` | `info.Readings[obis.PowerActive]` | Use OBIS constants from `obis` package for standardized access |
| No unit info | `reading.Unit` | Readings now include unit type and enum values |
| No timestamps | `reading.Timestamp` | Readings include capture timestamp |

### Quick Migration Example

**v1.x:**
```go
client, _ := emhcasa.NewClient(uri, user, pass, "")
values, _ := client.GetMeterValues()
power := values["16.7.0"]
```

**v2.0.0:**
```go
client, _ := emhcasa.NewClient(uri, user, pass, "")
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

info, _ := client.GetReadings(ctx)
power := info.Readings[obis.PowerActive].Value
```

## API Overview (v2.0.0)

### Gateway Interface (recommended)

```go
import (
	"context"
	smgwreader "github.com/iseeberg79/emh-casa-go"  // core interfaces
	"github.com/iseeberg79/emh-casa-go/emhcasa"     // implementation
)

// Create a gateway implementation
var gw smgwreader.Gateway
client, err := emhcasa.NewClient(uri, user, pass, meterID)
gw = client

// Fetch readings (vendor-agnostic)
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

info, err := gw.GetReadings(ctx)
meter, err := gw.DiscoverMeterID(ctx)
```

### Optional Interfaces

```go
// MeterProvider - for explicit meter ID management
// Supported by: EMH CASA, PPC, Theben (all vendors)
if mp, ok := gw.(smgwreader.MeterProvider); ok {
	mp.SetMeterID("ABC123")  // Set meter ID explicitly
	id := mp.MeterID()       // Get current meter ID
}

// HostConfigurer - for custom Host headers (SSH tunnels)
// Supported by: EMH CASA only
if hc, ok := gw.(smgwreader.HostConfigurer); ok {
	hc.SetHostHeader("smgw.local")
}
```

**Interface Support Matrix:**

| Interface | EMH CASA | PPC | Theben |
|-----------|----------|-----|--------|
| `Gateway` (required) | ✅ | ✅ | ✅ |
| `MeterProvider` | ✅ | ✅ | ✅ |
| `HostConfigurer` | ✅ | ❌ | ❌ |

### Data Structures

```go
// Information - gateway/meter information with readings
type Information struct {
	Name            string              // Gateway/meter name
	Model           string              // Device model
	Manufacturer    string              // Manufacturer
	FirmwareVersion string              // Firmware version
	LastUpdate      time.Time           // Data retrieval timestamp
	Readings        map[string]Reading  // OBIS code → Reading
}

// Reading - single meter reading with metadata
type Reading struct {
	Value     float64   // Measured value (scaled)
	Unit      Unit      // Unit type (W, A, V, Hz, kWh)
	Timestamp time.Time // When the reading was captured
	OBIS      string    // OBIS code (e.g., "16.7.0")
	Quality   Quality   // Data quality indicator
}

// (These types are in the smgwreader package)
```

### OBIS Constants

```go
import "github.com/iseeberg79/emh-casa-go/obis"

// Use standardized constants instead of magic strings
power := info.Readings[obis.PowerActive].Value     // 16.7.0
energy := info.Readings[obis.EnergyImport].Value   // 1.8.0
voltage := info.Readings[obis.VoltageL1].Value     // 32.7.0

// Get human-readable descriptions
desc := obis.Description(obis.PowerActive)  // "Current active power (W)"
```

### Backward Compatibility

For v1.x compatibility, the old `GetMeterValues()` method is still available:

```go
// Old v1.x style still works (but not recommended)
client, _ := emhcasa.NewClient(uri, user, pass, "")
values, _ := client.GetMeterValues()  // Returns map[string]float64
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

For SSH tunnels or when the gateway requires a specific host header, use `SetHostHeader()` after creating the client:
```go
client, err := emhcasa.NewClient(
	"https://localhost:8443",
	"user",
	"pass",
	"",  // auto-discover meter ID
)
if err != nil {
	log.Fatal(err)
}

// Set custom Host header for gateway routing
client.SetHostHeader("smgw.local")
```

### Meter ID Auto-discovery

If no meter ID is provided, the library automatically discovers the first available contract:

```go
// Meter ID auto-discovered
client, err := emhcasa.NewClient(uri, user, pass, "")

// Or explicitly provide it if known
client, err := emhcasa.NewClient(uri, user, pass, "ABC123...")
```

## evcc Integration

This library aims to get used by [evcc](https://evcc.io) for CASA gateway meter support:

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

## Troubleshooting

### Connection Issues

1. **Verify host header**: Most CASA gateways need the IP address as host header
2. **Check credentials**: Verify username and password are correct
3. **Self-signed certificates**: The library automatically trusts self-signed certs

### Meter Discovery Fails

- Ensure the gateway has at least one contract with sensor domains configured
- Try providing the meter ID explicitly if known

### No Values Returned

- Confirm the meter ID is correct
- Check gateway API is responding with `/json/metering/origin/{meterID}/extended`

## Disclaimer

This project is an independent, open-source library and is **not affiliated with, endorsed by, or sponsored by EMH metering GmbH** or any of its partners.  
“EMH” and “CASA” are trademarks of their respective owners and are used for descriptive purposes only.

This software is provided **“as is”**, without warranty of any kind, express or implied.  
Use of this library is **at your own risk**.

⚠️ **Note**: This library is pre-1.0.0.  
Breaking API changes may occur between minor versions. See `CHANGELOG.md`.

---

## Regulatory Notice

This library accesses data via the HAN interface of EMH CASA smart meter gateways.  
It **does not replace** certified, BSI-compliant software and **does not claim compliance** with regulatory requirements such as the German *Messstellenbetriebsgesetz (MsbG)* or BSI protection profiles.

The responsibility for compliant and lawful operation lies entirely with the user of this software.

---

## Data Protection

This library does not collect, store, or transmit data on its own.  
Any processing of metering data, which may be considered personal data under applicable laws, is the responsibility of the integrating application and its operator.

---

## License

This project is licensed under the **MIT License**. See the `LICENSE` file for details.
