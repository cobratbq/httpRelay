package httprelay

import (
	"bufio"
	"bytes"
	"net/http"
	"testing"

	assert "github.com/cobratbq/goutils/std/testing"
)

// These tests are are based on instructions in the blog post
// https://www.mnot.net/blog/2011/07/11/what_proxies_must_do.
// This blog posts remarks on some (very) important checks that proxy servers
// must do to ensure correct behaviour under (sometimes questionable)
// circumstances. As most of these points are already covered by the net/http
// package in the standard library, these tests are there for testing that our
// assumptions still hold.

// 0. Advertise HTTP/1.1 Correctly is handled automatically since the relay
// initiates its own (new) HTTP client connection to the target server. This
// means that the connection basics, such as the protocol, are determined by
// the connection itself and not through copying the original request of the
// user.

// 1. Remove Hop-by-Hop Headers is explicitly handled in the code, so no need
// to test it as an assumption.

// TestAssumptionBadFraminingMultipleContentLength verifies that the standard
// package handles bad framing issues by testing the parsing behaviour of the
// net/http package. This is not a 100% perfect check, as we depend on http
// server for hosting the relay server.
// The assumption we make here is that the server and the ReadResponse function
// use the same underlying implementation.
// (2. Detect Bad Framing from the blog post.)
func TestAssumptionBadFramingMultipleContentLength(t *testing.T) {
	// Acquire request instance by creating a basic request
	requestRdr := bytes.NewBufferString(`GET / HTTP/1.1
Host: www.example.com

`)
	requestBufRdr := bufio.NewReader(requestRdr)
	req, err := http.ReadRequest(requestBufRdr)
	assert.Nil(t, err)
	// Verify behaviour reading a response by providing corrupted response content
	responseRdr := bytes.NewBufferString(`HTTP/1.1 200 OK
Content-Type: text/html; charset=utf-8
Content-Length: 10
Content-Length: 20

abcdefghij`)
	responseBufRdr := bufio.NewReader(responseRdr)
	_, err = http.ReadResponse(responseBufRdr, req)
	assert.NotNil(t, err)
}

// TestAssumptionBadFraminingContentLengthWithChunked verifies that the standard
// package handles bad framing issues by testing the parsing behaviour of the
// net/http package. This is not a 100% perfect check, as we depend on http
// server for hosting the relay server.
// The assumption we make here is that the server and the ReadResponse function
// use the same underlying implementation.
// (2. Detect Bad Framing from the blog post.)
func TestAssumptionBadFramingContentLengthWithChunked(t *testing.T) {
	// Acquire request instance by creating a basic request
	requestRdr := bytes.NewBufferString(`GET / HTTP/1.1
Host: www.example.com

`)
	requestBufRdr := bufio.NewReader(requestRdr)
	req, err := http.ReadRequest(requestBufRdr)
	assert.Nil(t, err)
	// Verify behaviour reading a response by providing corrupted response content
	responseRdr := bytes.NewBufferString(`HTTP/1.1 200 OK
Content-Type: text/html; charset=utf-8
Content-Length: 10
Transfer-Encoding: chunked

abcdefghij`)
	responseBufRdr := bufio.NewReader(responseRdr)
	resp, err := http.ReadResponse(responseBufRdr, req)
	assert.Nil(t, err)
	assert.Equal(t, resp.ContentLength, -1)
	assert.Equal(t, resp.TransferEncoding[0], "chunked")
}

// TestCorrectlyReadRoutingFromRequest tests whether conflicting information
// w.r.t. host name gets read correctly. Absolute request URI overrides
// information in Host header.
// (3. Route Well from blog post)
func TestCorrectlyReadRoutingFromRequest(t *testing.T) {
	// Acquire request instance by creating a basic request
	requestRdr := bytes.NewBufferString(`GET http://example.net/foo HTTP/1.1
Host: www.example.com:8000

`)
	requestBufRdr := bufio.NewReader(requestRdr)
	req, err := http.ReadRequest(requestBufRdr)
	assert.Nil(t, err)
	assert.Equal(t, req.Host, "example.net")
	assert.Equal(t, req.RequestURI, "http://example.net/foo")
}
