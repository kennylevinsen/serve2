package serve2

import "net"

// DiscardProtoHandler is a simple protocol for testing purposes. It requires that
// the connection is initiated by writing "DISCARD", as protocol recognition
// would not work otherwise, but after that, it simply reads and discards
// everything it receives, hence the name.
type DiscardProtoHandler struct{}

func (d DiscardProtoHandler) Handle(c net.Conn) net.Conn {
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

func (d DiscardProtoHandler) BytesRequired() int {
	return 7
}

func (d DiscardProtoHandler) Check(b []byte) bool {
	return string(b[:7]) == "DISCARD"
}

func NewDiscardProtoHandler() *DiscardProtoHandler {
	return &DiscardProtoHandler{}
}
