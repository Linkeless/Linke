package valueobject

import (
	"errors"
	"strings"
)

var (
	ErrEmptyCipher    = errors.New("cipher cannot be empty")
	ErrInvalidCipher  = errors.New("invalid cipher method")
)

// SupportedCiphers defines the list of supported cipher methods
var SupportedCiphers = map[string]bool{
	"aes-128-gcm":        true,
	"aes-192-gcm":        true,
	"aes-256-gcm":        true,
	"aes-128-cfb":        true,
	"aes-192-cfb":        true,
	"aes-256-cfb":        true,
	"aes-128-ctr":        true,
	"aes-192-ctr":        true,
	"aes-256-ctr":        true,
	"chacha20-poly1305":  true,
	"chacha20-ietf-poly1305": true,
	"xchacha20-ietf-poly1305": true,
}

// Cipher represents a Shadowsocks cipher method
type Cipher struct {
	value string
}

// NewCipher creates a new Cipher
func NewCipher(value string) (Cipher, error) {
	value = strings.TrimSpace(strings.ToLower(value))
	
	if value == "" {
		return Cipher{}, ErrEmptyCipher
	}
	
	if !SupportedCiphers[value] {
		return Cipher{}, ErrInvalidCipher
	}
	
	return Cipher{value: value}, nil
}

// Value returns the underlying value
func (c Cipher) Value() string {
	return c.value
}

// String returns string representation
func (c Cipher) String() string {
	return c.value
}

// IsAEAD checks if the cipher is an AEAD cipher
func (c Cipher) IsAEAD() bool {
	aeadCiphers := map[string]bool{
		"aes-128-gcm":             true,
		"aes-192-gcm":             true,
		"aes-256-gcm":             true,
		"chacha20-poly1305":       true,
		"chacha20-ietf-poly1305":  true,
		"xchacha20-ietf-poly1305": true,
	}
	return aeadCiphers[c.value]
}

// IsStreamCipher checks if the cipher is a stream cipher
func (c Cipher) IsStreamCipher() bool {
	return !c.IsAEAD()
}

// Equals checks equality with another Cipher
func (c Cipher) Equals(other Cipher) bool {
	return c.value == other.value
}