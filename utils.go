package httprelay

import (
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"sync"
)

// tokenPatternRegex is the raw string pattern that should be compiled.
const tokenPatternRegex = `^[\d\w\!#\$%&'\*\+\-\.\^_\|~` + "`" + `]+$`

// tokenPattern is the pattern of a valid token.
var tokenPattern = regexp.MustCompile(tokenPatternRegex)

// connectionHeader is the 'Connection' header which is an indicator for other
// headers that should be dropped as hop-by-hop headers.
const connectionHeader = "Connection"

// headers that are dedicated to a single connection and should not copied to
// the SOCKS proxy server connection
var hopByHopHeaders = map[string]struct{}{
	connectionHeader:       struct{}{},
	"Keep-Alive":           struct{}{},
	"Proxy-Authorization":  struct{}{},
	"Proxy-Authentication": struct{}{},
	"TE":                   struct{}{},
	"Trailer":              struct{}{},
	"Transfer-Encoding":    struct{}{},
	"Upgrade":              struct{}{},
}

// fullHost appends the default port to the provided host if no port is
// specified.
func fullHost(host string) string {
	fullhost := host
	if strings.IndexByte(host, ':') == -1 {
		fullhost += ":80"
	}
	return fullhost
}

// copyHeaders copies all the headers that are not classified as hop-to-hop
// headers. (This satisfies 1. Remove Hop-by-hop Headers.)
func copyHeaders(dst http.Header, src http.Header) {
	var dynDropHdrs = map[string]struct{}{}
	if vals, ok := src[connectionHeader]; ok {
		for _, v := range vals {
			processConnectionHdr(dynDropHdrs, v)
		}
	}
	for k, vals := range src {
		// This assumes that Connection header is also an element of
		// hop-by-hop headers such that it will not be processed twice,
		// but instead is dropped with the others.
		if _, drop := hopByHopHeaders[k]; drop {
			continue
		} else if _, drop := dynDropHdrs[k]; drop {
			continue
		}
		for _, v := range vals {
			dst.Add(k, v)
		}
	}
}

// processConnectionHdr processes the Connection header and adds all headers
// listed in value as droppable headers.
func processConnectionHdr(dropHdrs map[string]struct{}, value string) []string {
	var bad []string
	parts := strings.Split(value, ",")
	for _, part := range parts {
		header := strings.TrimSpace(part)
		if tokenPattern.MatchString(header) {
			dropHdrs[header] = struct{}{}
		} else {
			bad = append(bad, header)
		}
	}
	return bad
}

// transfer may be launched as goroutine. It that copies all content from one
// connection to the next.
func transfer(wg *sync.WaitGroup, dst io.Writer, src io.Reader) {
	_, _ = io.Copy(dst, src)
	// Skip all error handling, because we simply cannot distinguish between
	// expected and unexpected events. Logging this will only produce noise.
	wg.Done()
}

// logError logs an error if an error was returned.
func logError(err error, prefix string) {
	if err != nil {
		log.Println(prefix, err.Error())
	}
}

// log the request
func logRequest(req *http.Request) {
	log.Println(req.Proto, req.Method, req.Host)
}
