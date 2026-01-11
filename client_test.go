package emhcasa

import (
	"strings"
	"testing"
)

// TestParseURIHost tests URI host extraction
func TestParseURIHost(t *testing.T) {
	tests := []struct {
		name    string
		uri     string
		want    string
		wantErr bool
	}{
		{
			name:    "https with IP",
			uri:     "https://192.168.33.2",
			want:    "192.168.33.2",
			wantErr: false,
		},
		{
			name:    "https with IP and port",
			uri:     "https://192.168.33.2:8443",
			want:    "192.168.33.2",
			wantErr: false,
		},
		{
			name:    "http with host",
			uri:     "http://casa.local",
			want:    "casa.local",
			wantErr: false,
		},
		{
			name:    "empty URI",
			uri:     "",
			want:    "",
			wantErr: true,
		},
		{
			name:    "IPv6 link-local with zone",
			uri:     "https://[fe80::dead:beef:cafe:babe%eth1]",
			want:    "[fe80::dead:beef:cafe:babe%eth1]",
			wantErr: false,
		},
		{
			name:    "IPv6 with port",
			uri:     "https://[fe80::1%eth0]:8080",
			want:    "[fe80::1%eth0]",
			wantErr: false,
		},
		{
			name:    "IPv6 with path",
			uri:     "https://[::1]/api/v1",
			want:    "[::1]",
			wantErr: false,
		},
		{
			name:    "IPv6 without zone",
			uri:     "https://[2001:db8::1]",
			want:    "[2001:db8::1]",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseURIHost(tt.uri)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseURIHost() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("parseURIHost() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestDefaultScheme tests scheme addition
func TestDefaultScheme(t *testing.T) {
	tests := []struct {
		name   string
		uri    string
		scheme string
		want   string
	}{
		{
			name:   "add https to IP",
			uri:    "192.168.33.2",
			scheme: "https",
			want:   "https://192.168.33.2",
		},
		{
			name:   "keep existing https",
			uri:    "https://example.com",
			scheme: "https",
			want:   "https://example.com",
		},
		{
			name:   "keep existing http",
			uri:    "http://example.com",
			scheme: "https",
			want:   "http://example.com",
		},
		{
			name:   "add http",
			uri:    "example.com",
			scheme: "http",
			want:   "http://example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := defaultScheme(tt.uri, tt.scheme)
			if got != tt.want {
				t.Errorf("defaultScheme() = %v, want %v", got, tt.want)
			}
		})
	}
}

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

// TestNewClient tests client creation with validation
func TestNewClient(t *testing.T) {
	tests := []struct {
		name      string
		uri       string
		user      string
		password  string
		meterID   string
		host      string
		wantErr   bool
		errSubstr string
	}{
		{
			name:      "missing URI triggers discovery (will fail without gateway)",
			uri:       "",
			user:      "admin",
			password:  "pass",
			meterID:   "123",
			host:      "192.168.1.1",
			wantErr:   true,
			errSubstr: "failed to discover gateway",
		},
		{
			name:      "missing username",
			uri:       "https://example.com",
			user:      "",
			password:  "pass",
			meterID:   "123",
			host:      "192.168.1.1",
			wantErr:   true,
			errSubstr: "credentials are required",
		},
		{
			name:      "missing password",
			uri:       "https://example.com",
			user:      "admin",
			password:  "",
			meterID:   "123",
			host:      "192.168.1.1",
			wantErr:   true,
			errSubstr: "credentials are required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewClient(tt.uri, tt.user, tt.password, tt.meterID, tt.host)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && !strings.Contains(err.Error(), tt.errSubstr) {
				t.Errorf("NewClient() error = %v, want to contain %v", err, tt.errSubstr)
			}
		})
	}
}
