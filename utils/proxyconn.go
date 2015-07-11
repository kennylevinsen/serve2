package utils

import (
	"net"
)

// ProxyConn simulates reads for the buffered content. When buffer is empty, it
// simply behaves like the net.Conn it wraps.
type ProxyConn struct {
	net.Conn
	buffer    []byte
	storedErr error
	active    bool
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
