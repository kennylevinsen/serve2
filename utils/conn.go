package utils

import (
	"crypto/tls"
	"io"
	"net"
	"sync"
)

// HintedConn describes a net.Conn which also provides hints about transports.
type HintedConn interface {
	net.Conn

	// Hints retrieves the current hints.
	Hints() []interface{}
}

// HintConn is a simple net.Conn wrapper implemented HintedConn.
type HintConn struct {
	net.Conn
	hints []interface{}
}

// Hints retrieves the current hints.
func (h *HintConn) Hints() []interface{} {
	return h.hints
}

// NewHintConn provides a new *HintConn.
func NewHintConn(c net.Conn, hints []interface{}) *HintConn {
	return &HintConn{
		Conn:  c,
		hints: hints,
	}
}

// GetHints is a convenience function for retrieving hints from a net.Conn if
// available, otherwise returning nil.
func GetHints(c net.Conn) []interface{} {
	if hintedConn, ok := c.(HintedConn); ok {
		return hintedConn.Hints()
	}
	return nil
}

// ProxyConn simulates reads for the buffered content. When buffer is empty, it
// simply behaves like the net.Conn it wraps.
type ProxyConn struct {
	net.Conn
	hints     []interface{}
	buffer    []byte
	storedErr error
	active    bool
}

// Hints return the stored hints.
func (c *ProxyConn) Hints() []interface{} {
	return c.hints
}

// SetHints sets the stored hints.
func (c *ProxyConn) SetHints(hints []interface{}) {
	c.hints = hints
}

// Read reads data from the connection.  If buffer is available, it will try to
// serve the request from the buffer alone. If the buffer is empty, it simply
// calls read.
func (c *ProxyConn) Read(p []byte) (int, error) {
	if c.active {
		fromBuffer := len(p)
		if fromBuffer > len(c.buffer) {
			fromBuffer = len(c.buffer)
		}

		copy(p, c.buffer[:fromBuffer])
		c.buffer = c.buffer[fromBuffer:]

		if len(c.buffer) == 0 {
			c.active = false
			c.buffer = nil
		}
		return fromBuffer, c.storedErr
	}

	return c.Conn.Read(p)
}

// NewProxyConn returns a fully initialized ProxyConn.
func NewProxyConn(c net.Conn, buffer []byte, storedErr error) *ProxyConn {
	pc := ProxyConn{
		Conn:      c,
		buffer:    buffer,
		storedErr: storedErr,
		active:    len(buffer) > 0 || storedErr != nil,
	}
	return &pc
}

// DialAndProxy takes a net.Conn "a", and a destination "dest" to dial, and
// forwards traffic between the connections.
func DialAndProxy(a net.Conn, proto, dest string) error {
	b, err := net.Dial(proto, dest)
	if err != nil {
		return err
	}

	proxy(a, b)
	return nil
}

// DialAndProxyTLS takes a net.Conn "a", and a destination "dest" to dial with
// TLS, and forwards traffic between the connections.
func DialAndProxyTLS(a net.Conn, proto, dest string, config *tls.Config) error {
	b, err := tls.Dial(proto, dest, config)
	if err != nil {
		return err
	}

	proxy(a, b)
	return nil
}

// proxy takes a net.Conn "a" and a net.Conn "b", and forwards traffic between
// the connections.
func proxy(a net.Conn, b net.Conn) {
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
}
