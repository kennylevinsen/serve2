package proto

import (
	"net"
	"net/http"

	"github.com/joushou/serve2/utils"
)

var (
	// HTTP method names used for detection. Must be sorted by length.
	methods = []string{"GET", "PUT", "HEAD", "POST", "TRACE", "PATCH", "DELETE", "OPTIONS", "CONNECT"}
)

// HTTP provides a HTTP server in the form of a ProtocolHandler,
// with a custom http.Handler provided by the user.
type HTTP struct {
	listener *utils.ChannelListener
}

func (HTTP) String() string {
	return "HTTP"
}

// Setup installs the http handler, and stores the address for use of the
// ChannelListener.
func (h *HTTP) Setup(handler http.Handler) {
	h.listener = utils.NewChannelListener(make(chan net.Conn, 10), nil)

	httpServer := http.Server{Addr: ":http", Handler: handler}
	go httpServer.Serve(h.listener)
}

// Handle pushes the connection to the HTTP server.
func (h *HTTP) Handle(c net.Conn) net.Conn {
	h.listener.Push(c)
	return nil
}

// Check looks through the known HTTP methods, returning true if there is a
// match.
func (h *HTTP) Check(header []byte) (bool, int) {
	str := string(header)
	required := 0

	for _, v := range methods {
		if len(v) > len(str) {
			if v[:len(str)] == str {
				// We found the smallest potential future match
				required = len(v)
				break
			}
		} else if str[:len(v)] == v {
			return true, 0
		}
	}

	return false, required

}

// NewHTTP returns a fully initialized HTTP.
func NewHTTP(handler http.Handler) *HTTP {
	h := HTTP{}
	h.Setup(handler)
	return &h
}
