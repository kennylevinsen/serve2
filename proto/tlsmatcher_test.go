package proto

import (
	"crypto/tls"
	"testing"
)

type ConnectionStater struct {
	CS tls.ConnectionState
}

func (c ConnectionStater) ConnectionState() tls.ConnectionState {
	return c.CS
}

func TestTLSMatcher(t *testing.T) {
	tlsCheck := &TLSMatcher{}

	tlsCheck.ServerName = "TestName"
	tlsCheck.NegotiatedProtocol = "h2"
	tlsCheck.Version = tls.VersionTLS12
	tlsCheck.Verifications = TLSCheckServerName | TLSCheckNegotiatedProtocol | TLSCheckVersion

	tests := []struct {
		match bool
		cs    tls.ConnectionState
	}{
		{
			true,
			tls.ConnectionState{
				ServerName:         "TestName",
				NegotiatedProtocol: "h2",
				Version:            tls.VersionTLS12,
			},
		},
		{
			false,
			tls.ConnectionState{
				ServerName:         "TestNameR",
				NegotiatedProtocol: "h2",
				Version:            tls.VersionTLS12,
			},
		},
		{
			false,
			tls.ConnectionState{
				ServerName:         "TestName",
				NegotiatedProtocol: "h123",
				Version:            tls.VersionTLS12,
			},
		},
		{
			false,
			tls.ConnectionState{
				ServerName:         "TestName",
				NegotiatedProtocol: "h2",
				Version:            tls.VersionTLS11,
			},
		},
	}

	for i, test := range tests {
		c := ConnectionStater{test.cs}
		match, _ := tlsCheck.Check(nil, []interface{}{c})

		if match && !test.match {
			t.Errorf("checker matched (test %d)", i)
		} else if !match && test.match {
			t.Errorf("checker didn't match (test %d)", i)
		}
	}
}
