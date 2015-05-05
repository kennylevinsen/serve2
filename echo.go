package serve2

import (
	"io"
	"net"
)

// EchoProtoHandler is a simple protocol for testing purposes. It requires that the
// connection is initiated by writing "Echo", as protocol recognition would not
// work otherwise, but after that, it simply reads and echoes everything it
// receives, hence the name.
type EchoProtoHandler struct{}

func (d EchoProtoHandler) Handle(c net.Conn) net.Conn {
	go io.Copy(c, c)
	return nil
}

func (d EchoProtoHandler) BytesRequired() int {
	return 4
}

func (d EchoProtoHandler) Check(b []byte) bool {
	return string(b[:4]) == "ECHO"
}

func NewEchoProtoHandler() *EchoProtoHandler {
	return &EchoProtoHandler{}
}
