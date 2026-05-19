// Package main provides the entry point for the tcpserial application.
// tcpserial is a TCP-to-serial bridge designed to work with Z-Wave JS UI.
package main

import (
	"fmt"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.bug.st/serial"
	"go.uber.org/zap"

	"github.com/dlasher/tcpserial26/internal/bridge"
	"github.com/dlasher/tcpserial26/internal/config"
	"github.com/dlasher/tcpserial26/internal/serialport"
)

var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

func main() {
	cmd := NewRootCmd()
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// NewRootCmd creates the root CLI command.
func NewRootCmd() *cobra.Command {
	cfg := config.DefaultConfig()

	cmd := &cobra.Command{
		Use:   "tcpserial",
		Short: "TCP-to-serial bridge for Z-Wave JS UI",
		Long: `tcpserial is a robust TCP-to-serial bridge designed to expose serial devices
over TCP for use with Z-Wave JS UI and similar applications.

Features:
- Automatic serial port detection by USB VID/PID
- Configurable buffer sizes and timeouts
- Automatic reconnection on device disconnect
- systemd socket activation support
- IP whitelist for access control

Examples:
  # Auto-detect serial port and listen on port 2000
  tcpserial

  # Specify device and baud rate
  tcpserial --device /dev/ttyUSB0 --baud 115200

  # Auto-detect by USB VID/PID
  tcpserial --vid 10C4 --pid EA60

  # Listen on specific address
  tcpserial --address 192.168.1.100 --port 6638`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return cfg.Validate()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd, &cfg)
		},
	}

	// TCP server flags
	cmd.Flags().StringVarP(&cfg.Address, "address", "a", cfg.Address, "Listen address")
	cmd.Flags().IntVarP(&cfg.Port, "port", "p", cfg.Port, "Listen port")
	cmd.Flags().StringSliceVar(&cfg.IPWhitelist, "whitelist", cfg.IPWhitelist, "IP whitelist (CIDR notation, repeatable)")

	// Serial port flags
	cmd.Flags().StringVarP(&cfg.Device, "device", "d", cfg.Device, "Serial device path (empty for auto-detect)")
	cmd.Flags().StringVar(&cfg.VID, "vid", cfg.VID, "USB Vendor ID for auto-detection (e.g., 10C4)")
	cmd.Flags().StringVar(&cfg.PID, "pid", cfg.PID, "USB Product ID for auto-detection (e.g., EA60)")
	cmd.Flags().IntVarP(&cfg.BaudRate, "baud", "b", cfg.BaudRate, "Baud rate")
	cmd.Flags().IntVar(&cfg.DataBits, "data-bits", cfg.DataBits, "Data bits (5-8)")
	cmd.Flags().StringVar(&cfg.Parity, "parity", cfg.Parity, "Parity: none, odd, even, mark, space")
	cmd.Flags().StringVar(&cfg.StopBits, "stop-bits", cfg.StopBits, "Stop bits: 1, 1.5, 2")

	// Buffer flags
	cmd.Flags().IntVar(&cfg.ReadBuffers, "read-buffers", cfg.ReadBuffers, "Number of read buffers (1-32)")
	cmd.Flags().IntVarP(&cfg.BufferSize, "buffer-size", "s", cfg.BufferSize, "TCP buffer size in bytes (1-10MB)")

	// Timeout flags
	cmd.Flags().DurationVar(&cfg.ReadTimeout, "read-timeout", cfg.ReadTimeout, "Socket read timeout")
	cmd.Flags().DurationVar(&cfg.WriteTimeout, "write-timeout", cfg.WriteTimeout, "Socket write timeout")

	// Retry flags
	cmd.Flags().DurationVar(&cfg.RetryInterval, "retry-interval", cfg.RetryInterval, "Retry interval for reconnection")
	cmd.Flags().IntVar(&cfg.MaxRetries, "max-retries", cfg.MaxRetries, "Max retry attempts (0 = infinite)")

	// Other flags
	cmd.Flags().StringVar(&cfg.LogLevel, "log-level", cfg.LogLevel, "Log level: debug, info, warn, error")
	cmd.Flags().StringVar(&cfgFile, "config", "", "Config file path")

	// Bind to viper for config file support
	_ = viper.BindPFlags(cmd.Flags())

	// Version flag
	cmd.Version = fmt.Sprintf("%s (commit: %s, built: %s)", version, commit, date)

	return cmd
}

