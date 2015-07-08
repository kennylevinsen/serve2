package utils

import (
	"net"
	"testing"
)

func TestChannelListenerAccept(t *testing.T) {

	cl := NewChannelListener(make(chan net.Conn, 10), nil)

	cl.Push(nil)
	_, err := cl.Accept()
	if err != nil {
		t.Errorf("Unexpected result. Accept failed with: %s", err)
	}

	cl.Close()
	_, err = cl.Accept()
	if err == nil {
		t.Errorf("Unexpected result. Accept did not fail.")
	}
}
