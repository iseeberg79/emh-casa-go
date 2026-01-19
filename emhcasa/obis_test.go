package emhcasa

import (
	"testing"
)

// TestConvertToOBIS tests CASA logical name to OBIS conversion
func TestConvertToOBIS(t *testing.T) {
	tests := []struct {
		name        string
		logicalName string
		want        string
		wantErr     bool
	}{
		{
			name:        "current power OBIS 16.7.0",
			logicalName: "0100100700FF",
			want:        "16.7.0",
			wantErr:     false,
		},
		{
			name:        "total energy OBIS 1.8.0",
			logicalName: "0100010800FF",
			want:        "1.8.0",
			wantErr:     false,
		},
		{
			name:        "phase 1 current OBIS 31.7.0",
			logicalName: "01001F0700FF",
			want:        "31.7.0",
			wantErr:     false,
		},
		{
			name:        "phase 1 voltage OBIS 32.7.0",
			logicalName: "0100200700FF",
			want:        "32.7.0",
			wantErr:     false,
		},
		{
			name:        "grid export OBIS 2.8.0",
			logicalName: "0100020800FF",
			want:        "2.8.0",
			wantErr:     false,
		},
		{
			name:        "invalid hex length",
			logicalName: "010010",
			want:        "",
			wantErr:     true,
		},
		{
			name:        "invalid hex characters",
			logicalName: "0100ZZZZ00FF",
			want:        "",
			wantErr:     true,
		},
		{
			name:        "with decimal suffix (should ignore)",
			logicalName: "0100100700FF.1",
			want:        "16.7.0",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := convertToOBIS(tt.logicalName)
			if (err != nil) != tt.wantErr {
				t.Errorf("convertToOBIS() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("convertToOBIS() = %v, want %v", got, tt.want)
			}
		})
	}
}
