package httprelay

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/cobratbq/goutils/std/builtin/set"
	"github.com/cobratbq/goutils/std/log"
)

// headers that are dedicated to a single connection and should not copied to
// the SOCKS proxy server connection
var hopByHopHeaders = set.Create(connectionHeader, "Keep-Alive", "Proxy-Authorization",
	"Proxy-Authentication", "TE", "Trailer", "Transfer-Encoding", "Upgrade")

// connectionHeader is the 'Connection' header which is an indicator for other
// headers that should be dropped as hop-by-hop headers.
const connectionHeader = "Connection"

// fullHost appends the default port to the provided host if no port is specified.
// FIXME check if this causes problems in case of IPv6
func fullHost(host string) string {
	fullhost := host
	if strings.IndexByte(host, ':') == -1 {
		fullhost += ":80"
	}
	return fullhost
}

// copyHeaders copies all the headers that are not classified as hop-to-hop headers. (This satisfies
// 1. Remove Hop-by-hop Headers.)
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

// processConnectionHdr processes the Connection header and adds all headers listed in value as
// droppable headers.
func processConnectionHdr(dropHdrs map[string]struct{}, value string) []string {
	var bad []string
	parts := strings.Split(value, ",")
	for _, part := range parts {
		header := strings.TrimSpace(part)
		if tokenPattern.MatchString(header) {
			set.Insert(dropHdrs, header)
		} else {
			bad = append(bad, header)
		}
	}
	return bad
}

// tokenPattern is the pattern of a valid token.
var tokenPattern = regexp.MustCompile(`^[\d\w\!#\$%&'\*\+\-\.\^_\|~` + "`" + `]+$`)

// log the request
func logRequest(req *http.Request) {
	log.Infoln(req.Proto, req.Method, req.Host)
}

func logWarning(parts ...any) {
	log.Warnln(parts...)
}
