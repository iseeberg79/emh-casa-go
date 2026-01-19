package emhcasa

import (
	"fmt"
	"strconv"
	"strings"
)

// convertToOBIS converts CASA logical name to OBIS C.D.E format.
// CASA logical names are in a specific hex format that needs to be parsed to extract
// the OBIS code bytes C, D, and E.
// Example: "0600011B0600" → "1.27.0"
func convertToOBIS(logicalName string) (string, error) {
	// Extract the hex part (before the dot if present)
	hex := strings.SplitN(logicalName, ".", 2)[0]

	if len(hex) != 12 {
		return "", fmt.Errorf("unexpected logical name: %s", logicalName)
	}

	// Parse OBIS bytes from hex: bytes 2-3=C, 3-4=D, 4-5=E
	// Hex format: [2 bytes A][2 bytes B][2 bytes C][2 bytes D][2 bytes E][2 bytes F]
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
