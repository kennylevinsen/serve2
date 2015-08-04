package proto

import (
	"bytes"
	"fmt"
	"net"
	"sort"
)

type matchByLength [][]byte

func (ms matchByLength) Len() int           { return len(ms) }
func (ms matchByLength) Swap(i, j int)      { ms[i], ms[j] = ms[j], ms[i] }
func (ms matchByLength) Less(i, j int) bool { return len(ms[i]) < len(ms[j]) }

// MultiProxy connects to an external protocol handler after establishing
// the connection type. An example use would be redirecting the connection to a
// non-Go SSH server. Unlike Proxy, MultiProxy is capable of matching against
// more than one string, progressively requesting more data as nmecessary. This
// enables proxying to other HTTP servers by listing all the HTTP methods as
// match patterns, to give an example.
type MultiProxy struct {
	match [][]byte
	proto string
	dest  string
}

func (p *MultiProxy) String() string {
	return fmt.Sprintf("MultiProxy [dest: %s]", p.dest)
}

// Handle dials the destination, and establishes a simple proxy between the
// connecting party and the destination.
func (p *MultiProxy) Handle(c net.Conn) (net.Conn, error) {
	return nil, proxy(c, p.proto, p.dest)
}

// Check checks the protocol.
func (p *MultiProxy) Check(b []byte, _ []interface{}) (bool, int) {
	required := 0

	for _, v := range p.match {
		if len(v) > len(b) {

			if bytes.Compare(v[:len(b)], b) == 0 {
				// We found the smallest potential future match
				required = len(v)
				break
			}
		} else if bytes.Compare(b[:len(v)], v) == 0 {
			return true, 0
		}
	}

	return false, required
}

// NewMultiProxy returns a fully initialized MultiProxy.
func NewMultiProxy(match [][]byte, proto, dest string) *MultiProxy {
	matches := make([][]byte, len(match))
	for i := 0; i < len(matches); i++ {
		matches[i] = make([]byte, len(match[i]))
		copy(matches[i], match[i])
	}

	sort.Sort(matchByLength(matches))
	return &MultiProxy{matches, proto, dest}
}
