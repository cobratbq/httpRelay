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
