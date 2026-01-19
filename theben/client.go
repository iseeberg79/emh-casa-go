// Package theben provides a client for Theben Conexa Smart Meter Gateways.
package theben

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	smgwreader "github.com/iseeberg79/emh-casa-go"
	"github.com/jpfielding/go-http-digest/pkg/digest"
)

// Client is a Theben Conexa smart meter gateway client that implements the Gateway interface.
type Client struct {
	httpClient *http.Client
	baseURL    string
	username   string
	password   string
	meterID    string // Usage point ID
}

// jsonRequest represents a JSON-RPC style request to Theben gateway
type jsonRequest struct {
	Method string `json:"method"`
}

// userInfoResponse represents the response from user-info method
type userInfoResponse struct {
	UserInfo struct {
		UsagePoints []usagePoint `json:"usage-points"`
	} `json:"user-info"`
}

// usagePoint represents a meter connection point
type usagePoint struct {
	ID        string `json:"id"`
	TafState  string `json:"taf-state"`
	TafNumber string `json:"taf-number"`
}

// readingsResponse represents the response from readings method
type readingsResponse struct {
	Readings struct {
		Channels []channel `json:"channels"`
	} `json:"readings"`
}

// channel represents a meter channel with readings
type channel struct {
	Readings []reading `json:"readings"`
}

// reading represents a single meter reading
type reading struct {
	OBIS        string `json:"obis"`
	Value       string `json:"value"`
	CaptureTime string `json:"capture-time"`
}

// smgwInfoResponse represents the response from smgw-info method
type smgwInfoResponse struct {
	SMGWInfo struct {
		FirmwareVersion string `json:"firmware-version"`
		Manufacturer    string `json:"manufacturer"`
		Model           string `json:"model"`
	} `json:"smgw-info"`
}

// NewClient creates a new Theben Conexa gateway client with HTTP digest authentication.
//
// Parameters:
//   - baseURL: Gateway base URL (e.g., "https://192.168.1.100")
//   - username: Username for digest authentication
//   - password: Password for digest authentication
//
// Theben Conexa gateways use self-signed certificates and JSON API.
func NewClient(baseURL, username, password string) (*Client, error) {
	if username == "" || password == "" {
		return nil, fmt.Errorf("credentials are required")
	}

	if baseURL == "" {
		return nil, fmt.Errorf("base URL is required")
	}

	// Ensure baseURL has scheme
	if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
		baseURL = "https://" + baseURL
	}

	// Create HTTP client with self-signed cert support
	customTransport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	httpClient := &http.Client{
		Transport: digest.NewTransport(username, password, customTransport),
		Timeout:   10 * time.Second,
	}

	return &Client{
		httpClient: httpClient,
		baseURL:    baseURL,
		username:   username,
		password:   password,
	}, nil
}

// GetReadings implements the smgwreader.Gateway interface.
func (c *Client) GetReadings(ctx context.Context) (*smgwreader.Information, error) {
	// Step 1: Get SMGW info (firmware, manufacturer, model)
	smgwInfo, err := c.getSMGWInfo(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get SMGW info: %w", err)
	}

	// Step 2: Get usage point ID (auto-discover if not set)
	if c.meterID == "" {
		usagePointID, err := c.getUsagePointID(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get usage point: %w", err)
		}
		c.meterID = usagePointID
	}

	// Step 3: Get readings
	readings, err := c.getReadings(ctx, c.meterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get readings: %w", err)
	}

	// Build Information object
	info := &smgwreader.Information{
		Name:            c.meterID,
		Model:           smgwInfo.SMGWInfo.Model,
		Manufacturer:    smgwInfo.SMGWInfo.Manufacturer,
		FirmwareVersion: smgwInfo.SMGWInfo.FirmwareVersion,
		LastUpdate:      time.Now(),
		Readings:        readings,
	}

	return info, nil
}

// DiscoverMeterID implements the smgwreader.Gateway interface.
func (c *Client) DiscoverMeterID(ctx context.Context) (string, error) {
	meterID, err := c.getUsagePointID(ctx)
	if err != nil {
		return "", err
	}
	c.meterID = meterID
	return c.meterID, nil
}

// MeterID implements the smgwreader.MeterProvider interface.
// Returns the configured meter ID (usage point ID) or discovers automatically.
func (c *Client) MeterID() string {
	if c.meterID == "" {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_, _ = c.DiscoverMeterID(ctx) // silently ignore errors for interface compatibility
	}
	return c.meterID
}

// SetMeterID implements the smgwreader.MeterProvider interface.
// Sets the meter ID (usage point ID) for subsequent operations.
func (c *Client) SetMeterID(meterID string) {
	c.meterID = meterID
}

