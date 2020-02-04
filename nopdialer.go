package httprelay

import (
	"errors"
	"net"
)

// ErrBlockedHost indicates that host is blocked.
var ErrBlockedHost = errors.New("host is blocked")

// NopDialer does not perform dialing operation as host is blocked.
type NopDialer struct{}

// Dial performs no-op dial operation.
func (NopDialer) Dial(network, addr string) (net.Conn, error) {
	return nil, ErrBlockedHost
}
