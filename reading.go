package smgwreader

import "time"

// Information represents gateway/meter information and readings.
// It contains device metadata and all current meter readings indexed by OBIS code.
type Information struct {
	// Name is the gateway or meter name/identifier
	Name string

	// Model is the device model identifier (e.g., "EMH CASA 1.1")
	Model string

	// Manufacturer is the device manufacturer name
	Manufacturer string

	// FirmwareVersion is the device firmware version
	FirmwareVersion string

	// LastUpdate is the timestamp when this information was retrieved from the gateway
	LastUpdate time.Time

	// Readings is a map of OBIS codes to Reading objects
	// OBIS codes (e.g., "16.7.0" for current power) are used as keys for vendor-agnostic access
	Readings map[string]Reading
}

// Reading represents a single meter reading with metadata.
// It includes the measured value, unit, timestamp, quality indicator, and OBIS code.
type Reading struct {
	// Value is the measured value, already scaled and unit-converted
	// For example: Power in Watts, Energy in kWh, Current in Amperes
	Value float64

	// Unit is the unit type of the reading
	Unit Unit

	// Timestamp is when the reading was captured by the gateway
	Timestamp time.Time

	// OBIS is the OBIS code for this reading (e.g., "16.7.0")
	OBIS string

	// Quality indicates the data quality of this reading
	Quality Quality
}

// Unit represents the unit of measurement for a reading.
// The numeric values correspond to DLMS/COSEM unit codes.
type Unit int

// Standard DLMS/COSEM unit codes
const (
	// UnitWatt - Active Power (W)
	UnitWatt Unit = 27

	// UnitWh - Energy (Wh, typically converted to kWh)
	UnitWh Unit = 30

	// UnitAmpere - Electric Current (A)
	UnitAmpere Unit = 33

	// UnitVolt - Voltage (V)
	UnitVolt Unit = 35

	// UnitHertz - Frequency (Hz)
	UnitHertz Unit = 44
)

// String returns the unit abbreviation.
func (u Unit) String() string {
	switch u {
	case UnitWatt:
		return "W"
	case UnitWh:
		return "kWh"
	case UnitAmpere:
		return "A"
	case UnitVolt:
		return "V"
	case UnitHertz:
		return "Hz"
	default:
		return "?"
	}
}

// Quality indicates the reliability of a reading.
type Quality int

// Quality levels
const (
	// QualityGood - Reading is current and valid
	QualityGood Quality = 0

	// QualityStale - Reading is valid but may be outdated (older than expected)
	QualityStale Quality = 1

	// QualityInvalid - Reading is invalid or unavailable
	QualityInvalid Quality = 2
)

// String returns the quality description.
func (q Quality) String() string {
	switch q {
	case QualityGood:
		return "good"
	case QualityStale:
		return "stale"
	case QualityInvalid:
		return "invalid"
	default:
		return "unknown"
	}
}
