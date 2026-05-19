package bridge

import (
	"testing"
)

func TestIsIPWhitelisted(t *testing.T) {
	tests := []struct {
		name      string
		ip        string
		whitelist []string
		want      bool
	}{
		{
			name:      "single IP match",
			ip:        "192.168.1.100",
			whitelist: []string{"192.168.1.100/32"},
			want:      true,
		},
		{
			name:      "single IP no match",
			ip:        "192.168.1.101",
			whitelist: []string{"192.168.1.100/32"},
			want:      false,
		},
		{
			name:      "CIDR /24 match",
			ip:        "192.168.100.50",
			whitelist: []string{"192.168.100.0/24"},
			want:      true,
		},
		{
			name:      "CIDR /24 no match",
			ip:        "192.168.101.50",
			whitelist: []string{"192.168.100.0/24"},
			want:      false,
		},
		{
			name:      "CIDR /16 match",
			ip:        "10.0.255.255",
			whitelist: []string{"10.0.0.0/16"},
			want:      true,
		},
		{
			name:      "CIDR /8 match",
			ip:        "10.50.100.200",
			whitelist: []string{"10.0.0.0/8"},
			want:      true,
		},
		{
			name:      "multiple entries match first",
			ip:        "192.168.1.50",
			whitelist: []string{"192.168.1.0/24", "10.0.0.0/8"},
			want:      true,
		},
		{
			name:      "multiple entries match second",
			ip:        "10.50.100.200",
			whitelist: []string{"192.168.1.0/24", "10.0.0.0/8"},
			want:      true,
		},
		{
			name:      "multiple entries no match",
			ip:        "172.16.0.1",
			whitelist: []string{"192.168.1.0/24", "10.0.0.0/8"},
			want:      false,
		},
		{
			name:      "empty whitelist",
			ip:        "192.168.1.1",
			whitelist: []string{},
			want:      false,
		},
		{
			name:      "invalid CIDR skipped",
			ip:        "192.168.1.1",
			whitelist: []string{"invalid", "192.168.1.0/24"},
			want:      true,
		},
		{
			name:      "localhost",
			ip:        "127.0.0.1",
			whitelist: []string{"127.0.0.0/8"},
			want:      true,
		},
		{
			name:      "all interfaces",
			ip:        "0.0.0.0",
			whitelist: []string{"0.0.0.0/0"},
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isIPWhitelisted(tt.ip, tt.whitelist); got != tt.want {
				t.Errorf("isIPWhitelisted(%q, %v) = %v, want %v", tt.ip, tt.whitelist, got, tt.want)
			}
		})
	}
}
