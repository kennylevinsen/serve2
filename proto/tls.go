package proto

import (
	"crypto/tls"
	"net"

	"github.com/joushou/serve2/utils"
)

// TLS field constants
const (
	TLSMajor        = tls.VersionTLS12 >> 8
	TLSHighestMinor = tls.VersionTLS12 & 0xFF // Bump when new releases are made available
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

// Handle returns a connection with TLS abstracted away. Adds the tls.Conn for
// the connection as a hint.
func (t *TLS) Handle(c net.Conn) (net.Conn, error) {
	s := tls.Server(c, t.config)
	hints := append(utils.GetHints(c), s)
	return utils.NewHintConn(s, hints), nil
}

// Check checks if the protocol is TLS
func (t *TLS) Check(header []byte, _ []interface{}) (bool, int) {
	if len(header) < 6 {
		// We can try to check the handhake or the version
		if len(header) >= 2 && header[1] != TLSMajor {
			return false, 0
		}
		if len(header) >= 1 && header[0] != TLSHandshake {
			return false, 0
		}
		return false, 6
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
