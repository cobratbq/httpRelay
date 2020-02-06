package httprelay

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
	b.Dial("tcp", "hello.world:80")
	t.FailNow()
}

func TestBlocklistDialerDirectDialer(t *testing.T) {
	b := BlocklistDialer{List: make(map[string]struct{}, 0), Dialer: &TestNopDialer{}}
	_, err := b.Dial("tcp", "hello.world:80")
	if err != nil {
		t.FailNow()
	}
}

func TestBlocklistDialerBlockedAddress(t *testing.T) {
	b := BlocklistDialer{List: map[string]struct{}{
		"hello.world": struct{}{},
	}, Dialer: &TestNopDialer{}}
	_, err := b.Dial("tcp", "hello.world:80")
	if err == ErrBlockedHost {
		return
	}
	t.FailNow()
}

func TestBlocklistDialerLoadHosts(t *testing.T) {
	hostsFile := []byte("127.0.0.1 localhost\n0.0.0.0 hello.world\n# the next line tests 2 host names for one destination address\n0.0.0.0 hello.world.too hello.world.future\n")
	b := BlocklistDialer{List: make(map[string]struct{}, 0), Dialer: &TestNopDialer{}}
	b.Load(bytes.NewReader(hostsFile))
	_, err := b.Dial("tcp", "hello.world:80")
	if err != ErrBlockedHost {
		t.FailNow()
	}
	_, err = b.Dial("tcp", "hello.world.too:443")
	if err != ErrBlockedHost {
		t.FailNow()
	}
	_, err = b.Dial("tcp", "hello.world.future:443")
	if err != ErrBlockedHost {
		t.FailNow()
	}
	_, err = b.Dial("tcp", "hello.world.past:80")
	if err != nil {
		t.FailNow()
	}
}

func TestLoadBlocklistFromFile(t *testing.T) {
	dialer := BlocklistDialer{List: make(map[string]struct{}, 0), Dialer: &TestNopDialer{}}
	LoadHostsFile(&dialer, "test/hosts")
	if _, err := dialer.Dial("tcp", "hello.world:443"); err != ErrBlockedHost {
		t.Fail()
	}
	if _, err := dialer.Dial("tcp", "hello.past:80"); err != ErrBlockedHost {
		t.Fail()
	}
	if _, err := dialer.Dial("tcp", "hello.future:443"); err != nil {
		t.Fail()
	}
}

type TestNopDialer struct{}

func (*TestNopDialer) Dial(network, addr string) (net.Conn, error) {
	return nil, nil
}
