package proto

import (
	"crypto/tls"
	"net"
)

// TLS field constants
const (
	TLSMajor        = 0x03
	TLSHighestMinor = 0x04 // Bump when new releases are made available
	TLSHandshake    = 0x16
	TLSClientHello  = 0x01
)

// TLS handles abstraction of TLS connections, in order to feed
// them back into the protocol detectors.
type TLS struct {
	// config stores the TLS configuration, including supported protocols and
	// certificates.
	config *tls.Config
}

func (TLS) String() string {
	return "TLS"
}

// Setup loads the certificates and sets up supported protocols.
func (t *TLS) Setup(protos []string, cert, key string) error {
	t.config = &tls.Config{}
	t.config.NextProtos = protos
	t.config.Certificates = make([]tls.Certificate, 1)

	var err error
	t.config.Certificates[0], err = tls.LoadX509KeyPair(cert, key)

	return err
}

// Handle returns a connection with TLS abstracted away.
func (t *TLS) Handle(c net.Conn) net.Conn {
	return tls.Server(c, t.config)
}

// Check checks if the protocol is TLS
func (t *TLS) Check(header []byte) (bool, int) {
	if len(header) < 6 {
		return false, 0
	}

	return header[0] == TLSHandshake &&
		header[1] == TLSMajor &&
		header[2] <= TLSHighestMinor &&
		header[5] == TLSClientHello, 0
}

// NewTLS returns an initialized TLS.
func NewTLS(protos []string, cert, key string) (*TLS, error) {
	h := TLS{}
	err := h.Setup(protos, cert, key)
	if err != nil {
		return nil, err
	}
	return &h, nil
}
