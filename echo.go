package serve2

import (
	"io"
	"net"
)

type EchoHandler struct{}

func (d EchoHandler) Handle(c net.Conn) net.Conn {
	go io.Copy(c, c)
	return nil
}

func (d EchoHandler) BytesRequired() int {
	return 4
}

func (d EchoHandler) Check(b []byte) bool {
	return string(b[:4]) == "ECHO"
}

func NewEchoHandler() *EchoHandler {
	return &EchoHandler{}
}