// getSMGWInfo retrieves gateway information
func (c *Client) getSMGWInfo(ctx context.Context) (*smgwInfoResponse, error) {
	req := jsonRequest{Method: "smgw-info"}
	var resp smgwInfoResponse

	if err := c.doJSONRequest(ctx, req, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// getUsagePointID retrieves the usage point ID (meter ID)
func (c *Client) getUsagePointID(ctx context.Context) (string, error) {
	req := jsonRequest{Method: "user-info"}
	var resp userInfoResponse

	if err := c.doJSONRequest(ctx, req, &resp); err != nil {
		return "", err
	}

	// Prefer usage points with taf-state="running" and taf-number="7"
	for _, up := range resp.UserInfo.UsagePoints {
		if up.TafState == "running" && up.TafNumber == "7" {
			return up.ID, nil
		}
	}

	// Fallback to first available usage point
	if len(resp.UserInfo.UsagePoints) > 0 {
		return resp.UserInfo.UsagePoints[0].ID, nil
	}

	return "", fmt.Errorf("no usage points found")
}

// getReadings retrieves meter readings for a usage point
func (c *Client) getReadings(ctx context.Context, usagePointID string) (map[string]smgwreader.Reading, error) {
	req := jsonRequest{Method: "readings"}
	var resp readingsResponse

	if err := c.doJSONRequest(ctx, req, &resp); err != nil {
		return nil, err
	}

	readings := make(map[string]smgwreader.Reading)

	// Extract readings from first channel (limitation: only one reading per channel currently)
	for _, ch := range resp.Readings.Channels {
		for _, r := range ch.Readings {
			// Convert hex OBIS to standard format
			obis := c.convertOBIS(r.OBIS)
			if obis == "" {
				continue
			}

			// Parse value (stored in deciWatts, convert to Watts or kWh)
			value, err := c.parseValue(r.Value, obis)
			if err != nil {
				continue
			}

			// Parse timestamp
			ts := time.Now()
			if r.CaptureTime != "" {
				if parsed, err := time.Parse(time.RFC3339, r.CaptureTime); err == nil {
					ts = parsed
				}
			}

			// Determine unit
			unit := c.determineUnit(obis)

			readings[obis] = smgwreader.Reading{
				Value:     value,
				Unit:      unit,
				Timestamp: ts,
				OBIS:      obis,
				Quality:   smgwreader.QualityGood,
			}
		}
	}

	return readings, nil
}

// doJSONRequest performs a JSON request to the gateway
func (c *Client) doJSONRequest(ctx context.Context, request interface{}, response interface{}) error {
	body, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL, strings.NewReader(string(body)))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if err := json.Unmarshal(respBody, response); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return nil
}

// convertOBIS converts Theben hex OBIS code to standard format
// Theben uses hex format: 0100010800ff -> 1-0:1.8.0
func (c *Client) convertOBIS(hexOBIS string) string {
	// Remove any non-hex characters
	hexOBIS = strings.ToLower(strings.TrimSpace(hexOBIS))

	// OBIS code mappings for Theben
	mappings := map[string]string{
		"0100010800ff": "1-0:1.8.0", // Active energy import
		"0100020800ff": "1-0:2.8.0", // Active energy export
		"0100100700ff": "1-0:16.7.0", // Current active power (NEW - added for 16.7.0 support)
		"01001f0700ff": "1-0:31.7.0", // Phase 1 current
		"0100200700ff": "1-0:32.7.0", // Phase 1 voltage
		"0100240700ff": "1-0:36.7.0", // Phase 1 power
		"0100330700ff": "1-0:51.7.0", // Phase 2 current
		"0100340700ff": "1-0:52.7.0", // Phase 2 voltage
		"0100380700ff": "1-0:56.7.0", // Phase 2 power
		"0100470700ff": "1-0:71.7.0", // Phase 3 current
		"0100480700ff": "1-0:72.7.0", // Phase 3 voltage
		"01004c0700ff": "1-0:76.7.0", // Phase 3 power
		"01000e0700ff": "1-0:14.7.0", // Frequency
	}

	if standard, ok := mappings[hexOBIS]; ok {
		return standard
	}

	return "" // Unknown OBIS code
}

// parseValue parses the value string and converts units
func (c *Client) parseValue(valueStr, obis string) (float64, error) {
	// Parse the numeric value
	var value float64
	if _, err := fmt.Sscanf(valueStr, "%f", &value); err != nil {
		return 0, err
	}

	// Theben stores values in different scales depending on the OBIS code
	// Energy values (x.8.0) are in Wh and need to be converted to kWh
	// Power values (x.7.0) are in deciWatts and need to be divided by 10000
	if strings.Contains(obis, ".8.") {
		// Energy: convert Wh to kWh
		return value / 1000, nil
	}

	// Power and other instantaneous values: convert from deciWatts
	return value / 10000, nil
}

// determineUnit determines the unit based on OBIS code
func (c *Client) determineUnit(obis string) smgwreader.Unit {
	// OBIS format: 1-0:C.D.E
	parts := strings.Split(obis, ":")
	if len(parts) < 2 {
		return smgwreader.UnitWatt
	}

	obisCode := parts[1]
	codeParts := strings.Split(obisCode, ".")
	if len(codeParts) < 2 {
		return smgwreader.UnitWatt
	}

	switch codeParts[1] {
	case "8": // Energy
		return smgwreader.UnitWh
	case "7": // Instantaneous values
		c := codeParts[0]
		switch {
		case c == "16" || c == "36" || c == "56" || c == "76": // Power
			return smgwreader.UnitWatt
		case c == "31" || c == "51" || c == "71": // Current
			return smgwreader.UnitAmpere
		case c == "32" || c == "52" || c == "72": // Voltage
			return smgwreader.UnitVolt
		case c == "14": // Frequency
			return smgwreader.UnitHertz
		}
	}

	return smgwreader.UnitWatt // Default
}
