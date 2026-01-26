# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go client library for EMH CASA 1.1 Smart Meter Gateways. It provides HTTP digest authentication, mDNS gateway discovery, OBIS value conversion, and automatic meter ID discovery.

## Build & Test Commands

```bash
# Run all tests (unit + integration)
go test -v ./...

# Run tests with race detection and coverage
go test -v -race -coverprofile=coverage.out ./...

# Run only unit tests (skip integration tests requiring gateway)
go test -v -short ./...

# Build library
go build -v ./...

# Build example discovery tool
go build -o tmp/discover.bin ./tmp/discover/

# Format code
go fmt ./...

# Run linter
golangci-lint run --timeout 5m

# Verify go.mod is tidy
go mod tidy && git diff --exit-code go.mod go.sum
```

## Architecture

### Core Components

**Client (`client.go`)**: Main entry point. Manages HTTP client with digest auth, custom transport chain, and coordinates discovery. The `GetMeterValues()` method is the primary interface for retrieving meter data.

**Transport Layer (`transport.go`)**: Two-layer transport chain:
1. `hostHeaderTransport` - Innermost, wraps base HTTP transport, handles custom Host header (needed for SSH tunnels)
2. `digest.Transport` - Outermost, wraps hostHeaderTransport, adds HTTP digest authentication

**Discovery (`discover.go`)**: Gateway auto-discovery via mDNS using `smgw-discover-go` module. Queries "smgw.local" with 300ms timeout, handles IPv6 link-local addresses with zone identifiers.

**Types (`types.go`)**:
- `DerivedContract` - Contract metadata with sensor domains
- `MeterValue` - Individual readings with unit codes and scalers
- `MeterReading` - API response wrapper

### OBIS Conversion

CASA gateways return logical names in hex format (12 chars). The library extracts bytes 4-9 and converts to standard OBIS C.D.E format:
- Input: `"0100010700FF.255"` (hex logical name)
- Extract: bytes at positions 4-5 (C), 6-7 (D), 8-9 (E)
- Output: `"1.7.0"` (OBIS code)

### Unit Handling

The library converts units based on DLMS/COSEM unit codes:
- 27 = W (Watts) - stored as-is
- 30 = Wh (Watthours) - converted to kWh (/1000)
- 33 = A (Amperes) - stored as-is
- 35 = V (Volts) - stored as-is
- 44 = Hz (Hertz) - stored as-is

Values are scaled using: `value * 10^scaler`

### Client Initialization Flow

1. If URI empty → discover via mDNS (`discover.go`)
2. Create base transport with `InsecureSkipVerify` (self-signed certs)
3. Wrap with `hostHeaderTransport` (initially empty host)
4. Wrap with digest auth transport
5. If meterID empty → auto-discover on first `GetMeterValues()` call

### Meter ID Discovery

Queries `/json/metering/derived` for contracts, iterates until finding one with non-empty `sensor_domains`, uses first sensor domain as meter ID.

## Package Structure

This is a single-package library (`package emhcasa`) with all core code at the root. The `tmp/discover/` directory contains an example CLI tool (not part of the library).

## Integration Tests

Integration tests in `client_integration_test.go` require:
- Environment variables: `CASA_URI`, `CASA_USER`, `CASA_PASS`
- Skip with `-short` flag if no gateway available

## Common OBIS Codes

- `16.7.0` - Current power (W)
- `1.8.0` / `2.8.0` - Energy import/export (kWh)
- `31.7.0`, `51.7.0`, `71.7.0` - Phase currents (A)
- `32.7.0`, `52.7.0`, `72.7.0` - Phase voltages (V)
- `36.7.0`, `56.7.0`, `76.7.0` - Phase powers (W)

## Key Design Constraints

- HTTP/1.1 enforced (`ForceAttemptHTTP2: false`) - required for CASA gateways
- Self-signed certificate support (`InsecureSkipVerify: true`)
- Preserves IPv6 zone identifiers in URIs (e.g., `fe80::1%eth1`)
- Meter ID auto-discovery fails gracefully if no contracts have sensor domains
