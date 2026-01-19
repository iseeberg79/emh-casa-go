package obis

import (
	"testing"
)

// TestOBISConstants verifies that OBIS constants are correctly defined.
func TestOBISConstants(t *testing.T) {
	tests := []struct {
		name  string
		code  string
		valid bool
	}{
		// Energy
		{name: "EnergyImport", code: EnergyImport, valid: true},
		{name: "EnergyExport", code: EnergyExport, valid: true},

		// Power
		{name: "PowerActive", code: PowerActive, valid: true},
		{name: "PowerL1", code: PowerL1, valid: true},
		{name: "PowerL2", code: PowerL2, valid: true},
		{name: "PowerL3", code: PowerL3, valid: true},

		// Current
		{name: "CurrentL1", code: CurrentL1, valid: true},
		{name: "CurrentL2", code: CurrentL2, valid: true},
		{name: "CurrentL3", code: CurrentL3, valid: true},

		// Voltage
		{name: "VoltageL1", code: VoltageL1, valid: true},
		{name: "VoltageL2", code: VoltageL2, valid: true},
		{name: "VoltageL3", code: VoltageL3, valid: true},

		// System
		{name: "Frequency", code: Frequency, valid: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.code == "" {
				t.Error("OBIS code is empty")
			}

			// Verify format: should be "C.D.E"
			parts := 0
			for _, c := range tt.code {
				if c == '.' {
					parts++
				}
			}

			if parts != 2 {
				t.Errorf("OBIS code %s should have format C.D.E (got %d dots)", tt.code, parts)
			}
		})
	}
}

// TestDescription verifies that descriptions are provided for known OBIS codes.
func TestDescription(t *testing.T) {
	tests := []struct {
		code     string
		hasDesc  bool
		substring string
	}{
		{code: PowerActive, hasDesc: true, substring: "power"},
		{code: EnergyImport, hasDesc: true, substring: "energy"},
		{code: CurrentL1, hasDesc: true, substring: "current"},
		{code: VoltageL1, hasDesc: true, substring: "voltage"},
		{code: Frequency, hasDesc: true, substring: "frequency"},
		{code: "99.99.99", hasDesc: true, substring: "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			desc := Description(tt.code)
			if tt.hasDesc && desc == "" {
				t.Error("Description should not be empty")
			}
		})
	}
}

// TestOBISCodeUniqueness verifies that all OBIS codes are unique.
func TestOBISCodeUniqueness(t *testing.T) {
	codes := []string{
		EnergyImport, EnergyExport,
		PowerActive, PowerL1, PowerL2, PowerL3,
		CurrentL1, CurrentL2, CurrentL3,
		VoltageL1, VoltageL2, VoltageL3,
		Frequency,
	}

	seen := make(map[string]bool)
	for _, code := range codes {
		if seen[code] {
			t.Errorf("Duplicate OBIS code: %s", code)
		}
		seen[code] = true
	}
}