var cfgFile string

func init() {
	cobra.OnInitialize(initConfig)
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
		if err := viper.ReadInConfig(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not read config file: %v\n", err)
		}
	}

	viper.SetConfigName("tcpserial")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/etc/tcpserial")
	viper.AddConfigPath("$HOME/.tcpserial")
	viper.AddConfigPath(".")

	_ = viper.ReadInConfig()
	viper.AutomaticEnv()
}

func run(cmd *cobra.Command, cfg *config.Config) error {
	// Setup logger
	logger, err := bridge.SetupLogger(cfg.LogLevel)
	if err != nil {
		return fmt.Errorf("failed to setup logger: %w", err)
	}
	defer logger.Sync()

	logger.Info("tcpserial starting",
		zap.String("version", version),
		zap.String("commit", commit),
	)

	// Auto-detect serial device if not specified
	if cfg.Device == "" {
		if cfg.HasUSBFilter() {
			logger.Info("searching for USB device",
				zap.String("vid", cfg.VID),
				zap.String("pid", cfg.PID),
			)
		} else {
			logger.Info("auto-detecting serial port")
		}

		device, err := serialport.FindDevice(cfg)
		if err != nil {
			return fmt.Errorf("failed to find serial device: %w", err)
		}
		cfg.Device = device
		logger.Info("found serial device", zap.String("device", device))
	}

	// Log configuration
	logger.Info("configuration",
		zap.String("device", cfg.Device),
		zap.String("address", cfg.Address),
		zap.Int("port", cfg.Port),
		zap.Int("baud_rate", cfg.BaudRate),
		zap.Int("buffer_size", cfg.BufferSize),
	)

	// Signal handling
	stopCh := make(chan struct{})
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	var shuttingDown atomic.Bool

	go func() {
		sig := <-sigCh
		logger.Info("received signal, shutting down", zap.String("signal", sig.String()))
		shuttingDown.Store(true)
		close(stopCh)
	}()

	// Create bridge
	b := bridge.New(cfg, logger)

	// Notify systemd we're ready
	bridge.NotifyReady()

	// Run with auto-reconnect
	err = runWithReconnect(b, cfg, logger, stopCh, &shuttingDown)

	// Notify systemd we're stopping
	bridge.NotifyStopping()

	if err != nil {
		logger.Error("tcpserial stopped with error", zap.Error(err))
	} else {
		logger.Info("tcpserial stopped gracefully")
	}

	return nil
}

// runWithReconnect runs the bridge with automatic reconnection on serial port errors.
func runWithReconnect(b *bridge.Bridge, cfg *config.Config, logger *zap.Logger, stopCh <-chan struct{}, shuttingDown *atomic.Bool) error {
	retries := 0

	for {
		if shuttingDown.Load() {
			return nil
		}

		bridge.NotifyStatus("Connecting to serial device...")

		// Create a closure that opens the serial port
		openPort := func() (serial.Port, error) {
			return serialport.OpenPort(cfg.Device, cfg)
		}

		err := b.Run(openPort, stopCh)

		// Check if we're shutting down
		if shuttingDown.Load() {
			return nil
		}

		if err != nil {
			logger.Error("bridge error", zap.Error(err))
		}

		// Check retry limits
		if cfg.MaxRetries > 0 && retries >= cfg.MaxRetries {
			return fmt.Errorf("max retries (%d) exceeded", cfg.MaxRetries)
		}

		retries++
		bridge.NotifyStatus(fmt.Sprintf("Reconnecting in %s (attempt %d)...", cfg.RetryInterval, retries))

		logger.Info("reconnecting",
			zap.Duration("interval", cfg.RetryInterval),
			zap.Int("attempt", retries),
			zap.Int("max_retries", cfg.MaxRetries),
		)

		// Wait for retry interval or shutdown
		select {
		case <-time.After(cfg.RetryInterval):
			// Continue with reconnection
		case <-stopCh:
			return nil
		}
	}
}
