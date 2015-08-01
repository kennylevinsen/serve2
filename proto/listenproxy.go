package proto

import (
	"net"

	"github.com/joushou/serve2/utils"
)

// ListenChecker is the provided Check function for ListenProxy, identical to
// ProtocolHandler.Check
type ListenChecker func(header []byte, hints []interface{}) (match bool, required int)

// ListenProxy provides a net.Listener whose Accept will only return matched
// protocols.
type ListenProxy struct {
	listener *utils.ChannelListener
	Checker  ListenChecker
}

func (ListenProxy) String() string {
	return "ListenProxy"
}

// Listener returns the proxy net.Listener.
func (l *ListenProxy) Listener() net.Listener {
	return l.listener
}

// Handle pushes the connection to the ListenProxy server.
func (l *ListenProxy) Handle(c net.Conn) (net.Conn, error) {
	l.listener.Push(c)
	return nil, nil
}

// Check just calls the ListenChecker.
func (l *ListenProxy) Check(header []byte, hints []interface{}) (bool, int) {
	return l.Checker(header, hints)
}

// NewListenProxy returns a fully initialized ListenProxy.
func NewListenProxy(checker ListenChecker, buffer int) *ListenProxy {
	listener := utils.NewChannelListener(make(chan net.Conn, buffer), nil)
	return &ListenProxy{listener: listener, Checker: checker}
}
