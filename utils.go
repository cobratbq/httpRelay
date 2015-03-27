package main

import (
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
)

func fullHost(host string) string {
	fullhost := host
	if strings.IndexByte(host, ':') == -1 {
		fullhost += ":80"
	}
	return fullhost
}

func copyHeaders(dst http.ResponseWriter, src *http.Response) {
	for k, vals := range src.Header {
		if _, drop := hopByHopHeaders[k]; drop {
			continue
		}
		for _, v := range vals {
			dst.Header().Add(k, v)
		}
	}
}

func acquireConn(resp http.ResponseWriter) (net.Conn, error) {
	hijacker, ok := resp.(http.Hijacker)
	if !ok {
		return nil, errors.New("failed to acquire raw client connection")
	}
	clientConn, _, err := hijacker.Hijack()
	return clientConn, err
}

func tunnel(dst, src net.Conn) {
	_, _ = io.Copy(dst, src)
	logError(dst.Close(), "error while closing tunnel destination connection:")
}

func logError(err error, prefix string) {
	if err == nil {
		return
	}
	log.Println(prefix, err.Error())
}
