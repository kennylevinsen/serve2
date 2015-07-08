package proto

import "net"

// Discard is a simple protocol for testing purposes. It requires that
// the connection is initiated by writing "DISCARD", as protocol recognition
// would not work otherwise, but after that, it simply reads and discards
// everything it receives, hence the name.
type Discard struct{}

// Handle implements the DISCARD protocol (discarding all data).
func (d Discard) Handle(c net.Conn) net.Conn {
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

// BytesRequired returns how many bytes are required to detect DISCARD.
func (d Discard) BytesRequired() int {
	return 7
}

// Check checks the protocol.
func (d Discard) Check(b []byte) bool {
	return string(b[:7]) == "DISCARD"
}

// NewDiscard returns a new Discard protocol handler.
func NewDiscard() *Discard {
	return &Discard{}
}
