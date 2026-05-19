package config

import (
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Address != "0.0.0.0" {
		t.Errorf("expected default address 0.0.0.0, got %s", cfg.Address)
	}
	if cfg.Port != 2000 {
		t.Errorf("expected default port 2000, got %d", cfg.Port)
	}
	if cfg.BaudRate != 115200 {
		t.Errorf("expected default baud rate 115200, got %d", cfg.BaudRate)
	}
	if cfg.MaxRetries != 0 {
		t.Errorf("expected default max retries 0 (infinite), got %d", cfg.MaxRetries)
	}
}

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		modify  func(*Config)
		wantErr bool
	}{
		{
			name:    "valid default config",
			modify:  func(c *Config) {},
			wantErr: false,
		},
		{
			name:    "invalid port zero",
			modify:  func(c *Config) { c.Port = 0 },
			wantErr: true,
		},
		{
			name:    "invalid port too high",
			modify:  func(c *Config) { c.Port = 65536 },
			wantErr: true,
		},
		{
			name:    "invalid baud rate",
			modify:  func(c *Config) { c.BaudRate = 0 },
			wantErr: true,
		},
		{
			name:    "invalid data bits low",
			modify:  func(c *Config) { c.DataBits = 4 },
			wantErr: true,
		},
		{
			name:    "invalid data bits high",
			modify:  func(c *Config) { c.DataBits = 9 },
			wantErr: true,
		},
		{
			name:    "invalid parity",
			modify:  func(c *Config) { c.Parity = "invalid" },
			wantErr: true,
		},
		{
			name:    "valid odd parity",
			modify:  func(c *Config) { c.Parity = "odd" },
			wantErr: false,
		},
		{
			name:    "valid even parity",
			modify:  func(c *Config) { c.Parity = "even" },
			wantErr: false,
		},
		{
			name:    "invalid stop bits",
			modify:  func(c *Config) { c.StopBits = "3" },
			wantErr: true,
		},
		{
			name:    "valid 1.5 stop bits",
			modify:  func(c *Config) { c.StopBits = "1.5" },
			wantErr: false,
		},
		{
			name:    "invalid read buffers zero",
			modify:  func(c *Config) { c.ReadBuffers = 0 },
			wantErr: true,
		},
		{
			name:    "invalid read buffers too high",
			modify:  func(c *Config) { c.ReadBuffers = 33 },
			wantErr: true,
		},
		{
			name:    "invalid buffer size zero",
			modify:  func(c *Config) { c.BufferSize = 0 },
			wantErr: true,
		},
		{
			name:    "invalid buffer size too large",
			modify:  func(c *Config) { c.BufferSize = 10485761 },
			wantErr: true,
		},
		{
			name:    "invalid read timeout",
			modify:  func(c *Config) { c.ReadTimeout = 500 * time.Millisecond },
			wantErr: true,
		},
		{
			name:    "invalid write timeout",
			modify:  func(c *Config) { c.WriteTimeout = 0 },
			wantErr: true,
		},
		{
			name:    "invalid retry interval",
			modify:  func(c *Config) { c.RetryInterval = 0 },
			wantErr: true,
		},
		{
			name:    "invalid max retries negative",
			modify:  func(c *Config) { c.MaxRetries = -1 },
			wantErr: true,
		},
		{
			name:    "invalid log level",
			modify:  func(c *Config) { c.LogLevel = "verbose" },
			wantErr: true,
		},
		{
			name:    "valid debug log level",
			modify:  func(c *Config) { c.LogLevel = "debug" },
			wantErr: false,
		},
		{
			name:    "vid set but pid missing",
			modify:  func(c *Config) { c.VID = "10C4"; c.PID = "" },
			wantErr: true,
		},
		{
			name:    "pid set but vid missing",
			modify:  func(c *Config) { c.VID = ""; c.PID = "EA60" },
			wantErr: true,
		},
		{
			name:    "both vid and pid set",
			modify:  func(c *Config) { c.VID = "10C4"; c.PID = "EA60" },
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			tt.modify(&cfg)
			err := cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHasUSBFilter(t *testing.T) {
	tests := []struct {
		name string
		vid  string
		pid  string
		want bool
	}{
		{"both set", "10C4", "EA60", true},
		{"vid only", "10C4", "", false},
		{"pid only", "", "EA60", false},
		{"neither set", "", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			cfg.VID = tt.vid
			cfg.PID = tt.pid
			if got := cfg.HasUSBFilter(); got != tt.want {
				t.Errorf("HasUSBFilter() = %v, want %v", got, tt.want)
			}
		})
	}
}
