package httpRelay

import (
	"errors"
	"net"
)

var ErrBlockedTarget = errors.New("target is blocked")

type NopDialer struct{}

func (NopDialer) Dial(network, addr string) (net.Conn, error) {
	return nil, ErrBlockedTarget
}
