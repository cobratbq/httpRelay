package httprelay

import (
	"bufio"
	"bytes"
	"io"
	"net"
	"net/http"
	"sync"
	"time"

	io_ "github.com/cobratbq/goutils/std/io"
	"github.com/cobratbq/goutils/std/log"
	http_ "github.com/cobratbq/goutils/std/net/http"
	"golang.org/x/net/proxy"
)

func DirectDialer() net.Dialer {
	return net.Dialer{
		Timeout:       0,
		Deadline:      time.Time{},
		KeepAlive:     -1,
		FallbackDelay: -1,
	}
}

// mnot's blog: https://www.mnot.net/blog/2011/07/11/what_proxies_must_do
// rfc: http://tools.ietf.org/html/draft-ietf-httpbis-p1-messaging-14#section-3.3
//
// 0. Advertise HTTP/1.1 Correctly - this is covered by starting a normal HTTP
// client connection through the SOCKS proxy which establishes a (sane)
// connection according to its own parameters, instead of blindly copying
// parameters from the original requesting client's connection.
//
// 1. Remove Hop-by-hop Headers - this is covered in the copyHeaders function
// which explicitly skips known hop-by-hop headers and checks 'Connection'
// header for additional headers we need to skip.
//
// Remarks from the blog post not covered explicitly are tested in the tests
// assumptions_test.go

// HTTPProxyHandler is a proxy handler that passes on request to a SOCKS5 proxy server.
type HTTPProxyHandler struct {
	// Dialer is the dialer for connecting to the SOCKS5 proxy.
	Dialer    proxy.Dialer
	UserAgent string
}

func (h *HTTPProxyHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	var err error
	switch req.Method {
	case "CONNECT":
		// TODO Go 1.20 added an OnProxyConnect callback for use by proxies. This probably voids the use for connection hijacking. Investigate and possibly use.
		err = h.handleConnect(resp, req)
	default:
		err = h.processRequest(resp, req)
	}
	if err != nil {
		log.Warnln("Error serving request:", err.Error())
	}
}

// TODO append body that explains the error as is expected from 5xx http status codes
func (h *HTTPProxyHandler) processRequest(resp http.ResponseWriter, req *http.Request) error {
	// TODO what to do when body of request is very large?
	body, err := io.ReadAll(req.Body)
	if err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		return err
	}
	io_.CloseLogged(req.Body, "Failed to close request body: %+v")
	// The request body is only closed in certain error cases. In other cases, we
	// let body be closed by during processing of request to remote host.
	log.Infoln(req.Proto, req.Method, req.URL.Host)
	// Verification of requests is already handled by net/http library.
	// Establish connection with socks proxy
	conn, err := h.Dialer.Dial("tcp", fullHost(req.URL.Host))
	if err == ErrBlockedHost {
		resp.WriteHeader(http.StatusForbidden)
		return err
	} else if err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		return err
	}
	defer io_.CloseLoggedWithIgnores(conn, "Error closing connection to socks proxy: %+v", io.ErrClosedPipe)
	// Prepare request for socks proxy
	proxyReq, err := http.NewRequest(req.Method, req.RequestURI, bytes.NewReader(body))
	if err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		return err
	}
	// Transfer headers to proxy request
	copyHeaders(proxyReq.Header, req.Header)
	if h.UserAgent != "" {
		// Add specified user agent as header.
		proxyReq.Header.Add("User-Agent", h.UserAgent)
	}
	// Send request to socks proxy
	if err = proxyReq.Write(conn); err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		return err
	}
	// Read proxy response
	proxyRespReader := bufio.NewReader(conn)
	proxyResp, err := http.ReadResponse(proxyRespReader, proxyReq)
	if err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		return err
	}
	// Transfer headers to client response
	copyHeaders(resp.Header(), proxyResp.Header)
	resp.Header().Set("Connection", "close")
	// Verification of response is already handled by net/http library.
	resp.WriteHeader(proxyResp.StatusCode)
	_, err = io.Copy(resp, proxyResp.Body)
	io_.CloseLoggedWithIgnores(proxyResp.Body, "Error closing response body: %+v", io.ErrClosedPipe)
	return err
}

// TODO append body that explains the error as is expected from 5xx http status codes
func (h *HTTPProxyHandler) handleConnect(resp http.ResponseWriter, req *http.Request) error {
	defer io_.CloseLoggedWithIgnores(req.Body, "Error while closing request body: %+v", io.ErrClosedPipe)
	log.Infoln(req.Proto, req.Method, req.URL.Host)
	// Establish connection with socks proxy
	proxyConn, err := h.Dialer.Dial("tcp", req.Host)
	if err == ErrBlockedHost {
		resp.WriteHeader(http.StatusForbidden)
		return err
	} else if err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		return err
	}
	defer io_.CloseLoggedWithIgnores(proxyConn, "Failed to close connection to remote location: %+v", io.ErrClosedPipe)
	// Acquire raw connection to the client
	clientInput, clientConn, err := http_.HijackConnection(resp)
	if err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		return err
	}
	defer io_.CloseLoggedWithIgnores(clientConn, "Failed to close connection to local client: %+v", io.ErrClosedPipe)
	// Send 200 Connection established to client to signal tunnel ready
	// Responses to CONNECT requests MUST NOT contain any body payload.
	// TODO add additional headers to proxy server's response? (Via)
	if _, err = clientConn.Write([]byte("HTTP/1.1 200 Connection established\r\n\r\n")); err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		return err
	}
	// Start copying data from one connection to the other
	var wg sync.WaitGroup
	wg.Add(2)
	go io_.Transfer(&wg, proxyConn, clientInput)
	go io_.Transfer(&wg, clientConn, proxyConn)
	wg.Wait()
	return nil
}
