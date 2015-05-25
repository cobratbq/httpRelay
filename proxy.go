package main

import (
	"bufio"
	"flag"
	"io"
	"log"
	"net/http"

	"golang.org/x/net/proxy"
)

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

func main() {
	socksAddr := flag.String("socks", "localhost:8000", "Address and port of SOCKS5 proxy server.")
	listenAddr := flag.String("listen", ":8080", "Listening address and port for HTTP relay proxy.")
	flag.Parse()
	// Prepare proxy relay with target SOCKS proxy
	dialer, err := proxy.SOCKS5("tcp", *socksAddr, nil, proxy.Direct)
	if err != nil {
		log.Println("error creating proxy definition:", err.Error())
		return
	}
	// Start HTTP proxy server
	log.Println("HTTP proxy relay server started on", *listenAddr, "relaying to", *socksAddr)
	log.Println(http.ListenAndServe(*listenAddr, &HTTPProxyHandler{Dialer: dialer}))
}

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
	if err != nil {
		log.Println("error:", err.Error())
	}
}

func (h *HTTPProxyHandler) processRequest(resp http.ResponseWriter, req *http.Request) error {
	log.Println(req.Method, req.Host, req.Proto)
	// Establish connection with socks proxy
	conn, err := h.Dialer.Dial("tcp", fullHost(req.Host))
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
	// Transfer headers to proxy request
	reqHdrs := make(map[string]struct{}, len(hopByHopHeaders))
	duplicateDropHeaderSet(reqHdrs, hopByHopHeaders)
	CopyHeaders(reqHdrs, proxyReq.Header, req.Header)
	proxyReq.Header.Add("User-Agent", req.UserAgent())
	// Send request to socks proxy
	if err = proxyReq.Write(conn); err != nil {
		resp.WriteHeader(500)
		return err
	}
	// Read proxy response
	proxyRespReader := bufio.NewReader(conn)
	proxyResp, err := http.ReadResponse(proxyRespReader, proxyReq)
	if err != nil {
		resp.WriteHeader(500)
		return err
	}
	// Transfer headers to client response
	respHdrs := make(map[string]struct{}, len(hopByHopHeaders))
	duplicateDropHeaderSet(respHdrs, hopByHopHeaders)
	CopyHeaders(respHdrs, resp.Header(), proxyResp.Header)
	if err = VerifyHeaders(resp.Header()); err != nil {
		resp.WriteHeader(502)
		return err
	}
	resp.WriteHeader(proxyResp.StatusCode)
	// Copy response body to client
	_, err = io.Copy(resp, proxyResp.Body)
	if err != nil {
		return err
	}
	logError(proxyResp.Body.Close(), "error closing response body:")
	return nil
}

func (h *HTTPProxyHandler) handleConnect(resp http.ResponseWriter, req *http.Request) error {
	log.Println("CONNECT", req.Host)
	// Establish connection with socks proxy
	proxyConn, err := h.Dialer.Dial("tcp", req.Host)
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
	go transfer(proxyConn, clientConn)
	go transfer(clientConn, proxyConn)
	return nil
}
