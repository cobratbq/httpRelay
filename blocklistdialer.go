package httprelay

import (
	"bufio"
	"io"
	"log"
	"net"
	"os"
	"strings"

	"github.com/cobratbq/goutils/assert"
	io_ "github.com/cobratbq/goutils/std/io"
	"golang.org/x/net/proxy"
)

// BlocklistDialer checks the loaded blocklist before dialing.
type BlocklistDialer struct {
	List   map[string]struct{}
	Dialer proxy.Dialer
}

// Dial checks the address against the blocklist and if not present uses the
// provided dialer to dial the address.
func (b *BlocklistDialer) Dial(network, addr string) (net.Conn, error) {
	if i := strings.IndexByte(addr, ':'); i > -1 {
		addr = addr[:i]
	}
	if _, ok := b.List[addr]; ok {
		return nil, ErrBlockedHost
	}
	return b.Dialer.Dial(network, addr)
}

// Load loads a blocklist from provided reader that has content formatted like
// the operating system 'hosts' files.
func (b *BlocklistDialer) Load(in io.Reader) error {
	reader := bufio.NewReader(in)
	skipped := uint(0)
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		assert.RequireSuccess(err, "Failed to read hosts content: %+v")
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			// skip comment lines
			continue
		}
		parts := strings.Fields(line)
		if parts[0] != "0.0.0.0" {
			skipped++
			// for now, only accept resolutions to 0.0.0.0 for purpose of
			// blocking
			continue
		}
		for _, entry := range parts[1:] {
			b.List[entry] = struct{}{}
		}
	}
	if skipped > 0 {
		log.Printf("Skipped %d lines for not using destination address '0.0.0.0'.", skipped)
	}
	log.Printf("Loaded %d entries into blocklist.", len(b.List))
	return nil
}

// LoadHostsFile loads a `hosts`-formatted blocklist into provided
// BlocklistDialer.
func LoadHostsFile(dialer *BlocklistDialer, filename string) error {
	hostsFile, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer io_.CloseLogged(hostsFile, "failed to close hosts file")
	if err := dialer.Load(hostsFile); err != nil {
		return err
	}
	return nil
}
