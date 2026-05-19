// Package serialport provides serial port detection, enumeration, and management.
package serialport

import (
	"fmt"
	"strings"

	"go.bug.st/serial"
	"go.bug.st/serial/enumerator"

	"github.com/dlasher/tcpserial26/internal/config"
)

// FindDevice locates a serial port based on the configuration.
// If cfg.Device is set, it returns that directly.
// If cfg.VID and cfg.PID are set, it searches for a matching USB device.
// If neither is set, it returns the first available serial port.
func FindDevice(cfg *config.Config) (string, error) {
	// Explicit device path specified
	if cfg.Device != "" {
		return cfg.Device, nil
	}

	// USB VID/PID filtering
	if cfg.HasUSBFilter() {
		vid := strings.TrimPrefix(cfg.VID, "0x")
		pid := strings.TrimPrefix(cfg.PID, "0x")
		return findPortByUSB(vid, pid)
	}

	// Auto-detect: return first available port
	return autoDetectPort()
}

// ListPorts returns a list of all available serial ports with details.
func ListPorts() ([]PortInfo, error) {
	ports, err := enumerator.GetDetailedPortsList()
	if err != nil {
		return nil, fmt.Errorf("failed to enumerate ports: %w", err)
	}

	var infos []PortInfo
	for _, p := range ports {
		info := PortInfo{
			Name:    p.Name,
			IsUSB:   p.IsUSB,
			VID:     p.VID,
			PID:     p.PID,
			Serial:  p.SerialNumber,
			Product: p.Product,
		}
		infos = append(infos, info)
	}
	return infos, nil
}

// PortInfo contains information about a serial port.
type PortInfo struct {
	Name    string
	IsUSB   bool
	VID     string
	PID     string
	Serial  string
	Product string
}

// OpenPort opens a serial port with the given configuration.
func OpenPort(device string, cfg *config.Config) (serial.Port, error) {
	mode := &serial.Mode{
		BaudRate: cfg.BaudRate,
		DataBits: cfg.DataBits,
	}

	// Set parity
	switch cfg.Parity {
	case "none":
		mode.Parity = serial.NoParity
	case "odd":
		mode.Parity = serial.OddParity
	case "even":
		mode.Parity = serial.EvenParity
	case "mark":
		mode.Parity = serial.MarkParity
	case "space":
		mode.Parity = serial.SpaceParity
	}

	// Set stop bits
	switch cfg.StopBits {
	case "1":
		mode.StopBits = serial.OneStopBit
	case "1.5":
		mode.StopBits = serial.OnePointFiveStopBits
	case "2":
		mode.StopBits = serial.TwoStopBits
	}

	port, err := serial.Open(device, mode)
	if err != nil {
		// Check for typed errors
		if pe, ok := err.(*serial.PortError); ok {
			switch pe.Code() {
			case serial.PortNotFound:
				return nil, fmt.Errorf("serial port %s not found", device)
			case serial.PortBusy:
				return nil, fmt.Errorf("serial port %s is already in use", device)
			case serial.PermissionDenied:
				return nil, fmt.Errorf("permission denied for serial port %s (try adding user to dialout/tty group)", device)
			default:
				return nil, fmt.Errorf("failed to open serial port %s: %w", device, err)
			}
		}
		return nil, fmt.Errorf("failed to open serial port %s: %w", device, err)
	}

	return port, nil
}

// findPortByUSB searches for a serial port by USB Vendor ID and Product ID.
func findPortByUSB(vid, pid string) (string, error) {
	ports, err := enumerator.GetDetailedPortsList()
	if err != nil {
		return "", fmt.Errorf("failed to enumerate USB devices: %w", err)
	}

	for _, p := range ports {
		if p.IsUSB && strings.EqualFold(p.VID, vid) && strings.EqualFold(p.PID, pid) {
			return p.Name, nil
		}
	}

	return "", fmt.Errorf("USB device %s:%s not found among available serial ports", vid, pid)
}

// autoDetectPort returns the first available serial port.
func autoDetectPort() (string, error) {
	// First try to get detailed list to prefer USB devices
	ports, err := enumerator.GetDetailedPortsList()
	if err != nil {
		// Fallback to basic port list
		basicPorts, err := serial.GetPortsList()
		if err != nil {
			return "", fmt.Errorf("failed to list serial ports: %w", err)
		}
		if len(basicPorts) == 0 {
			return "", fmt.Errorf("no serial ports found on the system")
		}
		return basicPorts[0], nil
	}

	// Prefer USB serial ports (like /dev/ttyUSB0 or /dev/ttyACM0)
	for _, p := range ports {
		if p.IsUSB {
			return p.Name, nil
		}
	}

	// Fall back to any available port
	if len(ports) == 0 {
		return "", fmt.Errorf("no serial ports found on the system")
	}

	return ports[0].Name, nil
}
