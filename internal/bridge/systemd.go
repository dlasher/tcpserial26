package bridge

import (
	"net"

	"github.com/coreos/go-systemd/v22/activation"
	"github.com/coreos/go-systemd/v22/daemon"
)

// getSystemdListeners returns TCP listeners from systemd socket activation.
// Returns nil slice if socket activation is not available.
func getSystemdListeners() ([]net.Listener, error) {
	listeners, err := activation.Listeners()
	if err != nil {
		return nil, err
	}

	var tcpListeners []net.Listener
	for _, l := range listeners {
		if _, ok := l.(*net.TCPListener); ok {
			tcpListeners = append(tcpListeners, l)
		}
	}

	return tcpListeners, nil
}

// NotifyReady sends a readiness notification to systemd.
// This is a no-op if not running under systemd.
func NotifyReady() {
	_, _ = daemon.SdNotify(false, daemon.SdNotifyReady)
}

// NotifyStopping sends a stopping notification to systemd.
func NotifyStopping() {
	_, _ = daemon.SdNotify(false, daemon.SdNotifyStopping)
}

// NotifyStatus sends a status update to systemd.
func NotifyStatus(status string) {
	_, _ = daemon.SdNotify(false, "STATUS="+status)
}
