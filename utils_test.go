package main

import (
	"bytes"
	"errors"
	"log"
	"testing"
)

func TestProcessConnectionHdr(t *testing.T) {
	headers := make(map[string]struct{})
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

func TestTransfer(t *testing.T) {
	var src = "Hello world, this is a bunch of data that is being transferred during a CONNECT session."
	srcBuf := bytes.NewBufferString(src)
	dstBuf := closeBuffer{w: bytes.Buffer{}}
	transfer(&dstBuf, srcBuf)
	if dstBuf.String() != src {
		t.Errorf("Failed to correctly transfer data from source to destination, source: '%s', destination: '%s'.", src, dstBuf.String())
	}
}

type closeBuffer struct {
	w bytes.Buffer
}

func (c *closeBuffer) Write(p []byte) (int, error) {
	return c.w.Write(p)
}

func (*closeBuffer) Close() error {
	return nil
}

func (c *closeBuffer) String() string {
	return c.w.String()
}
