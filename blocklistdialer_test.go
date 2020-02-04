package httpRelay

import (
	"bytes"
	"net"
	"testing"
)

func TestBlocklistDialerNilDialer(t *testing.T) {
	defer func() {
		if recover() != nil {
			return
		}
		t.FailNow()
	}()
	b := BlocklistDialer{List: make(map[string]struct{}, 0), Dialer: nil}
	b.Dial("tcp", "hello.world")
	t.FailNow()
}

func TestBlocklistDialerDirectDialer(t *testing.T) {
	b := BlocklistDialer{List: make(map[string]struct{}, 0), Dialer: &TestNopDialer{}}
	_, err := b.Dial("tcp", "hello.world")
	if err != nil {
		t.FailNow()
	}
}

func TestBlocklistDialerBlockedAddress(t *testing.T) {
	b := BlocklistDialer{List: map[string]struct{}{
		"hello.world": struct{}{},
	}, Dialer: &TestNopDialer{}}
	_, err := b.Dial("tcp", "hello.world")
	if err == ErrBlockedHost {
		return
	}
	t.FailNow()
}

func TestBlocklistDialerLoadHosts(t *testing.T) {
	hostsFile := []byte("127.0.0.1 localhost\n0.0.0.0 hello.world\n# the next line tests 2 host names for one destination address\n0.0.0.0 hello.world.too hello.world.future\n")
	b := BlocklistDialer{List: make(map[string]struct{}, 0), Dialer: &TestNopDialer{}}
	b.Load(bytes.NewReader(hostsFile))
	_, err := b.Dial("tcp", "hello.world")
	if err != ErrBlockedHost {
		t.FailNow()
	}
	_, err = b.Dial("tcp", "hello.world.too")
	if err != ErrBlockedHost {
		t.FailNow()
	}
	_, err = b.Dial("tcp", "hello.world.future")
	if err != ErrBlockedHost {
		t.FailNow()
	}
	_, err = b.Dial("tcp", "hello.world.past")
	if err != nil {
		t.FailNow()
	}
}

type TestNopDialer struct{}

func (*TestNopDialer) Dial(network, addr string) (net.Conn, error) {
	return nil, nil
}
