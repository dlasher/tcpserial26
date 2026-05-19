// Package bridge provides TCP-to-serial bridging logic.
package bridge

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/coreos/go-systemd/v22/daemon"
	"go.bug.st/serial"
	"go.uber.org/zap"

	"github.com/dlasher/tcpserial26/internal/config"
)

// Bridge manages a TCP-to-serial bridge connection.
type Bridge struct {
	cfg    *config.Config
	logger *zap.Logger
}

// New creates a new Bridge.
func New(cfg *config.Config, logger *zap.Logger) *Bridge {
	return &Bridge{
		cfg:    cfg,
		logger: logger,
	}
}

// Run starts the TCP server and bridges connections to the serial port.
// It blocks until a shutdown signal is received or a fatal error occurs.
func (b *Bridge) Run(openPort func() (serial.Port, error), stopCh <-chan struct{}) error {
	listener, err := b.startTCP()
	if err != nil {
		return err
	}
	defer listener.Close()

	b.logger.Info("TCP server started",
		zap.String("address", b.cfg.Address),
		zap.Int("port", b.cfg.Port),
	)

	// Notify systemd we're ready to accept connections
	_, _ = daemon.SdNotify(false, daemon.SdNotifyReady)

	for {
		select {
		case <-stopCh:
			b.logger.Info("shutdown signal received, stopping TCP server")
			return nil
		default:
		}

		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-stopCh:
				return nil
			default:
				b.logger.Error("failed to accept connection", zap.Error(err))
				time.Sleep(100 * time.Millisecond)
				continue
			}
		}

		// Check IP whitelist if configured
		if len(b.cfg.IPWhitelist) > 0 {
			remoteAddr := conn.RemoteAddr().String()
			host, _, err := net.SplitHostPort(remoteAddr)
			if err != nil {
				host = remoteAddr
			}
			if !isIPWhitelisted(host, b.cfg.IPWhitelist) {
				b.logger.Warn("client rejected - not in whitelist", zap.String("remote", remoteAddr))
				_ = conn.Close()
				continue
			}
		}

		b.logger.Info("client connected", zap.String("remote", conn.RemoteAddr().String()))
		_, _ = daemon.SdNotify(false, "STATUS=Client connected: "+conn.RemoteAddr().String())

		if err := b.handleConnection(conn, openPort, stopCh); err != nil {
			b.logger.Error("connection error", zap.Error(err))
		} else {
			b.logger.Info("client disconnected", zap.String("remote", conn.RemoteAddr().String()))
		}
	}
}

// startTCP creates and configures the TCP listener.
func (b *Bridge) startTCP() (net.Listener, error) {
	// Try systemd socket activation first
	listeners, err := getSystemdListeners()
	if err != nil {
		b.logger.Debug("systemd socket activation not available", zap.Error(err))
	}
	if len(listeners) > 0 {
		b.logger.Info("using systemd socket activation")
		return listeners[0], nil
	}

	// Fall back to manual binding
	addr := fmt.Sprintf("%s:%d", b.cfg.Address, b.cfg.Port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to bind to %s: %w", addr, err)
	}

	return listener, nil
}

// handleConnection bridges a single TCP connection to the serial port.
func (b *Bridge) handleConnection(tcpConn net.Conn, openPort func() (serial.Port, error), stopCh <-chan struct{}) error {
	defer tcpConn.Close()

	// Set timeouts
	if err := tcpConn.SetDeadline(time.Time{}); err != nil {
		b.logger.Error("failed to clear deadline", zap.Error(err))
	}

	// Open serial port
	port, err := openPort()
	if err != nil {
		return fmt.Errorf("failed to open serial port: %w", err)
	}
	defer port.Close()

	b.logger.Info("serial port opened", zap.String("device", b.cfg.Device))

	// Bridge TCP <-> Serial
	return b.bidirectionalCopy(tcpConn, port, stopCh)
}

// bidirectionalCopy copies data between TCP and serial in both directions.
func (b *Bridge) bidirectionalCopy(tcpConn net.Conn, port serial.Port, stopCh <-chan struct{}) error {
	var wg sync.WaitGroup
	wg.Add(2)

	errChan := make(chan error, 2)

	// Serial -> TCP
	go func() {
		defer wg.Done()
		buf := make([]byte, b.cfg.BufferSize)
		_, err := io.CopyBuffer(tcpConn, port, buf)
		if err != nil {
			errChan <- fmt.Errorf("serial->tcp copy failed: %w", err)
		}
		// Signal no more data to TCP client
		if tcp, ok := tcpConn.(*net.TCPConn); ok {
			_ = tcp.CloseWrite()
		}
	}()

	// TCP -> Serial
	go func() {
		defer wg.Done()
		buf := make([]byte, b.cfg.BufferSize)
		_, err := io.CopyBuffer(port, tcpConn, buf)
		if err != nil {
			errChan <- fmt.Errorf("tcp->serial copy failed: %w", err)
		}
		// Serial ports don't support CloseWrite
	}()

	// Wait for either goroutine to finish or shutdown signal
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Both copies completed
		select {
		case err := <-errChan:
			b.logger.Debug("bridge copy error", zap.Error(err))
		default:
		}
		return nil
	case err := <-errChan:
		return err
	case <-stopCh:
		b.logger.Info("shutdown requested during bridge operation")
		return nil
	}
}
