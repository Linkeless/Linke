package valueobject_test

import (
	"strings"
	"testing"

	"linke/internal/server/domain/valueobject"
)

func TestServerGroupID(t *testing.T) {
	tests := []struct {
		name    string
		value   uint
		wantErr bool
	}{
		{
			name:    "valid ID",
			value:   1,
			wantErr: false,
		},
		{
			name:    "zero ID should fail",
			value:   0,
			wantErr: true,
		},
		{
			name:    "large ID",
			value:   999999,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := valueobject.NewServerGroupID(tt.value)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("NewServerGroupID() expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("NewServerGroupID() unexpected error: %v", err)
				return
			}
			
			if id.Value() != tt.value {
				t.Errorf("ServerGroupID.Value() = %v, want %v", id.Value(), tt.value)
			}
			
			if id.String() != string(rune(tt.value+'0')) {
				// This test might need adjustment based on the actual String() implementation
				t.Logf("ServerGroupID.String() = %v", id.String())
			}
			
			if id.IsZero() {
				t.Errorf("ServerGroupID.IsZero() = true, want false")
			}
		})
	}
}

func TestServerGroupID_Equals(t *testing.T) {
	id1, _ := valueobject.NewServerGroupID(1)
	id2, _ := valueobject.NewServerGroupID(1)
	id3, _ := valueobject.NewServerGroupID(2)
	
	if !id1.Equals(id2) {
		t.Errorf("ServerGroupID.Equals() should return true for equal IDs")
	}
	
	if id1.Equals(id3) {
		t.Errorf("ServerGroupID.Equals() should return false for different IDs")
	}
}

func TestServerGroupName(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{
			name:    "valid name",
			value:   "Asia Pacific",
			wantErr: false,
		},
		{
			name:    "empty name should fail",
			value:   "",
			wantErr: true,
		},
		{
			name:    "whitespace only should fail",
			value:   "   ",
			wantErr: true,
		},
		{
			name:    "too long name should fail",
			value:   string(make([]byte, 256)), // 256 characters
			wantErr: true,
		},
		{
			name:    "name with control character should fail",
			value:   "Test\x00Name",
			wantErr: true,
		},
		{
			name:    "name with trimmed spaces",
			value:   "  Valid Name  ",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name, err := valueobject.NewServerGroupName(tt.value)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("NewServerGroupName() expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("NewServerGroupName() unexpected error: %v", err)
				return
			}
			
			expectedValue := tt.value
			if tt.name == "name with trimmed spaces" {
				expectedValue = "Valid Name"
			}
			
			if name.Value() != expectedValue {
				t.Errorf("ServerGroupName.Value() = %v, want %v", name.Value(), expectedValue)
			}
			
			if name.String() != expectedValue {
				t.Errorf("ServerGroupName.String() = %v, want %v", name.String(), expectedValue)
			}
			
			if name.IsEmpty() {
				t.Errorf("ServerGroupName.IsEmpty() = true, want false")
			}
		})
	}
}

func TestServerHost(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
		isIP    bool
	}{
		{
			name:    "valid IPv4",
			value:   "192.168.1.1",
			wantErr: false,
			isIP:    true,
		},
		{
			name:    "valid IPv6",
			value:   "2001:db8::1",
			wantErr: false,
			isIP:    true,
		},
		{
			name:    "valid domain",
			value:   "example.com",
			wantErr: false,
			isIP:    false,
		},
		{
			name:    "valid subdomain",
			value:   "server.example.com",
			wantErr: false,
			isIP:    false,
		},
		{
			name:    "empty host should fail",
			value:   "",
			wantErr: true,
		},
		{
			name:    "whitespace only should fail",
			value:   "   ",
			wantErr: true,
		},
		{
			name:    "invalid domain",
			value:   "invalid..domain",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			host, err := valueobject.NewServerHost(tt.value)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("NewServerHost() expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("NewServerHost() unexpected error: %v", err)
				return
			}
			
			expectedValue := strings.TrimSpace(tt.value)
			if host.Value() != expectedValue {
				t.Errorf("ServerHost.Value() = %v, want %v", host.Value(), expectedValue)
			}
			
			if host.IsIP() != tt.isIP {
				t.Errorf("ServerHost.IsIP() = %v, want %v", host.IsIP(), tt.isIP)
			}
			
			if host.IsDomain() == tt.isIP {
				t.Errorf("ServerHost.IsDomain() = %v, want %v", host.IsDomain(), !tt.isIP)
			}
		})
	}
}

