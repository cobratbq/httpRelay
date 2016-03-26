package httpRelay

import (
	"bufio"
	"io"
	"net/http"

	"golang.org/x/net/proxy"
)

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
	defer func() {
		logError(req.Body.Close(), "Error closing client request body:")
	}()
	var err error
	switch req.Method {
	case "CONNECT":
		err = h.handleConnect(resp, req)
	default:
		err = h.processRequest(resp, req)
	}
	logError(err, "Error serving proxy relay")
}

func (h *HTTPProxyHandler) processRequest(resp http.ResponseWriter, req *http.Request) error {
	var err error
	logRequest(req)
	// Verification of requests is already handled by net/http library.
	// Establish connection with socks proxy
	conn, err := h.Dialer.Dial("tcp", fullHost(req.Host))
	if err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		return err
	}
	defer func() {
		logError(conn.Close(), "Error closing connection to socks proxy:")
	}()
	// Prepare request for socks proxy
	proxyReq, err := http.NewRequest(req.Method, req.RequestURI, req.Body)
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
	// Verification of response is already handled by net/http library.
	resp.WriteHeader(proxyResp.StatusCode)
	_, err = io.Copy(resp, proxyResp.Body)
	if err != nil {
		return err
	}
	logError(proxyResp.Body.Close(), "Error closing response body:")
	return nil
}

func (h *HTTPProxyHandler) handleConnect(resp http.ResponseWriter, req *http.Request) error {
	logRequest(req)
	// Establish connection with socks proxy
	proxyConn, err := h.Dialer.Dial("tcp", req.Host)
	if err != nil {
		return err
	}
	// Acquire raw connection to the client
	clientConn, err := acquireConn(resp)
	if err != nil {
		logError(proxyConn.Close(), "Error while closing proxy connection:")
		resp.WriteHeader(http.StatusInternalServerError)
		return err
	}
	// Send 200 Connection established to client to signal tunnel ready
	// Responses to CONNECT requests MUST NOT contain any body payload.
	// TODO add additional headers to proxy server's response? (Via)
	// TODO decide on response message type based on req protocol (http2)
	_, err = clientConn.Write([]byte(req.Proto + " 200 Connection established\r\n\r\n"))
	if err != nil {
		logError(proxyConn.Close(), "Error while closing proxy connection:")
		resp.WriteHeader(http.StatusInternalServerError)
		return err
	}
	// Start copying data from one connection to the other
	go transfer(proxyConn, clientConn)
	go transfer(clientConn, proxyConn)
	return nil
}
