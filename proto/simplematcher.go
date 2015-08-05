package proto

import (
	"bytes"
	"net"
	"sort"
)

type matchByLength [][]byte

func (ms matchByLength) Len() int           { return len(ms) }
func (ms matchByLength) Swap(i, j int)      { ms[i], ms[j] = ms[j], ms[i] }
func (ms matchByLength) Less(i, j int) bool { return len(ms[i]) < len(ms[j]) }

// SimpleMatcher matches the provided bytes against a list of potential
// matches, quickly dismissing impossible matches.
type SimpleMatcher struct {
	Matches [][]byte
	Handler func(net.Conn) (net.Conn, error)
	Desc    string
}

// String returns the provided description.
func (s *SimpleMatcher) String() string {
	return s.Desc
}

// Handle calls the provided handler.
func (s *SimpleMatcher) Handle(c net.Conn) (net.Conn, error) {
	return s.Handler(c)
}

// Check looks through the provided matches.
func (s *SimpleMatcher) Check(header []byte, _ []interface{}) (bool, int) {
	for _, candidate := range s.Matches {
		if len(candidate) > len(header) {
			if bytes.Compare(candidate[:len(header)], header) == 0 {
				// We found the smallest potential future match
				return false, len(candidate)
			}
		} else if bytes.Compare(header[:len(candidate)], candidate) == 0 {
			return true, 0
		}
	}

	return false, 0
}

// Sort sorts the provided matches by length.
func (s *SimpleMatcher) Sort() {
	sort.Sort(matchByLength(s.Matches))
}

// NewSimpleMatcher returns a SimpleMatcher with the provided matches and
// handler, matches already sorted, and Desc initialized to "SimpleMatcher".
// Do note that this modified the provided matches slice to sort it by length.
func NewSimpleMatcher(matches [][]byte, handler func(net.Conn) (net.Conn, error)) *SimpleMatcher {
	sm := &SimpleMatcher{
		Matches: matches,
		Handler: handler,
		Desc:    "SimpleMatcher",
	}
	sm.Sort()
	return sm
}