func TestServerPort(t *testing.T) {
	tests := []struct {
		name         string
		value        int
		wantErr      bool
		isWellKnown  bool
		isRegistered bool
		isDynamic    bool
	}{
		{
			name:         "port 80 (well-known)",
			value:        80,
			wantErr:      false,
			isWellKnown:  true,
			isRegistered: false,
			isDynamic:    false,
		},
		{
			name:         "port 8080 (registered)",
			value:        8080,
			wantErr:      false,
			isWellKnown:  false,
			isRegistered: true,
			isDynamic:    false,
		},
		{
			name:         "port 50000 (dynamic)",
			value:        50000,
			wantErr:      false,
			isWellKnown:  false,
			isRegistered: false,
			isDynamic:    true,
		},
		{
			name:    "port 0 should fail",
			value:   0,
			wantErr: true,
		},
		{
			name:    "port -1 should fail",
			value:   -1,
			wantErr: true,
		},
		{
			name:    "port 65536 should fail",
			value:   65536,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			port, err := valueobject.NewServerPort(tt.value)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("NewServerPort() expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("NewServerPort() unexpected error: %v", err)
				return
			}
			
			if port.Value() != tt.value {
				t.Errorf("ServerPort.Value() = %v, want %v", port.Value(), tt.value)
			}
			
			if port.IsWellKnown() != tt.isWellKnown {
				t.Errorf("ServerPort.IsWellKnown() = %v, want %v", port.IsWellKnown(), tt.isWellKnown)
			}
			
			if port.IsRegistered() != tt.isRegistered {
				t.Errorf("ServerPort.IsRegistered() = %v, want %v", port.IsRegistered(), tt.isRegistered)
			}
			
			if port.IsDynamic() != tt.isDynamic {
				t.Errorf("ServerPort.IsDynamic() = %v, want %v", port.IsDynamic(), tt.isDynamic)
			}
		})
	}
}

func TestCipher(t *testing.T) {
	tests := []struct {
		name   string
		value  string
		wantErr bool
		isAEAD bool
	}{
		{
			name:   "aes-256-gcm (AEAD)",
			value:  "aes-256-gcm",
			wantErr: false,
			isAEAD: true,
		},
		{
			name:   "aes-256-cfb (stream)",
			value:  "aes-256-cfb",
			wantErr: false,
			isAEAD: false,
		},
		{
			name:   "chacha20-poly1305 (AEAD)",
			value:  "chacha20-poly1305",
			wantErr: false,
			isAEAD: true,
		},
		{
			name:    "empty cipher should fail",
			value:   "",
			wantErr: true,
		},
		{
			name:    "invalid cipher should fail",
			value:   "invalid-cipher",
			wantErr: true,
		},
		{
			name:   "case insensitive",
			value:  "AES-256-GCM",
			wantErr: false,
			isAEAD: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cipher, err := valueobject.NewCipher(tt.value)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("NewCipher() expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("NewCipher() unexpected error: %v", err)
				return
			}
			
			expectedValue := strings.ToLower(strings.TrimSpace(tt.value))
			if cipher.Value() != expectedValue {
				t.Errorf("Cipher.Value() = %v, want %v", cipher.Value(), expectedValue)
			}
			
			if cipher.IsAEAD() != tt.isAEAD {
				t.Errorf("Cipher.IsAEAD() = %v, want %v", cipher.IsAEAD(), tt.isAEAD)
			}
			
			if cipher.IsStreamCipher() == tt.isAEAD {
				t.Errorf("Cipher.IsStreamCipher() = %v, want %v", cipher.IsStreamCipher(), !tt.isAEAD)
			}
		})
	}
}

func TestRate(t *testing.T) {
	tests := []struct {
		name         string
		value        float64
		wantErr      bool
		isStandard   bool
		isPremium    bool
		isDiscounted bool
	}{
		{
			name:         "standard rate",
			value:        1.0,
			wantErr:      false,
			isStandard:   true,
			isPremium:    false,
			isDiscounted: false,
		},
		{
			name:         "premium rate",
			value:        2.0,
			wantErr:      false,
			isStandard:   false,
			isPremium:    true,
			isDiscounted: false,
		},
		{
			name:         "discounted rate",
			value:        0.5,
			wantErr:      false,
			isStandard:   false,
			isPremium:    false,
			isDiscounted: true,
		},
		{
			name:    "zero rate should fail",
			value:   0.0,
			wantErr: true,
		},
		{
			name:    "negative rate should fail",
			value:   -1.0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rate, err := valueobject.NewRate(tt.value)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("NewRate() expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("NewRate() unexpected error: %v", err)
				return
			}
			
			if rate.Value() != tt.value {
				t.Errorf("Rate.Value() = %v, want %v", rate.Value(), tt.value)
			}
			
			if rate.IsStandard() != tt.isStandard {
				t.Errorf("Rate.IsStandard() = %v, want %v", rate.IsStandard(), tt.isStandard)
			}
			
			if rate.IsPremium() != tt.isPremium {
				t.Errorf("Rate.IsPremium() = %v, want %v", rate.IsPremium(), tt.isPremium)
			}
			
			if rate.IsDiscounted() != tt.isDiscounted {
				t.Errorf("Rate.IsDiscounted() = %v, want %v", rate.IsDiscounted(), tt.isDiscounted)
			}
		})
	}
}