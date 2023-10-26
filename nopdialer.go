package httprelay

import (
	"net"

	"github.com/cobratbq/goutils/std/errors"
)

// NopDialer does not perform dialing operation as host is blocked.
type NopDialer struct{}

// Dial performs no-op dial operation.
func (NopDialer) Dial(network, addr string) (net.Conn, error) {
	return nil, ErrBlockedHost
}

// ErrBlockedHost indicates that host is blocked.
var ErrBlockedHost = errors.NewStringError("host is blocked")
