package serve2

import "net"

type DiscardHandler struct{}

func (d DiscardHandler) Handle(c net.Conn) net.Conn {
	go func(conn net.Conn) {
		s := make([]byte, 1024)
		for {
			_, err := conn.Read(s)
			if err != nil {
				return
			}
		}
	}(c)

	return nil
}

func (d DiscardHandler) BytesRequired() int {
	return 7
}

func (d DiscardHandler) Check(b []byte) bool {
	return string(b[:7]) == "DISCARD"
}

func NewDiscardHandler() *DiscardHandler {
	return &DiscardHandler{}
}
