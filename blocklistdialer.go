package httpRelay

import (
	"bufio"
	"io"
	"log"
	"net"
	"strings"

	"golang.org/x/net/proxy"
)

// BlocklistDialer checks the loaded blocklist before dialing.
// FIXME write unit tests.
type BlocklistDialer struct {
	List   map[string]struct{}
	Dialer proxy.Dialer
}

// Dial checks the address against the blocklist and if not present uses the
// provided dialer to dial the address.
func (b *BlocklistDialer) Dial(network, addr string) (net.Conn, error) {
	if _, ok := b.List[addr]; ok {
		// FIXME what HTTP status code to use for blocked domains, 404? Need appropriate error code.
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
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") {
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
