package proto

import (
	"bytes"
	"fmt"
	"io"
	"net"
)

// Proxy connects to an external protocol handler after establishing
// the connection type. An example use would be redirecting the connection to a
// non-Go SSH server.
type Proxy struct {
	match []byte
	proto string
	dest  string
}

func (p *Proxy) String() string {
	return fmt.Sprintf("Proxy [dest: %s]", p.dest)
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

// Check checks the protocol.
func (d *Proxy) Check(b []byte) (bool, int) {
	if len(b) < len(d.match) {
		return false, len(d.match)
	}
	return bytes.Compare(b, d.match) == 0, 0
}

// NewProxy returns a fully initialized Proxy.
func NewProxy(match []byte, proto, dest string) *Proxy {
	return &Proxy{match, proto, dest}
}
