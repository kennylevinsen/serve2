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

// Errors
var (
	ErrGreedyHandler = errors.New("remaining handlers too greedy")
)

// Protocol is the protocol detection and handling interface used by serve2.
type Protocol interface {
	// Check informs if the bytes match the protocol. If there is not enough
	// data yet, it should return false and the wanted amount of bytes, allowing
	// future calls when more data is available. It does not need to return the
	// same every time, and incrementally checking more and more data is
	// allowed. Returning false and 0 bytes needed means that the protocol
	// handler is 100% sure that this is not the proper protocol, and will not
	// result in any further calls.
	// Check, when called with nil, nil, must return false, N, where N is the
	// smallest amount of bytes that makes sense to call Check with.
	Check(header []byte, hints []interface{}) (ok bool, needed int)

	// Handle manages the protocol. In case of an encapsulating protocol, Handle
	// can return a net.Conn which will be thrown through the entire protocol
	// management show again.
	Handle(c net.Conn) (net.Conn, error)

	// String returns a pretty representation of the protocol to be used for
	// logging purposes.
	String() string
}

// Logger is used to provide logging functionality for serve2
type Logger func(format string, v ...interface{})

// Server handles a set of Protocols.
type Server struct {
	// DefaultProtocol is the protocol fallback if no match is made
	DefaultProtocol Protocol

	// Logger is used for logging if set
	Logger Logger

	// BytesToCheck is the max amount of bytes to check
	BytesToCheck int

	protocols   []Protocol
	minimumRead int
}

// AddHandler registers a Protocol
func (s *Server) AddHandler(p Protocol) {
	s.protocols = append(s.protocols, p)
}

// AddHandlers registers a set of Protocols
func (s *Server) AddHandlers(p ...Protocol) {
	for _, ph := range p {
		s.AddHandler(ph)
	}
}

// prepareHandlers sorts the protocols after how many bytes they require to
// detect their protocol (lowest first), and stores the highest number of bytes
// required.
func (s *Server) prepareHandlers() {
	var handlers []Protocol

	for range s.protocols {
		smallest := -1
		for i, v := range s.protocols {
			var contestant, current int
			_, contestant = v.Check(nil, nil)
			if smallest == -1 {
				smallest = i
			} else {
				_, current = s.protocols[smallest].Check(nil, nil)
				if contestant < current {
					smallest = i
				}
			}

		}
		handlers = append(handlers, s.protocols[smallest])
		s.protocols = append(s.protocols[:smallest], s.protocols[smallest+1:]...)
	}

	_, s.minimumRead = handlers[0].Check(nil, nil)

	s.protocols = handlers

	if s.Logger != nil {
		s.Logger("Sorted %d protocols:", len(s.protocols))

		for _, protocol := range s.protocols {
			s.Logger("\t%v", protocol)
		}
	}
}

func (s *Server) handle(h Protocol, c net.Conn, hints []interface{}, header []byte, readErr error) {
	proxy := utils.NewProxyConn(c, header, readErr)
	proxy.SetHints(hints)

	transport, err := h.Handle(proxy)
	if err != nil {
		s.Logger("Handling %v as %v failed: %v", c.RemoteAddr(), h, err)
	}

	if transport != nil {
		if s.Logger != nil {
			s.Logger("Handling %v as %v (transport)", c.RemoteAddr(), h)
		}
		if x, ok := transport.(utils.HintedConn); ok {
			hints = x.Hints()
		}
		s.HandleConnection(transport, hints)
	} else {
		if s.Logger != nil {
			s.Logger("Handling %v as %v", c.RemoteAddr(), h)
		}
	}
}

// HandleConnection runs a connection through protocol detection and handling
// as needed.
func (s *Server) HandleConnection(c net.Conn, hints []interface{}) error {
	var (
		failureReason, err error
		n                  int
		header             = make([]byte, 0, s.BytesToCheck)
		handlers           = make([]Protocol, len(s.protocols))
	)

	if hints == nil {
		hints = make([]interface{}, 0)
	}

	copy(handlers, s.protocols)

	smallest := s.minimumRead

	// This loop runs until we are out of candidate handlers, or until a handler
	// is selected.
	for len(handlers) > 0 {
		// Read the required data
		n, err = c.Read(header[len(header):smallest])
		header = header[:len(header)+n]

		if n == 0 && err != nil {
			// Can't read anything
			failureReason = err
			break
		}

		if len(header) < smallest {
			// We don't have enough data, try to read some more
			continue
		}

		smallest = -1

		// We run the current data through all candidate handlers.
		for i := 0; i < len(handlers); i++ {
			handler := handlers[i]

			ok, required := handler.Check(header, hints)
			if ok {
				s.handle(handler, c, hints, header, err)
				return nil
			}

			if required == 0 {
				// The handler is sure that it doesn't match, so remove it.
				handlers = append(handlers[:i], handlers[i+1:]...)
				i--
			} else if required <= len(header) {
				// Handler is broken, requesting less than we already gave it, so
				// we remove it.
				if s.Logger != nil {
					s.Logger("Handler %v is requesting %d bytes, but already read %d bytes. Skipping.",
						handler, required, len(header))
				}

				handlers = append(handlers[:i], handlers[i+1:]...)
				i--
			} else if smallest == -1 || required < smallest {
				// The handler needs more data to be certain.
				smallest = required
			}

		}

		if smallest > s.BytesToCheck {
			// The handlers want more data than we're allowed to read
			if s.Logger != nil {
				s.Logger("Next check requires %d bytes, but maximum read size set to %d",
					smallest, s.BytesToCheck)
			}
			failureReason = ErrGreedyHandler
			break
		}
	}

	if failureReason != nil && s.Logger != nil {
		s.Logger("Protocol detection failure: %v", failureReason)
	}

	if s.DefaultProtocol != nil {
		if s.Logger != nil {
			s.Logger("Defaulting %v: [%q]", c.RemoteAddr(), header)
		}
		s.handle(s.DefaultProtocol, c, hints, header, err)
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
			s.HandleConnection(conn, nil)
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
