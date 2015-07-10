package serve2

import (
	"errors"
	"net"

	"github.com/joushou/serve2/utils"
)

const (
	// DefaultBytesToCheck default maximum amount of bytes to check
	DefaultBytesToCheck = 128
)

var (
	// Errors
	ErrReadFailed    = errors.New("read failed")
	ErrGreedyHandler = errors.New("remaining handlers too greedy; protocol not recognized")
	ErrNoMatch       = errors.New("protocol not recognized")
)

// ProtocolHandler is the protocol detection and handling interface used by
// serve2.
type ProtocolHandler interface {
	// Check informs if the bytes match the protocol. If there is not enough
	// data yet, it should return false and the wanted amount of bytes, allowing
	// future calls when more data is available. It does not need to return the
	// same every time, and incrementally checking more and more data is
	// allowed. Returning false and 0 bytes needed means that the protocol
	// handler is 100% sure that this is not the proper protocol, and will not
	// result in any further calls.
	Check([]byte) (ok bool, needed int)

	// Handle manages the protocol. In case of an encapsulating protocol, Handle
	// can return a net.Conn which will be thrown through the entire protocol
	// management show again.
	Handle(net.Conn) net.Conn
}

// Logger is used to provide logging functionality for serve2
type Logger func(format string, v ...interface{})

// Server handles a set of ProtocolHandlers.
type Server struct {
	// DefaultProtocol is the protocol fallback if no match is made
	DefaultProtocol ProtocolHandler

	// Logger is used for logging if set
	Logger Logger

	// BytesToCheck is the max amount of bytes to check
	BytesToCheck int

	protocols   []ProtocolHandler
	minimumRead int
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
	if s.Logger != nil {
		s.Logger("Sorting %d protocols", len(s.protocols))
	}

	var handlers []ProtocolHandler

	for range s.protocols {
		smallest := -1
		for i, v := range s.protocols {
			var contestant, current int
			_, contestant = v.Check(nil)
			if smallest == -1 {
				smallest = i
			} else {
				_, current = s.protocols[smallest].Check(nil)
				if contestant < current {
					smallest = i
				}
			}

		}
		handlers = append(handlers, s.protocols[smallest])
		s.protocols = append(s.protocols[:smallest], s.protocols[smallest+1:]...)
	}

	_, s.minimumRead = handlers[0].Check(nil)

	s.protocols = handlers
}

func (s *Server) handle(h ProtocolHandler, c net.Conn, header []byte) {
	proxy := utils.NewProxyConn(c, header)
	x := h.Handle(proxy)
	if x != nil {
		if s.Logger != nil {
			s.Logger("Handler %v for %v is a transport, running again", h, c.RemoteAddr())
		}
		s.HandleConnection(x)
	}
}

// HandleConnection runs a connection through protocol detection and handling
// as needed.
func (s *Server) HandleConnection(c net.Conn) error {
	var failureReason error

	header := make([]byte, 0, s.BytesToCheck)
	c.Read(header) // Read a bit of data

	handlers := make([]ProtocolHandler, len(s.protocols))
	copy(handlers, s.protocols)

	smallest := s.minimumRead
	for len(handlers) > 0 {
		// Read the required data
		n, err := c.Read(header[len(header):smallest])
		header = header[:len(header)+n]

		if n == 0 && err != nil {
			// Can't read anything
			failureReason = ErrReadFailed
			break
		}

		if len(header) < smallest {
			// We don't have enough data, try to read some more
			continue
		}

		smallest = -1
		for i := 0; i < len(handlers); i++ {
			handler := handlers[i]

			ok, required := handler.Check(header)
			if ok {
				if s.Logger != nil {
					s.Logger("Handling connection %v as %v", c.RemoteAddr(), handler)
				}
				s.handle(handler, c, header)
				return nil
			}

			if required == 0 {
				// The handler is 100% that it doesn't match, so remove it
				handlers = append(handlers[:i], handlers[i+1:]...)
				i--
			} else {
				// The handler is unsure and needs more data
				if smallest == -1 || required < smallest {
					smallest = required
				}
			}

		}

		if smallest > s.BytesToCheck {
			// The handlers want more data than we're allowed to read
			failureReason = ErrGreedyHandler
			break
		}

		if err != nil {
			// We can't read anymore
			failureReason = ErrNoMatch
			break
		}
	}

	if s.DefaultProtocol != nil {
		if s.Logger != nil {
			s.Logger("Defaulting %v to %v: [%v]", c.RemoteAddr(), s.DefaultProtocol, header)
		}
		s.handle(s.DefaultProtocol, c, header)
		return nil
	}

	// No one knew what was going on on this connection
	if s.Logger != nil {
		s.Logger("Handling %v failed: [%v]", c.RemoteAddr(), header)
	}
	c.Close()
	return failureReason
}

// Serve accepts connections on a listener, handling them as appropriate.
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

// New returns a new Server.
func New() *Server {
	return &Server{
		BytesToCheck: DefaultBytesToCheck,
		Logger:       nil,
	}
}
