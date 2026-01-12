# Changelog

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
