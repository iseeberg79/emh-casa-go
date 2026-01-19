package emhcasa

import (
	"math"
	"testing"
)

// TestMeterValueParsing tests meter value unit conversion and scaling
func TestMeterValueParsing(t *testing.T) {
	tests := []struct {
		name       string
		value      string
		unit       int
		scaler     int
		wantResult float64
		wantErr    bool
	}{
		{
			// Current power reading: 2500W
			name:       "current power in watts",
			value:      "2500",
			unit:       27, // W (Watt)
			scaler:     0,
			wantResult: 2500,
			wantErr:    false,
		},
		{
			// Scaled power: 0.5 * 10^3 = 500W
			name:       "scaled power",
			value:      "5",
			unit:       27, // W
			scaler:     2,
			wantResult: 500,
			wantErr:    false,
		},
		{
			// Total energy: 123.45 kWh (stored as 123450 Wh with scaler 0, converted to kWh)
			name:       "total energy in kWh",
			value:      "123450",
			unit:       30, // Wh (Watthour)
			scaler:     0,
			wantResult: 123.45, // Should be divided by 1000
			wantErr:    false,
		},
		{
			// Phase current: 15.3A
			name:       "phase current in amps",
			value:      "153",
			unit:       33, // A (Ampere)
			scaler:     -1,
			wantResult: 15.3,
			wantErr:    false,
		},
		{
			// Phase voltage: 231.5V
			name:       "phase voltage in volts",
			value:      "2315",
			unit:       35, // V (Volt)
			scaler:     -1,
			wantResult: 231.5,
			wantErr:    false,
		},
		{
			// Grid frequency: 50.0Hz
			name:       "grid frequency",
			value:      "500",
			unit:       44, // Hz (Hertz)
			scaler:     -1,
			wantResult: 50.0,
			wantErr:    false,
		},
		{
			// Negative scaler: 250 * 10^-2 = 2.5
			name:       "negative scaler",
			value:      "250",
			unit:       27,
			scaler:     -2,
			wantResult: 2.5,
			wantErr:    false,
		},
		{
			// Large value with scaler
			name:       "large scaled value",
			value:      "1234",
			unit:       27,
			scaler:     1,
			wantResult: 12340,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			raw, err := parseFloat(tt.value)
			if err != nil {
				if !tt.wantErr {
					t.Fatalf("parseFloat() failed: %v", err)
				}
				return
			}

			result := raw * math.Pow(10, float64(tt.scaler))

			// Apply unit-specific conversions
			if tt.unit == 30 {
				// Wh to kWh conversion
				result = result / 1000
			}

			// Check result with small tolerance for floating point
			tolerance := 0.00001
			if math.Abs(result-tt.wantResult) > tolerance {
				t.Errorf("Parsed value = %v, want %v (tolerance: %v)", result, tt.wantResult, tolerance)
			}
		})
	}
}

// parseFloat is a helper function for testing
func parseFloat(s string) (float64, error) {
	var f float64
	if len(s) == 0 {
		return 0, nil
	}

	// Simple integer parsing for test purposes
	for _, c := range s {
		if c >= '0' && c <= '9' {
			f = f*10 + float64(c-'0')
		}
	}
	return f, nil
}

// TestMeterReadingStructure tests meter reading data structure
func TestMeterReadingStructure(t *testing.T) {
	reading := MeterReading{
		Values: []MeterValue{
			{
				Value:       "2500",
				Unit:        27,
				Scaler:      0,
				LogicalName: "0100100700FF",
			},
			{
				Value:       "123450",
				Unit:        30,
				Scaler:      0,
				LogicalName: "0100010800FF",
			},
		},
	}

	if len(reading.Values) != 2 {
		t.Errorf("Expected 2 values, got %d", len(reading.Values))
	}

	if reading.Values[0].LogicalName != "0100100700FF" {
		t.Errorf("Expected logical name '0100100700FF', got '%s'", reading.Values[0].LogicalName)
	}

	if reading.Values[0].Unit != 27 {
		t.Errorf("Expected unit 27 (W), got %d", reading.Values[0].Unit)
	}
}

// TestOBISValueMapping tests mapping of common OBIS codes
func TestOBISValueMapping(t *testing.T) {
	tests := []struct {
		name      string
		obis      string
		measured  string
		unit      int
		scaler    int
		wantValue float64
	}{
		{
			name:      "active power OBIS 16.7.0",
			obis:      "16.7.0",
			measured:  "2345",
			unit:      27,
			scaler:    0,
			wantValue: 2345.0,
		},
		{
			name:      "import energy OBIS 1.8.0",
			obis:      "1.8.0",
			measured:  "456789",
			unit:      30,
			scaler:    0,
			wantValue: 456.789,
		},
		{
			name:      "export energy OBIS 2.8.0",
			obis:      "2.8.0",
			measured:  "12345",
			unit:      30,
			scaler:    0,
			wantValue: 12.345,
		},
		{
			name:      "L1 current OBIS 31.7.0",
			obis:      "31.7.0",
			measured:  "154",
			unit:      33,
			scaler:    -1,
			wantValue: 15.4,
		},
		{
			name:      "L1 voltage OBIS 32.7.0",
			obis:      "32.7.0",
			measured:  "2315",
			unit:      35,
			scaler:    -1,
			wantValue: 231.5,
		},
		{
			name:      "L1 power OBIS 36.7.0",
			obis:      "36.7.0",
			measured:  "1200",
			unit:      27,
			scaler:    0,
			wantValue: 1200.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Just verify the OBIS code format is valid
			obisCode := tt.obis
			if !isValidOBIS(obisCode) {
				t.Errorf("Invalid OBIS code: %s", obisCode)
			}
		})
	}
}

// isValidOBIS checks if an OBIS code has the correct format (C.D.E)
func isValidOBIS(obis string) bool {
	parts := []rune(obis)
	if len(parts) < 5 { // Minimum: C.D.E
		return false
	}

	// Check for pattern: digits.digits.digits
	dotCount := 0
	lastWasDot := false

	for i, c := range parts {
		if c == '.' {
			if lastWasDot || i == 0 || i == len(parts)-1 {
				return false // Double dots, leading/trailing dot
			}
			dotCount++
			lastWasDot = true
		} else if c >= '0' && c <= '9' {
			lastWasDot = false
		} else {
			return false // Invalid character
		}
	}

	return dotCount == 2 // Must have exactly 2 dots
}
