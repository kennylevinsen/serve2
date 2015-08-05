package proto

import (
	"net"

	"github.com/joushou/serve2/utils"
)

// ListenProxy provides a net.Listener whose Accept will only return matched
// protocols.
type ListenProxy struct {
	listener *utils.ChannelListener
	Checker  func([]byte, []interface{}) (bool, int)
	Desc     string
}

func (lp *ListenProxy) String() string {
	return lp.Desc
}

// Listener returns the proxy net.Listener.
func (lp *ListenProxy) Listener() net.Listener {
	return lp.listener
}

// Handle pushes the connection to the ListenProxy server.
func (lp *ListenProxy) Handle(c net.Conn) (net.Conn, error) {
	lp.listener.Push(c)
	return nil, nil
}

// Check just calls the ListenChecker.
func (lp *ListenProxy) Check(header []byte, hints []interface{}) (bool, int) {
	return lp.Checker(header, hints)
}

// NewListenProxy returns a fully initialized ListenProxy.
func NewListenProxy(checker func([]byte, []interface{}) (bool, int), buffer int) *ListenProxy {
	listener := utils.NewChannelListener(make(chan net.Conn, buffer), nil)
	return &ListenProxy{
		listener: listener,
		Checker:  checker,
		Desc:     "ListenProxy",
	}
}
