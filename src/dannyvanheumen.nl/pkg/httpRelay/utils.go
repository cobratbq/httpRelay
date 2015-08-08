package httpRelay

import (
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"regexp"
	"strings"
)

var tokenPattern = regexp.MustCompile(`^[\d\w\!#\$%&'\*\+\-\.\^_\|~` + "`" + `]+$`)

// fullHost appends the default port to the provided host if no port is
// specified.
func fullHost(host string) string {
	fullhost := host
	if strings.IndexByte(host, ':') == -1 {
		fullhost += ":80"
	}
	return fullhost
}

func copyHeaders(dropHdrs map[string]struct{}, dst http.Header, src http.Header) {
	for k, vals := range src {
		if _, drop := dropHdrs[k]; drop {
			continue
		}
		for _, v := range vals {
			if k == "Connection" {
				processConnectionHdr(dropHdrs, v)
				continue
			}
			dst.Add(k, v)
		}
	}
}

func duplicateDropHeaderSet(dst map[string]struct{}, src map[string]struct{}) {
	for k, v := range src {
		dst[k] = v
	}
}

func processConnectionHdr(connHdrs map[string]struct{}, value string) {
	parts := strings.Split(value, ",")
	for _, part := range parts {
		header := strings.TrimSpace(part)
		if tokenPattern.MatchString(header) {
			connHdrs[header] = struct{}{}
		} else {
			log.Println("Skipping bad value in Connection header.")
		}
	}
}

// acquireConn acquires the underlying connection by inspecting the
// ResponseWriter provided.
func acquireConn(resp http.ResponseWriter) (net.Conn, error) {
	hijacker, ok := resp.(http.Hijacker)
	if !ok {
		return nil, errors.New("failed to acquire raw client connection")
	}
	clientConn, _, err := hijacker.Hijack()
	return clientConn, err
}

// transfer may be launched as goroutine. It that copies all content from one
// connection to the next.
func transfer(dst io.WriteCloser, src io.Reader) {
	_, err := io.Copy(dst, src)
	if err != nil {
		log.Println("error occurred while transferring data between connections", err.Error())
	}
	logError(dst.Close(), "error while closing tunnel destination connection:")
}

// logError logs an error if an error was returned.
func logError(err error, prefix string) {
	if err != nil {
		log.Println(prefix, err.Error())
	}
}
