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

func CopyHeaders(dropHdrs map[string]struct{}, dst http.Header, src http.Header) {
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

var ErrMultipleContentLengthVariations = errors.New("Encountered multiple variations on Content-Length header.")
var ErrMultipleContentLengthHeaders = errors.New("Encountered response with multiple Content-Length headers.")

func VerifyHeaders(headers http.Header) error {
	var contentLengthCount = 0
	var contentLength = false
	var contentLengthHeader string
	var transferEncodingChunked = false
	// verify all headers
	for k, v := range headers {
		if strings.ToLower(k) == "content-length" {
			contentLengthCount++
			contentLength = true
			contentLengthHeader = k
			if len(v) > 1 {
				log.Printf("Encountered response with multiple Content-Length headers: %+v", v)
				return ErrMultipleContentLengthHeaders
			}
		}
		if strings.ToLower(k) == "transfer-encoding" {
			for _, val := range v {
				if strings.ToLower(val) == "chunked" {
					transferEncodingChunked = true
				}
			}
		}
	}
	// additional validation
	if contentLengthCount > 1 {
		return ErrMultipleContentLengthVariations
	}
	if contentLength && transferEncodingChunked {
		delete(headers, contentLengthHeader)
		log.Println("Deleted Content-Length header since response also contains Transfer-Encoding: chunked header.")
	}
	return nil
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
		connHdrs[header] = struct{}{}
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

func transfer(dst, src net.Conn) {
	_, err := io.Copy(dst, src)
	if err != nil {
		log.Println("error occurred while transferring data between connections", err.Error())
	}
	logError(dst.Close(), "error while closing tunnel destination connection:")
}

func logError(err error, prefix string) {
	if err == nil {
		return
	}
	log.Println(prefix, err.Error())
}
