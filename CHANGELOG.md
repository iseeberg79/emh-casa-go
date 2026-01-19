# Changelog

## [2.0.0] – Vendor-agnostic architecture with unified interfaces

**This is a major version release with breaking changes.**
v2.0.0 refactors emh-casa-go into a vendor-agnostic smart meter gateway library with unified interfaces and rich data structures. EMH CASA support is now in the `vendor/emhcasa` package.

### Added
- **Gateway interface**: Unified vendor-agnostic API for smart meter gateways
- **Optional interfaces**: `MeterProvider` for meter ID management, `HostConfigurer` for custom host headers
- **Rich data structures**: `Information` and `Reading` types with timestamps, metadata, and quality indicators
- **OBIS registry**: Centralized OBIS constants in `obis` package (obis.PowerActive, obis.EnergyImport, etc.)
- **Context support**: Full context support for cancellation and timeouts in all API methods
- **Unit enums**: `Unit` type with constants (UnitWatt, UnitWh, UnitAmpere, UnitVolt, UnitHertz)
- **Quality indicators**: `Quality` type for data quality tracking
- **Vendor separation**: EMH CASA implementation moved to `vendor/emhcasa/` package
- **Future extensibility**: Clean architecture for adding Theben, PPC, and other vendors

### Changed
- `GetMeterValues()` → `GetReadings(ctx context.Context) (*Information, error)`
- `DiscoverMeterID()` → `DiscoverMeterID(ctx context.Context) (string, error)` (now with context)
- OBIS code access: Use `obis.PowerActive` constant instead of magic string `"16.7.0"`
- All readings now include timestamps, unit types, and quality indicators
- Client package structure: EMH CASA implementation in `vendor/emhcasa` sub-package

### Deprecated
- None (clean break for v2.0.0)

### Removed
- None (all functionality preserved)

### Breaking Changes
- `GetMeterValues()` now returns `map[string]float64` via legacy compatibility method only. Use `GetReadings(ctx)` instead.
- `DiscoverMeterID()` now requires `context.Context` parameter
- Import path: EMH CASA client now at `github.com/iseeberg79/emh-casa-go/vendor/emhcasa`
- OBIS codes: Import and use constants from `github.com/iseeberg79/emh-casa-go/obis`

### Migration Guide
See [README.md#migration-guide-v1--v20](README.md#migration-guide-v1--v20) for detailed migration instructions.

**Old (v1.x):**
```go
client, _ := emhcasa.NewClient(uri, user, pass, "")
values, _ := client.GetMeterValues()
power := values["16.7.0"]
```

**New (v2.0.0):**
```go
client, _ := emhcasa.NewClient(uri, user, pass, "")
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
info, _ := client.GetReadings(ctx)
power := info.Readings[obis.PowerActive].Value
```

### PPC and Theben Support Added

**NEW** - PPC SMGW implementation (`ppc/` package):
- HTML scraping with HTTP Digest Auth
- Dynamic OBIS code extraction from meter data table
- Supports all OBIS codes reported by the device (including 16.7.0 for current power)
- Self-signed certificate support

**NEW** - Theben Conexa implementation (`theben/` package):
- JSON API with HTTP Digest Auth
- OBIS 16.7.0 (current active power) newly added
- Full OBIS support: 1.8.0, 2.8.0, 16.7.0, 31.7.0, 32.7.0, 36.7.0, 51.7.0, 52.7.0, 56.7.0, 71.7.0, 72.7.0, 76.7.0, 14.7.0
- Hex to OBIS code conversion (e.g., `0100100700ff` → `1-0:16.7.0`)
- Self-signed certificate support

### Future Roadmap
- Additional vendor support (community contributions welcome)

---

## [0.1.0] – API refactor and auto-discovery

This release refactors the client API to simplify configuration and
adds automatic gateway and meter discovery.

### Added
- mDNS-based gateway auto-discovery (`smgw.local`)
- `NewClientDiscover()` for zero-config client setup
- Automatic meter ID discovery when not explicitly provided
- Improved handling of IPv6 link-local addresses

### Changed
- Simplified `NewClient` constructor
- Host header configuration moved to `SetHostHeader()`
- `MeterID()` now returns `(string, error)`

### Breaking changes
- `NewClient` signature changed
  **Before**:
  ```go
  NewClient(uri, user, password, meterID, hostHeader)
  ```
  **After**:
  ```go
  NewClient(uri, user, password, meterID)
  client.SetHostHeader("smgw.local")
  ```


## [0.0.2] – Stabilization

Internal refactoring


## [0.0.1] – Initial release

Initial CASA 1.1 client implementation
- Digest authentication
- OBIS parsing and unit conversion
