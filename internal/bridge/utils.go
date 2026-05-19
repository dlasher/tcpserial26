package bridge

import (
	"net"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// isIPWhitelisted checks if an IP address is in the whitelist (CIDR notation).
func isIPWhitelisted(ip string, whitelist []string) bool {
	clientIP := net.ParseIP(ip)
	if clientIP == nil {
		return false
	}

	for _, cidr := range whitelist {
		_, ipNet, err := net.ParseCIDR(cidr)
		if err != nil {
			// Invalid CIDR, skip
			continue
		}
		if ipNet.Contains(clientIP) {
			return true
		}
	}

	return false
}

// SetupLogger creates a configured zap logger with the given level.
func SetupLogger(level string) (*zap.Logger, error) {
	zapLevel, err := zap.ParseAtomicLevel(level)
	if err != nil {
		zapLevel = zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	config := zap.Config{
		Level:       zapLevel,
		Development: false,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding:         "console",
		EncoderConfig:    encoderConfig,
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
	}

	return config.Build()
}
