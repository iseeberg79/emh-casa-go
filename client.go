// Package emhcasa provides a client for EMH CASA 1.1 Smart Meter Gateways
package emhcasa

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"
)

// Client is a CASA 1.1 smart meter gateway client.
// It handles HTTP digest authentication, custom host headers, and meter data retrieval.
type Client struct {
	httpClient *http.Client
	uri        string
	meterID    string
}

// NewClient creates a new CASA client with HTTP digest authentication.
//
// Parameters:
//   - uri: Gateway URI (http or https, e.g., "https://192.168.33.2")
//   - user: Username for digest authentication
//   - password: Password for digest authentication
//   - meterID: Meter ID to use (empty string to auto-discover from available contracts)
//   - hostHeader: Custom Host header for routing (typically the gateway IP)
//
// The client automatically discovers the meter ID if not provided.
// Returns an error if credentials are missing or meter ID discovery fails.
func NewClient(uri, user, password, meterID, hostHeader string) (*Client, error) {
	if uri == "" {
		return nil, fmt.Errorf("uri is required")
	}

	if user == "" || password == "" {
		return nil, fmt.Errorf("credentials are required")
	}

	uri = defaultScheme(uri, "https")
	host := hostHeader

	// If no host provided, try to derive from URI
	if host == "" {
		derived, err := parseURIHost(uri)
		if err != nil {
			return nil, fmt.Errorf("host required and could not be derived: %w", err)
		}
		host = derived
	}

	// Create HTTP client with custom transport for self-signed certs
	customTransport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
		ForceAttemptHTTP2: false,
	}

	hostTransport := &hostHeaderTransport{
		base: customTransport,
		host: host,
	}

	// Add digest authentication
	httpClient := &http.Client{
		Transport: NewDigestTransport(user, password, hostTransport),
	}

	c := &Client{
		httpClient: httpClient,
		uri:        uri,
		meterID:    meterID,
	}

	// Discover meter ID if not provided
	if c.meterID == "" {
		if err := c.DiscoverMeterID(); err != nil {
			return nil, fmt.Errorf("failed to discover meter ID: %w", err)
		}
	}

	return c, nil
}

// DiscoverMeterID finds the first contract with sensor domains and sets the client's meter ID.
// This is automatically called by NewClient if no meter ID is provided.
// Returns an error if no contract with sensor domains is found.
func (c *Client) DiscoverMeterID() error {
	var contracts []string
	uri := fmt.Sprintf("%s/json/metering/derived", c.uri)

	if err := c.getJSON(uri, &contracts); err != nil {
		return fmt.Errorf("failed to get contracts: %w", err)
	}

	for _, id := range contracts {
		var contract DerivedContract
		uri := fmt.Sprintf("%s/json/metering/derived/%s", c.uri, id)

		if err := c.getJSON(uri, &contract); err != nil {
			continue
		}

		if len(contract.SensorDomains) > 0 {
			c.meterID = contract.SensorDomains[0]
			return nil
		}
	}

	return fmt.Errorf("no contract with sensor domains found")
}

// GetMeterValues fetches and parses current meter readings from the gateway.
//
// Returns a map of OBIS codes to float64 values. OBIS codes use the format C.D.E
// where common values include:
//   - 16.7.0: Current power (W)
//   - 1.8.0: Total imported energy (kWh)
//   - 2.8.0: Total exported energy (kWh)
//   - 31.7.0, 51.7.0, 71.7.0: Phase currents (A)
//   - 32.7.0, 52.7.0, 72.7.0: Phase voltages (V)
//
// Returns an error if the gateway request fails or no valid values are found.
func (c *Client) GetMeterValues() (map[string]float64, error) {
	if c.meterID == "" {
		return nil, fmt.Errorf("meter ID not set")
	}

	var reading MeterReading
	uri := fmt.Sprintf("%s/json/metering/origin/%s/extended", c.uri, c.meterID)

	if err := c.getJSON(uri, &reading); err != nil {
		return nil, fmt.Errorf("failed to get meter values: %w", err)
	}

	values := make(map[string]float64)

	for _, item := range reading.Values {
		obis, err := convertToOBIS(item.LogicalName)
		if err != nil {
			continue
		}

		raw, err := strconv.ParseFloat(item.Value, 64)
		if err != nil {
			continue
		}

		val := raw * math.Pow(10, float64(item.Scaler))

		switch item.Unit {
		case 27: // W (Watt)
			values[obis] = val
		case 30: // Wh (Watthour) → kWh
			values[obis] = val / 1000
		case 33: // A (Ampere)
			values[obis] = val
		case 35: // V (Volt)
			values[obis] = val
		case 44: // Hz (Hertz)
			values[obis] = val
		}
	}

	if len(values) == 0 {
		return nil, fmt.Errorf("no valid meter values found")
	}

	return values, nil
}

// MeterID returns the currently configured meter ID.
// This is set either explicitly during NewClient or discovered automatically.
func (c *Client) MeterID() string {
	return c.meterID
}

// getJSON makes a JSON API call and unmarshals the response
func (c *Client) getJSON(uri string, result interface{}) error {
	resp, err := c.httpClient.Get(uri)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if err := json.Unmarshal(body, result); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return nil
}

// convertToOBIS converts CASA logical name to OBIS C.D.E format
func convertToOBIS(logicalName string) (string, error) {
	hex := strings.SplitN(logicalName, ".", 2)[0]

	if len(hex) != 12 {
		return "", fmt.Errorf("unexpected logical name: %s", logicalName)
	}

	c, err := strconv.ParseInt(hex[4:6], 16, 64)
	if err != nil {
		return "", err
	}
	d, err := strconv.ParseInt(hex[6:8], 16, 64)
	if err != nil {
		return "", err
	}
	e, err := strconv.ParseInt(hex[8:10], 16, 64)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%d.%d.%d", c, d, e), nil
}

// parseURIHost extracts the host from a URI
func parseURIHost(uri string) (string, error) {
	uri = strings.TrimPrefix(uri, "https://")
	uri = strings.TrimPrefix(uri, "http://")

	if idx := strings.Index(uri, "/"); idx != -1 {
		uri = uri[:idx]
	}
	if idx := strings.Index(uri, ":"); idx != -1 {
		uri = uri[:idx]
	}

	if uri == "" {
		return "", fmt.Errorf("invalid uri")
	}

	return uri, nil
}

// defaultScheme adds a default scheme if missing
func defaultScheme(uri, scheme string) string {
	if !strings.HasPrefix(uri, "http://") && !strings.HasPrefix(uri, "https://") {
		return scheme + "://" + uri
	}
	return uri
}
