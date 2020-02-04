package httprelay

import (
	"errors"
	"net"
)

var ErrBlockedHost = errors.New("host is blocked")

type NopDialer struct{}

func (NopDialer) Dial(network, addr string) (net.Conn, error) {
	return nil, ErrBlockedHost
}
