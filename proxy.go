package main

import (
	"bufio"
	"io"
	"log"
	"net/http"

	"golang.org/x/net/proxy"
)

func main() {
	// TODO add cli flag support
	// TODO change default socks proxy port to 8080
	dialer, err := proxy.SOCKS5("tcp", "localhost:8000", nil, proxy.Direct)
	if err != nil {
		log.Println("error creating proxy definition:", err.Error())
		return
	}
	log.Println("HTTP proxy relay server started.")
	log.Println(http.ListenAndServe(":8080", &HTTPProxyHandler{Dialer: dialer}))
}

var hopByHopHeaders map[string]struct{}

func init() {
	hopByHopHeaders = make(map[string]struct{})
	hopByHopHeaders["Connection"] = struct{}{}
	hopByHopHeaders["Keep-Alive"] = struct{}{}
	hopByHopHeaders["Proxy-Authorization"] = struct{}{}
	hopByHopHeaders["Proxy-Authentication"] = struct{}{}
	hopByHopHeaders["TE"] = struct{}{}
	hopByHopHeaders["Trailer"] = struct{}{}
	hopByHopHeaders["Transfer-Encoding"] = struct{}{}
	hopByHopHeaders["Upgrade"] = struct{}{}
}

// HTTPProxyHandler is a proxy handler that passes on request to a SOCKS5 proxy server.
type HTTPProxyHandler struct {
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
	if err != nil {
		log.Println("error:", err.Error())
	}
}

func (h *HTTPProxyHandler) processRequest(resp http.ResponseWriter, req *http.Request) error {
	log.Println(req.Method, req.Host, req.Proto)
	// Establish connection with socks proxy
	conn, err := h.Dialer.Dial("tcp4", fullHost(req.Host))
	if err != nil {
		resp.WriteHeader(500)
		return err
	}
	defer func() {
		logError(conn.Close(), "error closing connection to socks proxy:")
	}()
	// Prepare request for socks proxy
	proxyReq, err := http.NewRequest(req.Method, req.RequestURI, req.Body)
	if err != nil {
		resp.WriteHeader(500)
		return err
	}
	proxyReq.Header.Add("User-Agent", req.UserAgent())
	// Send request to socks proxy
	if err = proxyReq.Write(conn); err != nil {
		log.Println("failed to write request meant for proxy to the proxy connection", err.Error())
	}
	// Read proxy response
	proxyRespReader := bufio.NewReader(conn)
	proxyResp, err := http.ReadResponse(proxyRespReader, proxyReq)
	if err != nil {
		resp.WriteHeader(500)
		return err
	}
	// Prepare response to client
	copyHeaders(resp, proxyResp)
	_, err = io.Copy(resp, proxyResp.Body)
	if err != nil {
		resp.WriteHeader(500)
		return err
	}
	logError(proxyResp.Body.Close(), "error closing response body:")
	return nil
}

func (h *HTTPProxyHandler) handleConnect(resp http.ResponseWriter, req *http.Request) error {
	log.Println("CONNECT", req.Host)
	// Establish connection with socks proxy
	proxyConn, err := h.Dialer.Dial("tcp4", req.Host)
	if err != nil {
		return err
	}
	// Acquire raw connection to the client
	clientConn, err := acquireConn(resp)
	if err != nil {
		logError(proxyConn.Close(), "error while closing proxy connection:")
		resp.WriteHeader(500)
		return err
	}
	// Send 200 Connection established to client to signal tunnel ready
	_, err = clientConn.Write([]byte("HTTP/1.1 200 Connection established\r\n\r\n\r\n"))
	if err != nil {
		logError(proxyConn.Close(), "error while closing proxy connection:")
		resp.WriteHeader(500)
		return err
	}
	// Start copying data from one connection to the other
	go tunnel(proxyConn, clientConn)
	go tunnel(clientConn, proxyConn)
	return nil
}
