package httpRelay

import "testing"

func TestNopDialerDial(t *testing.T) {
	d := &NopDialer{}
	if conn, err := d.Dial("tcp", "www.google.com"); conn != nil || err == nil {
		t.Fatal("Expected error from NopDialer, but got actual connection or no error.")
	}
}
