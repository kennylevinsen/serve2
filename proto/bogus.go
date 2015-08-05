package proto

import (
	"fmt"
	"io"
	"net"
	"net/http"

	"github.com/joushou/serve2/utils"
)

var (
	// HTTPMethods are the HTTP method names that can be used for detection,
	// sorted by length.
	HTTPMethods = [][]byte{
		[]byte("GET"),
		[]byte("PUT"),
		[]byte("HEAD"),
		[]byte("POST"),
		[]byte("PATCH"),
		[]byte("TRACE"),
		[]byte("DELETE"),
		[]byte("CONNECT"),
		[]byte("OPTIONS"),
	}
)

// NewHTTP returns a ListenProxy with a http as the listening service. This is
// a convenience wrapper kept in place for compatibility with older checkouts.
// Might be removed in the future.
func NewHTTP(handler http.Handler) *ListenProxy {
	sm := &SimpleMatcher{Matches: HTTPMethods}
	lp := NewListenProxy(sm.Check, 10)
	lp.Desc = "HTTP"

	httpServer := http.Server{Addr: ":http", Handler: handler}
	go httpServer.Serve(lp.Listener())
	return lp
}

// NewEcho returns a new ECHO protocol handler.
// Echo is a simple protocol for testing purposes. It requires that the
// connection is initiated by writing "Echo", as protocol recognition would not
// work otherwise, but after that, it simply reads and echoes everything it
// receives, hence the name.
func NewEcho() *SimpleMatcher {
	handler := func(c net.Conn) (net.Conn, error) {
		go io.Copy(c, c)
		return nil, nil
	}

	return &SimpleMatcher{
		Matches: [][]byte{
			[]byte("ECHO"),
		},
		Desc:    "Echo",
		Handler: handler,
	}
}

// NewDiscard returns a new Discard protocol handler.
// Discard is a simple protocol for testing purposes. It requires that
// the connection is initiated by writing "DISCARD", as protocol recognition
// would not work otherwise, but after that, it simply reads and discards
// everything it receives, hence the name.
func NewDiscard() *SimpleMatcher {
	handler := func(c net.Conn) (net.Conn, error) {
		go func(conn net.Conn) {
			s := make([]byte, 128)
			for {
				_, err := conn.Read(s)
				if err != nil {
					return
				}
			}
		}(c)
		return nil, nil
	}

	return &SimpleMatcher{
		Matches: [][]byte{
			[]byte("DISCARD"),
		},
		Desc:    "Discard",
		Handler: handler,
	}
}

// NewMultiProxy returns a SimpleMatcher set up to call DialAndProxy.
func NewMultiProxy(matches [][]byte, proto, dest string) *SimpleMatcher {
	handler := func(c net.Conn) (net.Conn, error) {
		return nil, utils.DialAndProxy(c, proto, dest)
	}

	sm := &SimpleMatcher{
		Matches: matches,
		Desc:    fmt.Sprintf("Proxy [dest: %s]", dest),
		Handler: handler,
	}

	sm.Sort()

	return sm
}

// NewProxy returns a SimpleMatcher set up to call DialAndProxy.
func NewProxy(match []byte, proto, dest string) *SimpleMatcher {
	return NewMultiProxy([][]byte{match}, proto, dest)
}
