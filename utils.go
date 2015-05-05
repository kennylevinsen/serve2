package serve2

import (
	"errors"
	"net"
)

// ChannelListener simulates a net.Listener, with Accept simply waiting on a
// channel to feed it connections. This is needed, as Go's http library does
// not provide a way to simply handle a single connection, but only supports
// accepting the connections itself.
type ChannelListener struct {
	input chan net.Conn
	addr  net.Addr
}

// Accept waits on the ChannelListeners channel for new connections
func (c *ChannelListener) Accept() (conn net.Conn, err error) {
	x, ok := <-c.input
	if !ok {
		return nil, errors.New("No more clients!")
	}
	return x, nil
}

// Close closes the channel
func (c *ChannelListener) Close() error {
	close(c.input)
	return nil
}

// Addr returns the net.Addr of the listener
func (c *ChannelListener) Addr() net.Addr {
	return c.addr
}

// Push pushes a net.Conn to the ChannelListeners channel
func (c *ChannelListener) Push(conn net.Conn) {
	c.input <- conn
}

// NewChannelListener returns a fully iniitalized ChannelListener
func NewChannelListener(input chan net.Conn, addr net.Addr) *ChannelListener {
	cl := ChannelListener{
		input: input,
		addr:  addr,
	}
	return &cl
}

// ProxyConn simulates reads for the buffered content. When buffer is empty, it
// simply behaves like the net.Conn it wraps.
type ProxyConn struct {
	net.Conn
	buffer []byte
	active bool
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
		return fromBuffer, nil
	}

	return c.Conn.Read(p)
}

// NewProxyConn returns a fully initialized ProxyConn
func NewProxyConn(c net.Conn, buffer []byte) *ProxyConn {
	pc := ProxyConn{
		Conn:   c,
		buffer: buffer,
		active: true,
	}
	return &pc
}
