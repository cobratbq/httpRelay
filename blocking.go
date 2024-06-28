package httprelay

import (
	"bufio"
	"io"
	"log"
	"net"
	"os"
	"strings"

	bufio_ "github.com/cobratbq/goutils/std/bufio"
	"github.com/cobratbq/goutils/std/builtin/set"
	"github.com/cobratbq/goutils/std/builtin/slices"
	"github.com/cobratbq/goutils/std/errors"
	io_ "github.com/cobratbq/goutils/std/io"
	net_ "github.com/cobratbq/goutils/std/net"
	"golang.org/x/net/proxy"
)

// WrapPerHostBlocking wraps a dialer with a PerHost conditional bypass dialer that refuses dialing
// any address that is local or custom specified according to parameters specified.
func WrapPerHostBlocking(dialer proxy.Dialer, local bool, custom string) proxy.Dialer {
	// Prepare dialer to block addresses
	perHostDialer := proxy.NewPerHost(dialer, &NopDialer{})
	if local {
		slices.ForEach(net_.PrivateNetworks, perHostDialer.AddNetwork)
	}
	if custom != "" {
		perHostDialer.AddFromString(custom)
	}
	return perHostDialer
}

// WrapBlocklistBlocking loads a blocklist from specified file and includes it in the dialer. Any
// address present on the blocklist will not be allowed to dial.
func WrapBlocklistBlocking(dialer proxy.Dialer, fileName string) (proxy.Dialer, error) {
	blocklistDialer := BlocklistDialer{
		List:   make(map[string]struct{}, 0),
		Dialer: dialer}
	if err := loadHostsFile(&blocklistDialer, fileName); err != nil {
		return nil, errors.Context(err, "failed to load blocklist: "+fileName)
	}
	return &blocklistDialer, nil
}

// loadHostsFile loads a `hosts`-formatted blocklist into provided BlocklistDialer.
func loadHostsFile(dialer *BlocklistDialer, filename string) error {
	hostsFile, err := os.Open(filename)
	if err != nil {
		return errors.Context(err, "failed to open file "+filename)
	}
	defer io_.CloseLogged(hostsFile, "failed to close hosts file")
	return dialer.Load(hostsFile)
}

// BlocklistDialer checks the loaded blocklist before dialing.
type BlocklistDialer struct {
	List   map[string]struct{}
	Dialer proxy.Dialer
}

// Dial checks the address against the blocklist and if not present uses the provided dialer to dial
// the address.
func (b *BlocklistDialer) Dial(network, addr string) (net.Conn, error) {
	if i := strings.IndexByte(addr, ':'); i > -1 {
		addr = addr[:i]
	}
	if _, ok := b.List[addr]; ok {
		return nil, ErrBlockedHost
	}
	return b.Dialer.Dial(network, addr)
}

// Load loads a blocklist from provided reader that has content formatted like the operating system
// 'hosts' files.
func (b *BlocklistDialer) Load(in io.Reader) error {
	reader := bufio.NewReader(in)
	var skipped uint
	if err := bufio_.ReadStringLinesFunc(reader, '\n', func(line string) error {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			// skip comment lines
			return nil
		}
		parts := strings.Fields(line)
		if parts[0] != "0.0.0.0" {
			// for now, only allow resolutions to 0.0.0.0 for purpose of blocking
			skipped++
			return nil
		}
		set.InsertMany(b.List, parts[1:])
		return nil
	}); err != nil {
		return errors.Context(err, "failed to read hosts content")
	}
	if skipped > 0 {
		log.Printf("Skipped %d lines for not using destination address '0.0.0.0'.", skipped)
	}
	log.Println("Total entries in blocklist:", len(b.List))
	return nil
}
