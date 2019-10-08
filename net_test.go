package gobt

import "testing"

func testAvailablePort(t *testing.T) {
	ln, port, err := availablePort()
	if err != nil {
		t.Errorf("port error: %v", err)
	}
	if port != 6881 {
		t.Errorf("port not 6881")
	}
	defer ln.Close()
}
