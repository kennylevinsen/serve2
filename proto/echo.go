package proto

import (
	"io"
	"net"
)

// Echo is a simple protocol for testing purposes. It requires that the
// connection is initiated by writing "Echo", as protocol recognition would not
// work otherwise, but after that, it simply reads and echoes everything it
// receives, hence the name.
type Echo struct{}

func (Echo) String() string {
	return "Echo"
}

// Handle implements the ECHO protocol.
func (Echo) Handle(c net.Conn) (net.Conn, error) {
	go io.Copy(c, c)
	return nil, nil
}

// Check checks the protocol.
func (Echo) Check(b []byte) (bool, int) {
	if len(b) < 4 {
		return false, 4
	}
	return string(b[:4]) == "ECHO", 0
}

// NewEcho returns a new ECHO protocol handler.
func NewEcho() *Echo {
	return &Echo{}
}
