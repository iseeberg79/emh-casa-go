// Package ppc provides a client for PPC Smart Meter Gateways.
package ppc

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	smgwreader "github.com/iseeberg79/emh-casa-go"
	"github.com/jpfielding/go-http-digest/pkg/digest"
	"golang.org/x/net/html"
)

// Client is a PPC smart meter gateway client that implements the Gateway interface.
type Client struct {
	httpClient *http.Client
	baseURL    string
	username   string
	password   string
	meterID    string
}

// NewClient creates a new PPC gateway client with HTTP digest authentication.
//
// Parameters:
//   - baseURL: Gateway base URL (e.g., "https://192.168.1.100")
//   - username: Username for digest authentication
//   - password: Password for digest authentication
//
// PPC gateways use self-signed certificates and HTML-based responses.
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
		Timeout:   30 * time.Second,
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
	// Step 1: Get meter ID (auto-discover if not set)
	if c.meterID == "" {
		meterID, err := c.getMeterID(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get meter ID: %w", err)
		}
		c.meterID = meterID
	}

	// Step 2: Get meter profile data
	readings, err := c.getMeterProfile(ctx, c.meterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get meter profile: %w", err)
	}

	// Build Information object
	info := &smgwreader.Information{
		Name:         c.meterID,
		Model:        "PPC SMGW",
		Manufacturer: "PPC",
		LastUpdate:   time.Now(),
		Readings:     readings,
	}

	return info, nil
}

// DiscoverMeterID implements the smgwreader.Gateway interface.
func (c *Client) DiscoverMeterID(ctx context.Context) (string, error) {
	meterID, err := c.getMeterID(ctx)
	if err != nil {
		return "", err
	}
	c.meterID = meterID
	return c.meterID, nil
}

// MeterID implements the smgwreader.MeterProvider interface.
// Returns the configured meter ID or discovers automatically.
func (c *Client) MeterID() string {
	if c.meterID == "" {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		_, _ = c.DiscoverMeterID(ctx) // silently ignore errors for interface compatibility
	}
	return c.meterID
}

// SetMeterID implements the smgwreader.MeterProvider interface.
// Sets the meter ID for subsequent operations.
func (c *Client) SetMeterID(meterID string) {
	c.meterID = meterID
}

// getMeterID retrieves the meter ID from the meterform endpoint
func (c *Client) getMeterID(ctx context.Context) (string, error) {
	formData := url.Values{}
	formData.Set("action", "meterform")

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Parse HTML to extract meter ID
	doc, err := html.Parse(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to parse HTML: %w", err)
	}

	meterID := c.extractMeterID(doc)
	if meterID == "" {
		return "", fmt.Errorf("meter ID not found in response")
	}

	return meterID, nil
}

// getMeterProfile retrieves meter readings from the showMeterProfile endpoint
func (c *Client) getMeterProfile(ctx context.Context, meterID string) (map[string]smgwreader.Reading, error) {
	formData := url.Values{}
	formData.Set("action", "showMeterProfile")
	formData.Set("meter_id", meterID)

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Parse HTML to extract readings
	doc, err := html.Parse(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	readings := c.extractReadings(doc)
	if len(readings) == 0 {
		return nil, fmt.Errorf("no readings found")
	}

	return readings, nil
}

// extractMeterID extracts the meter ID from HTML document
func (c *Client) extractMeterID(n *html.Node) string {
	// Look for meter ID in form or hidden input
	// This is a simplified implementation - adjust based on actual HTML structure
	if n.Type == html.ElementNode && n.Data == "input" {
		var name, value string
		for _, attr := range n.Attr {
			if attr.Key == "name" && strings.Contains(attr.Val, "meter") {
				name = attr.Val
			}
			if attr.Key == "value" {
				value = attr.Val
			}
		}
		if name != "" && value != "" {
			return value
		}
	}

	for child := n.FirstChild; child != nil; child = child.NextSibling {
		if result := c.extractMeterID(child); result != "" {
			return result
		}
	}

	return ""
}

// extractReadings extracts OBIS readings from HTML table
func (c *Client) extractReadings(n *html.Node) map[string]smgwreader.Reading {
	readings := make(map[string]smgwreader.Reading)

	// Find table with id="metervalue"
	table := c.findElementByID(n, "metervalue")
	if table == nil {
		return readings
	}

	// Extract rows from table
	c.extractTableRows(table, readings)

	return readings
}

// findElementByID finds an HTML element by ID
func (c *Client) findElementByID(n *html.Node, id string) *html.Node {
	if n.Type == html.ElementNode {
		for _, attr := range n.Attr {
			if attr.Key == "id" && attr.Val == id {
				return n
			}
		}
	}

	for child := n.FirstChild; child != nil; child = child.NextSibling {
		if result := c.findElementByID(child, id); result != nil {
			return result
		}
	}

	return nil
}

// extractTableRows extracts readings from table rows
func (c *Client) extractTableRows(table *html.Node, readings map[string]smgwreader.Reading) {
	var extractRow func(*html.Node)
	extractRow = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "tr" {
			// Extract OBIS code and value from row
			obis := c.extractCellByID(n, "table_metervalues_col_obis")
			value := c.extractCellByID(n, "table_metervalues_col_wert")
			timestamp := c.extractCellByID(n, "table_metervalues_col_timestamp")

			if obis != "" && value != "" {
				// Parse value
				val, err := strconv.ParseFloat(value, 64)
				if err != nil {
					return
				}

				// Determine unit based on OBIS code
				unit := c.determineUnit(obis)

				// Parse timestamp
				ts := time.Now()
				if timestamp != "" {
					if parsed, err := time.Parse("2006-01-02 15:04:05", timestamp); err == nil {
						ts = parsed
					}
				}

				readings[obis] = smgwreader.Reading{
					Value:     val,
					Unit:      unit,
					Timestamp: ts,
					OBIS:      obis,
					Quality:   smgwreader.QualityGood,
				}
			}
		}

		for child := n.FirstChild; child != nil; child = child.NextSibling {
			extractRow(child)
		}
	}

	extractRow(table)
}

// extractCellByID extracts text content from a cell with specific ID
func (c *Client) extractCellByID(row *html.Node, id string) string {
	cell := c.findElementByID(row, id)
	if cell == nil {
		return ""
	}

	return c.extractText(cell)
}

// extractText extracts all text content from a node
func (c *Client) extractText(n *html.Node) string {
	if n.Type == html.TextNode {
		return strings.TrimSpace(n.Data)
	}

	var text string
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		text += c.extractText(child)
	}

	return strings.TrimSpace(text)
}

// determineUnit determines the unit based on OBIS code
func (c *Client) determineUnit(obis string) smgwreader.Unit {
	// Common OBIS code patterns:
	// x.8.0 = Energy (kWh)
	// x.7.0 = Power (W) or Current/Voltage
	parts := strings.Split(obis, ".")
	if len(parts) >= 2 {
		switch parts[1] {
		case "8": // Energy
			return smgwreader.UnitWh
		case "7": // Instantaneous values
			if parts[0] == "16" || parts[0] == "36" || parts[0] == "56" || parts[0] == "76" {
				return smgwreader.UnitWatt // Power
			}
			if parts[0] == "31" || parts[0] == "51" || parts[0] == "71" {
				return smgwreader.UnitAmpere // Current
			}
			if parts[0] == "32" || parts[0] == "52" || parts[0] == "72" {
				return smgwreader.UnitVolt // Voltage
			}
			if parts[0] == "14" {
				return smgwreader.UnitHertz // Frequency
			}
		}
	}

	return smgwreader.UnitWatt // Default
}
