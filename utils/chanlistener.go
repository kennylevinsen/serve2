package utils

import (
	"errors"
	"net"
)

// ChannelListener simulates a net.Listener, with Accept simply waiting on a
// channel to feed it connections. This is needed, as Go's http library does
// not provide a way to simply handle a single connection, but only supports
// accepting the connections itself, but can be used for anything that only
// accepts a net.Listener.
type ChannelListener struct {
	input chan net.Conn
	addr  net.Addr
}

// Accept waits on the ChannelListeners channel for new connections.
func (c *ChannelListener) Accept() (conn net.Conn, err error) {
	x, ok := <-c.input
	if !ok {
		return nil, errors.New("No more clients!")
	}
	return x, nil
}

// Close closes the channel.
func (c *ChannelListener) Close() error {
	close(c.input)
	return nil
}

// Addr returns the net.Addr of the listener.
func (c *ChannelListener) Addr() net.Addr {
	return c.addr
}

// Push pushes a net.Conn to the ChannelListeners channel.
func (c *ChannelListener) Push(conn net.Conn) {
	c.input <- conn
}

// NewChannelListener returns a fully iniitalized ChannelListener.
func NewChannelListener(input chan net.Conn, addr net.Addr) *ChannelListener {
	cl := ChannelListener{
		input: input,
		addr:  addr,
	}
	return &cl
}
