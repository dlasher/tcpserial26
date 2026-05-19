// Package config provides configuration management for tcpserial.
// It supports CLI flags, environment variables, and config files.
package config

import (
	"fmt"
	"time"
)

// Config holds all configuration for the tcpserial application.
type Config struct {
	// TCP server settings
	Address     string
	Port        int
	IPWhitelist []string

	// Serial port settings
	Device   string
	VID      string // USB Vendor ID for auto-detection
	PID      string // USB Product ID for auto-detection
	BaudRate int
	DataBits int
	Parity   string // "none", "odd", "even", "mark", "space"
	StopBits string // "1", "1.5", "2"

	// Buffer settings
	ReadBuffers int
	BufferSize  int

	// Timeout settings
	ReadTimeout  time.Duration
	WriteTimeout time.Duration

	// Retry settings
	RetryInterval time.Duration
	MaxRetries    int // 0 = infinite retries

	// Logging
	LogLevel string
}

// DefaultConfig returns a Config populated with sensible defaults for Z-Wave devices.
func DefaultConfig() Config {
	return Config{
		Address:       "0.0.0.0",
		Port:          2000,
		IPWhitelist:   []string{},
		Device:        "",
		VID:           "",
		PID:           "",
		BaudRate:      115200,
		DataBits:      8,
		Parity:        "none",
		StopBits:      "1",
		ReadBuffers:   15,
		BufferSize:    512000,
		ReadTimeout:   30 * time.Second,
		WriteTimeout:  30 * time.Second,
		RetryInterval: 5 * time.Second,
		MaxRetries:    0,
		LogLevel:      "info",
	}
}

// Validate checks the configuration for invalid values.
func (c *Config) Validate() error {
	if c.Port < 1 || c.Port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535, got %d", c.Port)
	}
	if c.BaudRate < 1 {
		return fmt.Errorf("baud rate must be positive, got %d", c.BaudRate)
	}
	if c.DataBits < 5 || c.DataBits > 8 {
		return fmt.Errorf("data bits must be between 5 and 8, got %d", c.DataBits)
	}
	switch c.Parity {
	case "none", "odd", "even", "mark", "space":
		// valid
	default:
		return fmt.Errorf("invalid parity %q, must be one of: none, odd, even, mark, space", c.Parity)
	}
	switch c.StopBits {
	case "1", "1.5", "2":
		// valid
	default:
		return fmt.Errorf("invalid stop bits %q, must be one of: 1, 1.5, 2", c.StopBits)
	}
	if c.ReadBuffers < 1 || c.ReadBuffers > 32 {
		return fmt.Errorf("read buffers must be between 1 and 32, got %d", c.ReadBuffers)
	}
	if c.BufferSize < 1 || c.BufferSize > 10485760 {
		return fmt.Errorf("buffer size must be between 1 and 10485760 (10MB), got %d", c.BufferSize)
	}
	if c.ReadTimeout < 1*time.Second {
		return fmt.Errorf("read timeout must be at least 1 second")
	}
	if c.WriteTimeout < 1*time.Second {
		return fmt.Errorf("write timeout must be at least 1 second")
	}
	if c.RetryInterval < 1*time.Second {
		return fmt.Errorf("retry interval must be at least 1 second")
	}
	if c.MaxRetries < 0 {
		return fmt.Errorf("max retries must be >= 0 (0 = infinite)")
	}
	switch c.LogLevel {
	case "debug", "info", "warn", "error":
		// valid
	default:
		return fmt.Errorf("invalid log level %q, must be one of: debug, info, warn, error", c.LogLevel)
	}
	if c.Device == "" && (c.VID != "" || c.PID != "") {
		if c.VID == "" || c.PID == "" {
			return fmt.Errorf("both VID and PID must be specified for USB auto-detection")
		}
	}
	return nil
}

// HasUSBFilter returns true if USB VID/PID filtering is configured.
func (c *Config) HasUSBFilter() bool {
	return c.VID != "" && c.PID != ""
}
