package proto

import "net"

// Discard is a simple protocol for testing purposes. It requires that
// the connection is initiated by writing "DISCARD", as protocol recognition
// would not work otherwise, but after that, it simply reads and discards
// everything it receives, hence the name.
type Discard struct{}

func (Discard) String() string {
	return "Discard"
}

// Handle implements the DISCARD protocol (discarding all data).
func (Discard) Handle(c net.Conn) (net.Conn, error) {
	go func(conn net.Conn) {
		s := make([]byte, 1024)
		for {
			_, err := conn.Read(s)
			if err != nil {
				return
			}
		}
	}(c)

	return nil, nil
}

// Check checks the protocol.
func (Discard) Check(b []byte, _ []interface{}) (bool, int) {
	if len(b) < 7 {
		return false, 7
	}
	return string(b[:7]) == "DISCARD", 0
}

// NewDiscard returns a new Discard protocol handler.
func NewDiscard() *Discard {
	return &Discard{}
}
