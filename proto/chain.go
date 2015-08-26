package proto

import "net"

// Chain allows for easy chaining of multiple checks, which all must pass before
// the handler is called. For performance, have the easiest checks first in the
// line of checkers.
type Chain struct {
	Handler     func(net.Conn) (net.Conn, error)
	Checkers    []func([]byte, []interface{}) (bool, int)
	Description string
}

// Handle calls the provided Handler.
func (cm *Chain) Handle(c net.Conn) (net.Conn, error) {
	return cm.Handler(c)
}

// Check alls all the provided checkers in order, returning the output of the
// first failing check. If all checkers pass, Check returns (true, 0).
func (cm *Chain) Check(header []byte, hints []interface{}) (bool, int) {
	for _, check := range cm.Checkers {
		ok, n := check(header, hints)
		if !ok {
			return ok, n
		}
	}

	return true, 0
}

// NewChain returns a Chain initialized with the provided handler and list of
// checkers.
func NewChain(handler func(net.Conn) (net.Conn, error), checkers ...func([]byte, []interface{}) (bool, int)) *Chain {
	return &Chain{
		Handler:     handler,
		Checkers:    checkers,
		Description: "Chain",
	}
}
