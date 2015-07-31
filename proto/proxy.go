package proto

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"sync"
)

// proxy takes a ProxyConn "a", and a destination "dest" to dial, and forwards
// traffic between the connections. Only supports "TCP" at the current time.
func proxy(a net.Conn, proto, dest string) error {
	b, err := net.Dial(proto, dest)
	if err != nil {
		return err
	}

	var closer sync.Once
	closerFunc := func() {
		a.Close()
		b.Close()
	}

	go func() {
		io.Copy(a, b)
		closer.Do(closerFunc)
	}()

	go func() {
		io.Copy(b, a)
		closer.Do(closerFunc)
	}()

	return nil
}

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
func (p *Proxy) Handle(c net.Conn) (net.Conn, error) {
	return nil, proxy(c, p.proto, p.dest)
}

// Check checks the protocol.
func (p *Proxy) Check(b []byte) (bool, int) {
	if len(b) < len(p.match) {
		return false, len(p.match)
	}
	return bytes.Compare(b[:len(p.match)], p.match) == 0, 0
}

// NewProxy returns a fully initialized Proxy.
func NewProxy(match []byte, proto, dest string) *Proxy {
	return &Proxy{match, proto, dest}
}
