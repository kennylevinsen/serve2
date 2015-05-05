package serve2

import (
	"crypto/tls"
	"net"
)

// TLSProtoHandler handles abstraction of TLS connections, in order to feed
// them back into the protocol detectors
type TLSProtoHandler struct {
	// config stores the TLS configuration, including supported protocols and
	// certificates
	config *tls.Config
}

// Setup loads the certificates and sets up supported protocols
func (t *TLSProtoHandler) Setup(protos []string, cert, key string) error {
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

// Handle returns a connection with TLS abstracted away
func (t *TLSProtoHandler) Handle(c net.Conn) net.Conn {
	return tls.Server(c, t.config)
}

func (t *TLSProtoHandler) Check(header []byte) bool {
	return header[0] == 0x16 && header[5] == 0x01
}

func (t *TLSProtoHandler) BytesRequired() int {
	return 6
}

// NewTLSProtoHandler returns an initialized TLSProtoHandler
func NewTLSProtoHandler(protos []string, cert, key string) (*TLSProtoHandler, error) {
	h := TLSProtoHandler{}
	err := h.Setup(protos, cert, key)
	if err != nil {
		return nil, err
	}
	return &h, nil
}
