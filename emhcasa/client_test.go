package emhcasa

import (
	"strings"
	"testing"
)

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

// TestNewClient tests client creation with validation
func TestNewClient(t *testing.T) {
	tests := []struct {
		name      string
		uri       string
		user      string
		password  string
		meterID   string
		wantErr   bool
		errSubstr string
	}{
		{
			name:      "missing URI triggers discovery (will fail without gateway)",
			uri:       "",
			user:      "admin",
			password:  "pass",
			meterID:   "123",
			wantErr:   true,
			errSubstr: "failed to discover gateway",
		},
		{
			name:      "missing username",
			uri:       "https://example.com",
			user:      "",
			password:  "pass",
			meterID:   "123",
			wantErr:   true,
			errSubstr: "credentials are required",
		},
		{
			name:      "missing password",
			uri:       "https://example.com",
			user:      "admin",
			password:  "",
			meterID:   "123",
			wantErr:   true,
			errSubstr: "credentials are required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewClient(tt.uri, tt.user, tt.password, tt.meterID)
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
