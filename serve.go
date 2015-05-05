package serve2

import (
	"log"
	"net"
)

// ProtocolHandler is the protocol detection and handling interface used by serve2
type ProtocolHandler interface {
	// BytesRequired tells how many bytes the Checker needs
	BytesRequired() int

	// Checker informs if the bytes match the protocol
	Check([]byte) bool

	// Handle manages the protocol. In case of encapsulating protocol, Handle
	// can return a net.Conn which will be thrown through the entire protocol
	// management show again
	Handle(net.Conn) net.Conn
}

// Server handles a set of ProtocolHandlers
type Server struct {
	protocols   []ProtocolHandler
	bytesToRead int
}

// AddHandler registers a ProtocolHandler
func (s *Server) AddHandler(p ProtocolHandler) {
	s.protocols = append(s.protocols, p)
}

// AddHandlers registers a set of ProtocolHandlers
func (s *Server) AddHandlers(p ...ProtocolHandler) {
	for _, ph := range p {
		s.AddHandler(ph)
	}
}

// prepareHandlers sorts the protocols after how many bytes they require to
// detect their protocol (lowest first), and stores the highest number of bytes
// required.
func (s *Server) prepareHandlers() {
	s.bytesToRead = 0
	handlers := make([]ProtocolHandler, 0)

	for range s.protocols {
		smallest := -1
		for i, v := range s.protocols {
			if smallest == -1 || v.BytesRequired() < s.protocols[smallest].BytesRequired() {
				smallest = i
			}

			if v.BytesRequired() > s.bytesToRead {
				s.bytesToRead = v.BytesRequired()
			}
		}
		handlers = append(handlers, s.protocols[smallest])
		s.protocols = append(s.protocols[:smallest], s.protocols[smallest+1:]...)
	}

	s.protocols = handlers

	for _, v := range s.protocols {
		log.Printf("Handler: %T, bytes required: %d\n", v, v.BytesRequired())
	}
}

// HandleConnection runs a connection through protocol detection and handling
// as needed.
func (s *Server) HandleConnection(c net.Conn) {
	read := 0
	header := make([]byte, 0, s.bytesToRead)

	for _, ph := range s.protocols {
		required := ph.BytesRequired()
		if read < required {
			n, err := c.Read(header[read:required])
			if err != nil {
				return
			}
			header = header[:read+n]
			read = len(header)
		}

		if ph.Check(header) {
			log.Printf("Connection recognized as: %T\n", ph)
			proxy := NewProxyConn(c, header)

			x := ph.Handle(proxy)
			if x != nil {
				s.HandleConnection(x)
			}
			return
		}
	}

	c.Close()
	return
}

// Serve accepts connections on a listener, handling them as appropriate
func (s *Server) Serve(l net.Listener) {
	s.prepareHandlers()
	for {
		conn, err := l.Accept()
		if err != nil {
			panic(err)
		}

		go func() {
			s.HandleConnection(conn)
		}()
	}
}

// New returns a new Server
func New() *Server {
	return &Server{}
}
