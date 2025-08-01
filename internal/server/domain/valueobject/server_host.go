package valueobject

import (
	"errors"
	"net"
	"strings"
)

var (
	ErrEmptyServerHost    = errors.New("server host cannot be empty")
	ErrInvalidServerHost  = errors.New("invalid server host format")
)

// ServerHost represents a server host (IP address or domain name)
type ServerHost struct {
	value string
}

// NewServerHost creates a new ServerHost
func NewServerHost(value string) (ServerHost, error) {
	value = strings.TrimSpace(value)
	
	if value == "" {
		return ServerHost{}, ErrEmptyServerHost
	}
	
	// Validate if it's a valid IP address or domain name
	if net.ParseIP(value) == nil {
		// If not a valid IP, check if it's a valid domain name
		if !isValidDomain(value) {
			return ServerHost{}, ErrInvalidServerHost
		}
	}
	
	return ServerHost{value: value}, nil
}

// Value returns the underlying value
func (host ServerHost) Value() string {
	return host.value
}

// String returns string representation
func (host ServerHost) String() string {
	return host.value
}

// IsIP checks if the host is an IP address
func (host ServerHost) IsIP() bool {
	return net.ParseIP(host.value) != nil
}

// IsDomain checks if the host is a domain name
func (host ServerHost) IsDomain() bool {
	return !host.IsIP()
}

// Equals checks equality with another ServerHost
func (host ServerHost) Equals(other ServerHost) bool {
	return host.value == other.value
}

// isValidDomain performs basic domain name validation
func isValidDomain(domain string) bool {
	if len(domain) == 0 || len(domain) > 255 {
		return false
	}
	
	if domain[len(domain)-1] == '.' {
		domain = domain[:len(domain)-1]
	}
	
	labels := strings.Split(domain, ".")
	if len(labels) == 0 {
		return false
	}
	
	for _, label := range labels {
		if len(label) == 0 || len(label) > 63 {
			return false
		}
		
		for i, r := range label {
			if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || 
			     (r >= '0' && r <= '9') || (r == '-' && i > 0 && i < len(label)-1)) {
				return false
			}
		}
	}
	
	return true
}