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
// tls.ConnectionState fields described by Checks. If no verifications are
// enabled, TLSMatcher will simply match the presence of a TLS transport.
type TLSMatcher struct {
	ServerNames                []string
	NegotiatedProtocols        []string
	NegotiatedProtocolIsMutual bool
	PeerCertificates           []*x509.Certificate
	CipherSuites               []uint16
	Versions                   []uint16
	Checks                     TLSMatcherChecks
	Handler                    func(net.Conn) (net.Conn, error)
	Description                string
}

type connectionStater interface {
	ConnectionState() tls.ConnectionState
}

func (tc *TLSMatcher) String() string {
	return tc.Description
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

	if tc.Checks.IsSet(TLSCheckServerName) {
		for _, sn := range tc.ServerNames {
			if sn == cs.ServerName {
				goto serverNameOK
			}
		}
		return false, 0
	serverNameOK:
	}

	if tc.Checks.IsSet(TLSCheckNegotiatedProtocol) {
		for _, np := range tc.NegotiatedProtocols {
			if np == cs.NegotiatedProtocol {
				goto protocolOK
			}
		}
		return false, 0
	protocolOK:
	}

	if tc.Checks.IsSet(TLSCheckNegotiatedProtocolIsMutual) &&
		tc.NegotiatedProtocolIsMutual != cs.NegotiatedProtocolIsMutual {
		return false, 0
	}

	if tc.Checks.IsSet(TLSCheckClientCertificate) {
		// TODO: Is this how you do check the clients certificate? :/

		for _, cert := range tc.PeerCertificates {
			if cs.PeerCertificates[0].Equal(cert) {
				goto certOk
			}
		}
		return false, 0
	certOk:
	}

	if tc.Checks.IsSet(TLSCheckCipherSuite) {
		for _, cipher := range tc.CipherSuites {
			if cipher == cs.CipherSuite {
				goto cipherOK
			}
		}
		return false, 0
	cipherOK:
	}

	if tc.Checks.IsSet(TLSCheckVersion) {
		for _, version := range tc.Versions {
			if version == cs.Version {
				goto versionOK
			}
		}
		return false, 0
	versionOK:
	}

	return true, 0
}

// Handle simply calls the provided handler.
func (tc *TLSMatcher) Handle(c net.Conn) (net.Conn, error) {
	return tc.Handler(c)
}

// NewTLSMatcher returns a *TLSMatcher configured with the provided handler.
func NewTLSMatcher(handler func(net.Conn) (net.Conn, error)) *TLSMatcher {
	return &TLSMatcher{
		Handler:     handler,
		Description: "TLSMatcher",
	}
}
