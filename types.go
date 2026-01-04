// Package emhcasa provides a client library for EMH CASA 1.1 Smart Meter Gateways.
//
// This library handles HTTP digest authentication, OBIS value conversion, and
// automatic meter ID discovery for CASA 1.1 smart meter gateways.
//
// Basic usage:
//
//	client, err := emhcasa.NewClient(
//		"https://192.168.33.2",
//		"admin",
//		"password",
//		"",                  // auto-discover meter ID
//		"192.168.33.2",      // host header
//	)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	values, err := client.GetMeterValues()
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	power := values["16.7.0"] // OBIS 16.7.0 = current power in W
package emhcasa

// DerivedContract represents a metering contract from the CASA gateway.
type DerivedContract struct {
	TafType       string   `json:"taf_type"`
	SensorDomains []string `json:"sensor_domains"`
}

// MeterValue represents a single meter reading value from the gateway.
type MeterValue struct {
	Value       string `json:"value"`
	Unit        int    `json:"unit"`         // 27 = W, 30 = Wh, 33 = A, 35 = V, 44 = Hz
	Scaler      int    `json:"scaler"`       // power-of-10 multiplier
	LogicalName string `json:"logical_name"` // CASA logical name in hex format
}

// MeterReading represents the complete meter reading response from the gateway.
type MeterReading struct {
	Values []MeterValue `json:"values"`
}
