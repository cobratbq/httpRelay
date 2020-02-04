package httprelay

import (
	"bufio"
	"bytes"
	"errors"
	"log"
	"net"
	"net/http"
	"testing"
)

func TestProcessConnectionHdr(t *testing.T) {
	headers := map[string]struct{}{}
	processConnectionHdr(headers, "Keep-Alive, Foo ,Bar")
	if len(headers) != 3 {
		t.Error("Expected exactly 3 entries in headers map.")
	}
	if _, ok := headers["Keep-Alive"]; !ok {
		t.Error("Expected header Keep-Alive in map.")
	}
	if _, ok := headers["Foo"]; !ok {
		t.Error("Expected header Foo in map.")
	}
	if _, ok := headers["Bar"]; !ok {
		t.Error("Expected header Bar in map.")
	}
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
		if result != dst {
			t.Errorf("Expected '%s' to result in '%s', but instead is '%s'.", src, dst, result)
		}
	}
}

func TestLogError(t *testing.T) {
	logError(nil, "")
	logError(errors.New("test error"), "")
	logError(nil, "my prefix")
	logError(errors.New("test error"), "my prefix")
}

func TestLogRequest(t *testing.T) {
	req := http.Request{Method: "GET", Host: "localhost:1414", Proto: "HTTP/1.1"}
	logRequest(&req)
}

func TestProcessConnectionHdrs(t *testing.T) {
	var hdrs = map[string]struct{}{}
	var val = "Keep-Alive  ,  \tFoo,bar"
	processConnectionHdr(hdrs, val)
	if len(hdrs) != 3 {
		t.Error("Expected 3 header entries.")
	}
	var ok bool
	if _, ok = hdrs["Keep-Alive"]; !ok {
		t.Error("Expected to find header Keep-Alive.")
	}
	if _, ok = hdrs["Foo"]; !ok {
		t.Error("Expected to find header Foo.")
	}
	if _, ok = hdrs["bar"]; !ok {
		t.Error("Expected to find header bar.")
	}
}

func TestProcessConnectionHdrsBad(t *testing.T) {
	var hdrs = map[string]struct{}{}
	var val = "Illegal spaces, Capiche?, close"
	processConnectionHdr(hdrs, val)
	if len(hdrs) != 1 {
		t.Error("Expected only 1 header, since others were bad.")
	}
	log.Printf("Headers: %#v\n", hdrs)
	var ok bool
	if _, ok = hdrs["Illegal spaces"]; ok {
		t.Error("Header with spaces should not be in the headers map.")
	}
	if _, ok = hdrs["Capiche?"]; ok {
		t.Error("Header with bad chars should not be in the headers map.")
	}
	if _, ok = hdrs["close"]; !ok {
		t.Error("Expected the one correct header 'close' to be in the map.")
	}
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
	var ok bool
	var k string
	if len(dst) != 3 {
		t.Errorf("Expected exactly 2 headers, but found a different number: %#v\n", dst)
	}
	for k = range hopByHopHeaders {
		// check simple dropped headers
		if _, ok = dst[k]; ok {
			t.Error("Did not expect header in destination map. It should be dropped:", k)
		}
	}
	for _, k = range []string{"Connection", "Keep-Alive", "Foo"} {
		// check special treatment of Connection header and its values
		if _, ok = dst[k]; ok {
			t.Error("Did not expect Connection header of listings in destination map. It should be dropped:", k)
		}
	}
	for _, k = range []string{"Content-Type", "Content-Encoding", "Via"} {
		// check remaining headers
		if _, ok = dst[k]; !ok {
			t.Error("Expected header in destination map:", k)
		}
	}
}

func TestTransfer(t *testing.T) {
	var src = "Hello world, this is a bunch of data that is being transferred during a CONNECT session."
	srcBuf := bytes.NewBufferString(src)
	dstBuf := closeBuffer{}
	transfer(&dstBuf, srcBuf, "buffer to buffer")
	if dstBuf.String() != src {
		t.Errorf("Failed to correctly transfer data from source to destination, source: '%s', destination: '%s'.", src, dstBuf.String())
	}
}

func TestTransferError(t *testing.T) {
	var src = "Hello world, this is a bunch of data that is being transferred during a CONNECT session."
	srcBuf := bytes.NewBufferString(src)
	dstBuf := errorBuffer{}
	transfer(&dstBuf, srcBuf, "buffer to buffer")
	if dstBuf.String() != "Hello worl" {
		t.Errorf("Expected only part of message 'Hello worl' but got '%s'.", dstBuf.String())
	}
}

func TestNonHijackableWriter(t *testing.T) {
	var writer nonHijackableWriter
	conn, err := acquireConn(&writer)
	if err != ErrNonHijackableWriter {
		t.Error("Expected to receive non-hijackable writer error but got:", err.Error())
	}
	if conn != nil {
		t.Error("Expected to get no connection but got something anyway:", conn)
	}
}

func TestHijackableWriter(t *testing.T) {
	var writer hijackableWriter
	conn, err := acquireConn(&writer)
	if err != nil {
		t.Error("Expected to get no error but got:", err.Error())
	}
	if conn == nil {
		t.Error("Expected to get a connection but got nothing.")
	}
}

type closeBuffer struct {
	bytes.Buffer
}

func (*closeBuffer) Close() error {
	return nil
}

type errorBuffer struct {
	bytes.Buffer
}

func (e *errorBuffer) Write(p []byte) (int, error) {
	_, _ = e.Buffer.Write(p[:10])
	return 10, errors.New("bad stuff happened")
}

func (*errorBuffer) Close() error {
	return nil
}

type nonHijackableWriter struct{}

func (*nonHijackableWriter) Header() http.Header {
	return nil
}

func (*nonHijackableWriter) Write([]byte) (int, error) {
	return 0, nil
}

func (*nonHijackableWriter) WriteHeader(int) {
	return
}

type hijackableWriter struct {
	nonHijackableWriter
}

func (*hijackableWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	pipeR, _ := net.Pipe()
	return pipeR, nil, nil
}
