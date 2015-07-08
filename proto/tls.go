package proto

import (
	"crypto/tls"
	"net"
)

// TLS handles abstraction of TLS connections, in order to feed
// them back into the protocol detectors.
type TLS struct {
	// config stores the TLS configuration, including supported protocols and
	// certificates.
	config *tls.Config
}

// Setup loads the certificates and sets up supported protocols.
func (t *TLS) Setup(protos []string, cert, key string) error {
	var err error
	t.config = &tls.Config{}
	t.config.NextProtos = protos
	t.config.Certificates = make([]tls.Certificate, 1)
	t.config.Certificates[0], err = tls.LoadX509KeyPair(cert, key)
	if err != nil {
		return err
	}

	return nil
}

// Handle returns a connection with TLS abstracted away.
func (t *TLS) Handle(c net.Conn) net.Conn {
	return tls.Server(c, t.config)
}

// Check checks if the protocol is TLS
func (t *TLS) Check(header []byte) bool {
	return header[0] == 0x16 && header[5] == 0x01
}

// BytesRequired returns how many bytes are required to check for TLS
func (t *TLS) BytesRequired() int {
	return 6
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
