// Package emhcasa provides a client for EMH CASA 1.1 Smart Meter Gateways.
package emhcasa

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	smgwreader "github.com/iseeberg79/emh-casa-go"
)

// Client is a CASA 1.1 smart meter gateway client that implements the Gateway interface.
// It handles HTTP digest authentication, custom host headers, and meter data retrieval.
type Client struct {
	httpClient    *http.Client
	hostTransport *hostHeaderTransport
	uri           string
	meterID       string
}

// NewClientDiscover creates a new CASA client with full auto-discovery.
// Discovers the gateway via mDNS and the meter ID from available contracts.
func NewClientDiscover(user, password string) (*Client, error) {
	return NewClient("", user, password, "")
}

// NewClient creates a new CASA client with HTTP digest authentication.
//
// Parameters:
//   - uri: Gateway URI (empty to auto-discover via mDNS)
//   - user: Username for digest authentication
//   - password: Password for digest authentication
//   - meterID: Meter ID (empty to auto-discover from available contracts)
//
// For SSH tunnels, use SetHostHeader("smgw.local") after creating the client.
// Returns an error if credentials are missing or discovery/connection fails.
func NewClient(uri, user, password, meterID string) (*Client, error) {
	// Auto-discover gateway if URI is empty
	if uri == "" {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		discoveredURI, err := DiscoverGatewayURI(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to discover gateway: %w", err)
		}
		uri = discoveredURI
	}

	if user == "" || password == "" {
		return nil, fmt.Errorf("credentials are required")
	}

	uri = defaultScheme(uri, "https")

	// Create HTTP client with custom transport for self-signed certs
	customTransport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
		ForceAttemptHTTP2: false,
	}

	// Create host header transport (can be modified later via SetHostHeader)
	hostTransport := &hostHeaderTransport{
		base: customTransport,
		host: "", // empty = use default from request
	}

	// Add digest authentication
	httpClient := &http.Client{
		Transport: NewDigestTransport(user, password, hostTransport),
	}

	c := &Client{
		httpClient:    httpClient,
		hostTransport: hostTransport,
		uri:           uri,
		meterID:       meterID,
	}

	return c, nil
}

// GetReadings implements the smgwreader.Gateway interface.
// Retrieves current meter readings with metadata from the smgwreader.
// If no meter ID is set, it will be automatically discovered from available contracts.
// Returns an Information struct containing device metadata and readings indexed by OBIS code.
func (c *Client) GetReadings(ctx context.Context) (*smgwreader.Information, error) {
	// Auto-discover meter ID if needed
	if c.meterID == "" {
		if err := c.discoverMeterIDContext(ctx); err != nil {
			return nil, fmt.Errorf("failed to discover meter ID: %w", err)
		}
	}

	// Fetch raw CASA data
	var reading MeterReading
	uri := fmt.Sprintf("%s/json/metering/origin/%s/extended", c.uri, c.meterID)

	if err := c.getJSONContext(ctx, uri, &reading); err != nil {
		return nil, fmt.Errorf("failed to get meter values: %w", err)
	}

	// Convert to standard Information format
	info := &smgwreader.Information{
		Name:           c.meterID,
		Model:          "EMH CASA 1.1",
		Manufacturer:   "EMH",
		FirmwareVersion: "",
		LastUpdate:     time.Now(),
		Readings:       make(map[string]smgwreader.Reading),
	}

	for _, item := range reading.Values {
		obis, err := convertToOBIS(item.LogicalName)
		if err != nil {
			continue
		}

		reading := c.convertReading(item, obis)
		info.Readings[obis] = reading
	}

	if len(info.Readings) == 0 {
		return nil, fmt.Errorf("no valid meter values found")
	}

	return info, nil
}

// DiscoverMeterID implements the smgwreader.Gateway interface.
// Finds and returns the meter ID from the gateway via context.
func (c *Client) DiscoverMeterID(ctx context.Context) (string, error) {
	if err := c.discoverMeterIDContext(ctx); err != nil {
		return "", err
	}
	return c.meterID, nil
}

