package proto

import "net/http"

var (
	// HTTP method names used for detection. Must be sorted by length.
	methods = []string{"GET", "PUT", "HEAD", "POST", "TRACE", "PATCH", "DELETE", "OPTIONS", "CONNECT"}
)

// Check looks through the known HTTP methods, returning true if there is a
// match.
func checker(header []byte, _ []interface{}) (bool, int) {
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

// NewHTTP returns a ListenProxy with a http as the listening service. This is
// a convenience wrapper kept in place for compatibility with older checkouts.
// Might be removed in the future.
func NewHTTP(handler http.Handler) *ListenProxy {
	lp := NewListenProxy(checker, 10)

	httpServer := http.Server{Addr: ":http", Handler: handler}
	go httpServer.Serve(lp.Listener())
	return lp
}
