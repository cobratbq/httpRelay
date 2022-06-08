package httprelay

import (
	"bytes"
	"net"
	"testing"

	assert "github.com/cobratbq/goutils/std/testing"
)

func TestBlocklistDialerNilDialer(t *testing.T) {
	defer assert.RequirePanic(t)
	b := BlocklistDialer{List: make(map[string]struct{}, 0), Dialer: nil}
	b.Dial("tcp", "hello.world:80")
	t.FailNow()
}

func TestBlocklistDialerDirectDialer(t *testing.T) {
	b := BlocklistDialer{List: make(map[string]struct{}, 0), Dialer: &TestNopDialer{}}
	if _, err := b.Dial("tcp", "hello.world:80"); err != nil {
		t.FailNow()
	}
}

func TestBlocklistDialerBlockedAddress(t *testing.T) {
	b := BlocklistDialer{List: map[string]struct{}{
		"hello.world": struct{}{},
	}, Dialer: &TestNopDialer{}}
	if _, err := b.Dial("tcp", "hello.world:80"); err == ErrBlockedHost {
		return
	}
	t.FailNow()
}

func TestBlocklistDialerLoadHosts(t *testing.T) {
	hostsFile := []byte("127.0.0.1 localhost\n0.0.0.0 hello.world\n# the next line tests 2 host names for one destination address\n0.0.0.0 hello.world.too hello.world.future\n")
	b := BlocklistDialer{List: make(map[string]struct{}, 0), Dialer: &TestNopDialer{}}
	b.Load(bytes.NewReader(hostsFile))
	if _, err := b.Dial("tcp", "hello.world:80"); err != ErrBlockedHost {
		t.FailNow()
	}
	if _, err := b.Dial("tcp", "hello.world.too:443"); err != ErrBlockedHost {
		t.FailNow()
	}
	if _, err := b.Dial("tcp", "hello.world.future:443"); err != ErrBlockedHost {
		t.FailNow()
	}
	if _, err := b.Dial("tcp", "hello.world.past:80"); err != nil {
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