// discoverMeterIDContext finds the first contract with sensor domains and sets the client's meter ID.
// Returns an error if no contract with sensor domains is found.
func (c *Client) discoverMeterIDContext(ctx context.Context) error {
	var contracts []string
	uri := fmt.Sprintf("%s/json/metering/derived", c.uri)

	if err := c.getJSONContext(ctx, uri, &contracts); err != nil {
		return fmt.Errorf("failed to get contracts: %w", err)
	}

	for _, id := range contracts {
		var contract DerivedContract
		uri := fmt.Sprintf("%s/json/metering/derived/%s", c.uri, id)

		if err := c.getJSONContext(ctx, uri, &contract); err != nil {
			continue
		}

		if len(contract.SensorDomains) > 0 {
			c.meterID = contract.SensorDomains[0]
			return nil
		}
	}

	return fmt.Errorf("no contract with sensor domains found")
}

// MeterID implements the smgwreader.MeterProvider interface.
// Returns the configured meter ID or discovers automatically.
// Returns empty string if discovery fails (for interface compatibility).
// For better error handling, use DiscoverMeterID(ctx) instead.
func (c *Client) MeterID() string {
	if c.meterID == "" {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = c.discoverMeterIDContext(ctx) // silently ignore errors for interface compatibility
	}
	return c.meterID
}

// SetMeterID implements the smgwreader.MeterProvider interface.
// Sets the meter ID for subsequent operations.
func (c *Client) SetMeterID(meterID string) {
	c.meterID = meterID
}

// SetHostHeader implements the smgwreader.HostConfigurer interface.
// Overrides the Host header for all requests.
// Use this for SSH tunnels or proxies when the default doesn't work.
func (c *Client) SetHostHeader(host string) {
	c.hostTransport.host = host
}

// convertReading converts a CASA MeterValue to a standard smgwreader.Reading.
func (c *Client) convertReading(item MeterValue, obis string) smgwreader.Reading {
	raw, _ := strconv.ParseFloat(item.Value, 64)
	val := raw * math.Pow(10, float64(item.Scaler))

	var unit smgwreader.Unit
	switch item.Unit {
	case 27: // W (Watt)
		unit = smgwreader.UnitWatt
	case 30: // Wh (Watthour) → kWh
		val = val / 1000
		unit = smgwreader.UnitWh
	case 33: // A (Ampere)
		unit = smgwreader.UnitAmpere
	case 35: // V (Volt)
		unit = smgwreader.UnitVolt
	case 44: // Hz (Hertz)
		unit = smgwreader.UnitHertz
	}

	return smgwreader.Reading{
		Value:     val,
		Unit:      unit,
		Timestamp: time.Now(),
		OBIS:      obis,
		Quality:   smgwreader.QualityGood,
	}
}

// getJSONContext makes a JSON API call with context support and unmarshals the response.
func (c *Client) getJSONContext(ctx context.Context, uri string, result interface{}) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
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

// getJSON makes a JSON API call and unmarshals the response (for backward compatibility).
// Deprecated: Use getJSONContext instead.
func (c *Client) getJSON(uri string, result interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return c.getJSONContext(ctx, uri, result)
}

// GetMeterValues fetches and parses current meter readings from the smgwreader.
// Deprecated: Use GetReadings instead. This method is provided for backward compatibility.
// If no meter ID is set, it will be automatically discovered from available contracts.
// Returns a map of OBIS codes to float64 values.
func (c *Client) GetMeterValues() (map[string]float64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	info, err := c.GetReadings(ctx)
	if err != nil {
		return nil, err
	}

	// Convert Information back to map[string]float64 for backward compatibility
	values := make(map[string]float64)
	for obis, reading := range info.Readings {
		values[obis] = reading.Value
	}

	return values, nil
}

// defaultScheme adds a default scheme if missing
func defaultScheme(uri, scheme string) string {
	if !strings.HasPrefix(uri, "http://") && !strings.HasPrefix(uri, "https://") {
		return scheme + "://" + uri
	}
	return uri
}
