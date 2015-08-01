package proto

import (
	"net"
	"testing"
	"time"
)

type dummyConn struct {
	net.Conn
}

func (dummyConn) isAwesome() bool {
	return true
}

func TestNewListenProxy(t *testing.T) {
	checker := func([]byte) (bool, int) {
		return true, 0
	}

	lp := NewListenProxy(checker, 10)

	var l net.Listener
	l = lp.Listener()

	success := make(chan net.Conn, 1)

	go func() {
		c, err := l.Accept()
		if err != nil {
			t.Error("Accept() failed")
		}
		success <- c
	}()

	lp.Handle(dummyConn{})

	select {
	case <-time.After(300 * time.Millisecond):
		t.Error("timed out while waiting for Accept to return")
	case x := <-success:
		y, ok := x.(dummyConn)
		if !ok {
			t.Error("net.Conn received from Accept was incorrect")
		}

		if !y.isAwesome() {
			t.Error("dummyConn is not awesome")
		}
	}
}
