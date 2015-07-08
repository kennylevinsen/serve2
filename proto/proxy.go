package proto

import (
	"io"
	"net"
)

// Proxy connects to an external protocol handler after establishing
// the connection type. An example use would be redirecting the connection to a
// non-Go SSH server.
type Proxy struct {
	matchString string
	proto       string
	dest        string
}

// Handle dials the destination, and establishes a simple proxy between the
// connecting party and the destination.
func (d *Proxy) Handle(c net.Conn) net.Conn {
	pconn, err := net.Dial(d.proto, d.dest)
	if err != nil {
		return nil
	}

	// Proxy the connection
	go func() {
		io.Copy(pconn, c)
		pconn.Close()
	}()
	go func() {
		io.Copy(c, pconn)
		c.Close()
	}()

	return nil
}

// BytesRequired returns how many bytes are required to detect the protocol.
func (d *Proxy) BytesRequired() int {
	return len(d.matchString)
}

// Check checks the protocol.
func (d *Proxy) Check(b []byte) bool {
	return string(b) == d.matchString
}

// NewProxy returns a fully initialized Proxy.
func NewProxy(matchString, proto, dest string) *Proxy {
	return &Proxy{matchString, proto, dest}
}
