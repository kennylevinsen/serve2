package serve2

import (
	"io"
	"net"
)

// ProxyProtoHandler connects to an external protocol handler after establishing
// the connection type. An example use would be redirecting the connection to a
// non-Go SSH server.
type ProxyProtoHandler struct {
	matchString string
	proto       string
	dest        string
}

// Handle dials the destination, and establishes a simple proxy between the
// connecting party and the destination.
func (d *ProxyProtoHandler) Handle(c net.Conn) net.Conn {
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

func (d *ProxyProtoHandler) BytesRequired() int {
	return len(d.matchString)
}

func (d *ProxyProtoHandler) Check(b []byte) bool {
	return string(b) == d.matchString
}

// NewProxyProtoHandler returns a fully initialized ProxyProtoHandler.
func NewProxyProtoHandler(matchString, proto, dest string) *ProxyProtoHandler {
	return &ProxyProtoHandler{matchString, proto, dest}
}
