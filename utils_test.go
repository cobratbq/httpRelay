package main

import (
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
	for src, tgt := range tests {
		result = fullHost(src)
		if result != tgt {
			t.Errorf("Expected '%s' to result in '%s', but instead is '%s'.", src, tgt, result)
		}
	}
}
