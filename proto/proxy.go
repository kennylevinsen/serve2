package proto

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"

	"github.com/joushou/serve2/utils"
)

// proxy takes a ProxyConn "a", and a destination "dest" to dial, and forwards
// traffic between the connections. Only supports "TCP" at the current time.
func proxy(a net.Conn, proto, dest string) error {
	if proto != "tcp" {
		return errors.New("proxy only supports TCP for the time being")
	}

	aproxy, ok := a.(*utils.ProxyConn)
	if !ok {
		return errors.New("unable to convert source into *net.ProxyConn")
	}

	atcp, ok := aproxy.Conn.(*net.TCPConn)
	if !ok {
		return errors.New("unable to convert source into *net.TCPConn")
	}

	b, err := net.Dial(proto, dest)
	if err != nil {
		return err
	}

	btcp, ok := b.(*net.TCPConn)
	if !ok {
		return errors.New("unable to convert destination into *net.TCPConn")
	}

	go func() {
		io.Copy(a, b)
		atcp.CloseWrite()
		btcp.CloseRead()
	}()

	go func() {
		io.Copy(b, a)
		btcp.CloseWrite()
		atcp.CloseRead()
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
