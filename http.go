package serve2

import (
	"net"
	"net/http"
)

var (
	// HTTP method names used for detection
	methods = []string{"GET", "HEAD", "POST", "PUT", "DELETE", "TRACE", "OPTIONS", "CONNECT", "PATCH"}
)

// HTTPProtoHandler provides a HTTP server in the form of a ProtocolHandler,
// with a custom http.Handler provided by the user.
type HTTPProtoHandler struct {
	listener *ChannelListener
}

// Setup installs the http handler, and stores the address for use of the ChannelListener
func (h *HTTPProtoHandler) Setup(handler http.Handler) {
	h.listener = NewChannelListener(make(chan net.Conn, 10), nil)

	httpServer := http.Server{Addr: ":http", Handler: handler}
	go httpServer.Serve(h.listener)
}

// Handle pushes the connection to the HTTP server
func (h *HTTPProtoHandler) Handle(c net.Conn) net.Conn {
	h.listener.Push(c)
	return nil
}

// Check looks through the known HTTP methods, returning true if there is a match
func (h *HTTPProtoHandler) Check(header []byte) bool {
	str := string(header)

	for _, v := range methods {
		if str[:len(v)] == v {
			return true
		}
	}
	return false

}

func (h *HTTPProtoHandler) BytesRequired() int {
	return 7
}

// NewHTTPProtoHandler returns a fully initialized HTTPProtoHandler
func NewHTTPProtoHandler(handler http.Handler) *HTTPProtoHandler {
	h := HTTPProtoHandler{}
	h.Setup(handler)
	return &h
}
