package httprelay

import (
	"net"

	"github.com/cobratbq/goutils/std/errors"
)

// ErrBlockedHost indicates that host is blocked.
const ErrBlockedHost errors.StringError = "host is blocked"

// NopDialer does not perform dialing operation as host is blocked.
type NopDialer struct{}

// Dial performs no-op dial operation.
func (NopDialer) Dial(network, addr string) (net.Conn, error) {
	return nil, ErrBlockedHost
}
