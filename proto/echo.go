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

// Handle implements the ECHO protocol.
func (d Echo) Handle(c net.Conn) net.Conn {
	go io.Copy(c, c)
	return nil
}

// BytesRequired returns how many bytes are required to detech ECHO.
func (d Echo) BytesRequired() int {
	return 4
}

// Check checks the protocol.
func (d Echo) Check(b []byte) bool {
	return string(b[:4]) == "ECHO"
}

// NewEcho returns a new ECHO protocol handler.
func NewEcho() *Echo {
	return &Echo{}
}
