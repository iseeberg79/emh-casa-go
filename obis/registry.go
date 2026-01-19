// Package obis provides OBIS code constants and utilities for smart meter data.
// OBIS (Object Identification System) is a standardized system for identifying data elements
// in metering devices, allowing vendor-agnostic access to meter readings.
// See: https://www.dlms.com/
package obis

// Energy codes
const (
	// EnergyImport - Total imported energy (kWh)
	// Typically active energy consumed from the grid
	EnergyImport = "1.8.0"

	// EnergyExport - Total exported energy (kWh)
	// Typically active energy fed back to the grid (PV systems)
	EnergyExport = "2.8.0"
)

// Power codes
const (
	// PowerActive - Current active power (W)
	// Positive: consumption from grid, Negative: feed-in to grid
	PowerActive = "16.7.0"

	// PowerL1 - Phase 1 active power (W)
	PowerL1 = "36.7.0"

	// PowerL2 - Phase 2 active power (W)
	PowerL2 = "56.7.0"

	// PowerL3 - Phase 3 active power (W)
	PowerL3 = "76.7.0"
)

// Current codes (Phase currents in Amperes)
const (
	// CurrentL1 - Phase 1 current (A)
	CurrentL1 = "31.7.0"

	// CurrentL2 - Phase 2 current (A)
	CurrentL2 = "51.7.0"

	// CurrentL3 - Phase 3 current (A)
	CurrentL3 = "71.7.0"
)

// Voltage codes (Phase voltages in Volts)
const (
	// VoltageL1 - Phase 1 voltage (V)
	VoltageL1 = "32.7.0"

	// VoltageL2 - Phase 2 voltage (V)
	VoltageL2 = "52.7.0"

	// VoltageL3 - Phase 3 voltage (V)
	VoltageL3 = "72.7.0"
)

// System codes
const (
	// Frequency - Grid frequency (Hz)
	Frequency = "14.7.0"
)

// Description returns a human-readable description for a given OBIS code.
func Description(code string) string {
	descriptions := map[string]string{
		// Energy
		EnergyImport: "Total imported energy (kWh)",
		EnergyExport: "Total exported energy (kWh)",

		// Power
		PowerActive: "Current active power (W)",
		PowerL1:     "Phase 1 active power (W)",
		PowerL2:     "Phase 2 active power (W)",
		PowerL3:     "Phase 3 active power (W)",

		// Current
		CurrentL1: "Phase 1 current (A)",
		CurrentL2: "Phase 2 current (A)",
		CurrentL3: "Phase 3 current (A)",

		// Voltage
		VoltageL1: "Phase 1 voltage (V)",
		VoltageL2: "Phase 2 voltage (V)",
		VoltageL3: "Phase 3 voltage (V)",

		// System
		Frequency: "Grid frequency (Hz)",
	}

	if desc, ok := descriptions[code]; ok {
		return desc
	}
	return "Unknown OBIS code: " + code
}
