package proto

import (
	"crypto/tls"
	"crypto/x509"
	"net"
)

// TLSMatcherChecks is a bitmask describing what to verify.
type TLSMatcherChecks uint

// TLSMatcher verification flags.
const (
	TLSCheckServerName TLSMatcherChecks = 1 << iota
	TLSCheckNegotiatedProtocol
	TLSCheckNegotiatedProtocolIsMutual
	TLSCheckClientCertificate
	TLSCheckCipherSuite
	TLSCheckVersion
)

// IsSet checks the bitmask for the given bit.
func (tcv TLSMatcherChecks) IsSet(other TLSMatcherChecks) bool {
	return (tcv & other) != 0
}

// TLSMatcher is a TLS connection inspector, that will verify the
// tls.ConnectionState fields described by Verifications. If no verifications
// are enabled, TLSMatcher will simply match the presence of a TLS transport.
type TLSMatcher struct {
	ServerName                 string
	NegotiatedProtocol         string
	NegotiatedProtocolIsMutual bool
	PeerCertificates           []*x509.Certificate
	CipherSuite                uint16
	Version                    uint16
	Verifications              TLSMatcherChecks
	Handler                    func(net.Conn) (net.Conn, error)
}

type connectionStater interface {
	ConnectionState() tls.ConnectionState
}

// Check inspects the last transport hint, checking if it is a TLS transport.
func (tc *TLSMatcher) Check(_ []byte, hints []interface{}) (bool, int) {
	if len(hints) == 0 {
		return false, 0
	}

	h := hints[len(hints)-1]

	c, ok := h.(connectionStater)
	if !ok {
		return false, 0
	}

	cs := c.ConnectionState()

	if tc.Verifications.IsSet(TLSCheckServerName) && tc.ServerName != cs.ServerName {
		return false, 0
	}

	if tc.Verifications.IsSet(TLSCheckNegotiatedProtocol) &&
		tc.NegotiatedProtocol != cs.NegotiatedProtocol {
		return false, 0
	}

	if tc.Verifications.IsSet(TLSCheckNegotiatedProtocolIsMutual) &&
		tc.NegotiatedProtocolIsMutual != cs.NegotiatedProtocolIsMutual {
		return false, 0
	}

	if tc.Verifications.IsSet(TLSCheckClientCertificate) {
		// TODO: Is this how you do check the clients certificate? :/

		for _, cert := range tc.PeerCertificates {
			if cs.PeerCertificates[0].Equal(cert) {
				goto certOk
			}
		}
		return false, 0
	certOk:
	}

	if tc.Verifications.IsSet(TLSCheckCipherSuite) && tc.CipherSuite != cs.CipherSuite {
		return false, 0
	}

	if tc.Verifications.IsSet(TLSCheckVersion) && tc.Version != cs.Version {
		return false, 0
	}

	return true, 0
}

// Handle simply calls the provided handler.
func (tc *TLSMatcher) Handle(c net.Conn) (net.Conn, error) {
	return tc.Handler(c)
}

// NewTLSMatcher returns a *TLSMatcher configured with the provided handler.
func NewTLSMatcher(handler func(net.Conn) (net.Conn, error)) *TLSMatcher {
	return &TLSMatcher{Handler: handler}
}
