package httprelay

import (
	"log"
	"net/http"
	"testing"

	assert "github.com/cobratbq/goutils/std/testing"
)

func TestProcessConnectionHdr(t *testing.T) {
	headers := map[string]struct{}{}
	processConnectionHdr(headers, "Keep-Alive, Foo ,Bar")
	assert.Equal(t, len(headers), 3)
	assert.ElementPresent(t, headers, "Keep-Alive")
	assert.ElementPresent(t, headers, "Foo")
	assert.ElementPresent(t, headers, "Bar")
}

func TestFullHost(t *testing.T) {
	var tests = map[string]string{
		"localhost":          "localhost:80",
		"localhost:1234":     "localhost:1234",
		"bla":                "bla:80",
		"www.google.com":     "www.google.com:80",
		"www.google.com:80":  "www.google.com:80",
		"www.google.com:443": "www.google.com:443",
		"google.com:8080":    "google.com:8080",
	}
	var result string
	for src, dst := range tests {
		result = fullHost(src)
		assert.Equal(t, result, dst)
	}
}

func TestLogRequest(t *testing.T) {
	req := http.Request{Method: "GET", Host: "localhost:1414", Proto: "HTTP/1.1"}
	logRequest(&req)
}

func TestProcessConnectionHdrs(t *testing.T) {
	var hdrs = map[string]struct{}{}
	var val = "Keep-Alive  ,  \tFoo,bar"
	processConnectionHdr(hdrs, val)
	assert.Equal(t, len(hdrs), 3)
	assert.ElementPresent(t, hdrs, "Keep-Alive")
	assert.ElementPresent(t, hdrs, "Foo")
	assert.ElementPresent(t, hdrs, "bar")
}

func TestProcessConnectionHdrsBad(t *testing.T) {
	var hdrs = map[string]struct{}{}
	var val = "Illegal spaces, Capiche?, close"
	processConnectionHdr(hdrs, val)
	assert.Equal(t, len(hdrs), 1)
	log.Printf("Headers: %#v\n", hdrs)
	assert.ElementAbsent(t, hdrs, "Illegal spaces")
	assert.ElementAbsent(t, hdrs, "Capiche?")
	assert.ElementPresent(t, hdrs, "close")
}

func TestCopyHeaders(t *testing.T) {
	var src = http.Header{}
	src.Add("Transfer-Encoding", "gzip")
	src.Add("Content-Type", "image/jpeg")
	src.Add("Trailer", "something")
	src.Add("Content-Encoding", "gzip")
	src.Add("Via", "bla:1234")
	src.Add("Connection", "Keep-Alive, Foo")
	src.Add("Keep-Alive", "close")
	src.Add("Foo", "Bar")
	var dst = http.Header{}
	copyHeaders(dst, src)
	var k string
	if len(dst) != 3 {
		t.Errorf("Expected exactly 2 headers, but found a different number: %#v\n", dst)
	}
	for k = range hopByHopHeaders {
		// check simple dropped headers
		assert.KeyAbsent(t, dst, k)
	}
	for _, k = range []string{"Connection", "Keep-Alive", "Foo"} {
		// check special treatment of Connection header and its values
		assert.KeyAbsent(t, dst, k)
	}
	for _, k = range []string{"Content-Type", "Content-Encoding", "Via"} {
		// check remaining headers
		assert.KeyPresent(t, dst, k)
	}
}
