package httpRelay

import (
	"bufio"
	"io"
	"net/http"

	"golang.org/x/net/proxy"
)

// mnot's blog: https://www.mnot.net/blog/2011/07/11/what_proxies_must_do
// rfc: http://tools.ietf.org/html/draft-ietf-httpbis-p1-messaging-14#section-3.3

// HTTPProxyHandler is a proxy handler that passes on request to a SOCKS5 proxy server.
type HTTPProxyHandler struct {
	// Dialer is the dialer for connecting to the SOCKS5 proxy.
	Dialer proxy.Dialer
}

func (h *HTTPProxyHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	defer func() {
		logError(req.Body.Close(), "error closing client request body:")
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
		logError(conn.Close(), "error closing connection to socks proxy:")
	}()
	// Prepare request for socks proxy
	proxyReq, err := http.NewRequest(req.Method, req.RequestURI, req.Body)
	if err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		return err
	}
	// Transfer headers to proxy request
	copyHeaders(proxyReq.Header, req.Header)
	// FIXME add Via header
	// FIXME add what user agent? (Does setting header actually work?)
	proxyReq.Header.Add("User-Agent", "proxy")
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
	logError(proxyResp.Body.Close(), "error closing response body:")
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
		logError(proxyConn.Close(), "error while closing proxy connection:")
		resp.WriteHeader(http.StatusInternalServerError)
		return err
	}
	// Send 200 Connection established to client to signal tunnel ready
	_, err = clientConn.Write([]byte("HTTP/1.1 200 Connection established\r\n\r\n\r\n"))
	if err != nil {
		logError(proxyConn.Close(), "error while closing proxy connection:")
		resp.WriteHeader(http.StatusInternalServerError)
		return err
	}
	// Start copying data from one connection to the other
	go transfer(proxyConn, clientConn)
	go transfer(clientConn, proxyConn)
	return nil
}
