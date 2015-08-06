package proto_test

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"

	"github.com/joushou/serve2"
	"github.com/joushou/serve2/proto"
	"github.com/joushou/serve2/utils"
)

func ExampleNewTLS() {
	server := serve2.New()

	// We just set NextProto to "echo". Cert and key required!
	tls, err := proto.NewTLS([]string{"echo"}, "cert.pem", "key.pem")
	if err != nil {
		panic(err)
	}

	// We add the echo handler too
	server.AddHandlers(tls, proto.NewEcho())
	l, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}

	server.Serve(l)
}

func ExampleNewProxy() {
	server := serve2.New()

	// SSH just so happens to send "SSH" and version number as the first thing
	proxy := proto.NewProxy([]byte("SSH"), "tcp", "localhost:22")

	server.AddHandlers(proxy)
	l, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}

	server.Serve(l)

}

func ExampleNewEcho() {
	server := serve2.New()

	// If we send "ECHO" first, then everything is echoed afterwards
	server.AddHandlers(proto.NewEcho())
	l, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}

	server.Serve(l)
}

func ExampleNewDiscard() {
	server := serve2.New()

	// If we send "DISCARD" first, then everything is discarded afterwards
	server.AddHandlers(proto.NewDiscard())
	l, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}

	server.Serve(l)
}

type HTTPHandler struct{}

func (h *HTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" || r.Method == "HEAD" {
		return
	}

	fmt.Fprintf(w, "<!DOCTYPE html><html><head></head><body>Welcome to %s</body></html>", r.URL.Path)
}

func ExampleNewHTTP() {
	server := serve2.New()

	// Insert your http.Handler here
	http := proto.NewHTTP(&HTTPHandler{})

	server.AddHandlers(http)
	l, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}

	server.Serve(l)
}

func ExampleNewMultiProxy() {
	server := serve2.New()

	httpMethods := [][]byte{
		[]byte("GET"),
		[]byte("POST"),
	}

	mp := proto.NewMultiProxy(httpMethods, "tcp", "localhost:80")

	server.AddHandlers(mp)
	l, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}

	server.Serve(l)
}

func ExampleNewListenProxy() {
	server := serve2.New()

	checker := func(header []byte, _ []interface{}) (match bool, required int) {
		if len(header) < 3 {
			return false, 3
		}

		if bytes.Compare(header[:3], []byte("GET")) == 0 {
			return true, 0
		}
		return false, 0
	}

	lp := proto.NewListenProxy(checker, 10)
	lp.Description = "HTTP"

	go http.Serve(lp.Listener(), &HTTPHandler{})

	server.AddHandlers(lp)
	l, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}

	server.Serve(l)
}

func ExampleSimpleMatcher() {
	server := serve2.New()

	handler := func(c net.Conn) (net.Conn, error) {
		return nil, utils.DialAndProxy(c, "tcp", "localhost:80")
	}

	sm := proto.NewSimpleMatcher(proto.HTTPMethods, handler)
	sm.Description = "HTTP"

	server.AddHandlers(sm)
	l, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}

	server.Serve(l)
}

func ExampleTLSMatcher() {
	server := serve2.New()

	handler := func(c net.Conn) (net.Conn, error) {
		return nil, utils.DialAndProxyTLS(c, "tcp", "http2.golang.org:443", &tls.Config{
			NextProtos: []string{"h2", "h2-14"},
			ServerName: "http2.golang.org",
		})
	}

	tls, err := proto.NewTLS([]string{"h2", "h2-14"}, "cert.pem", "key.pem")
	if err != nil {
		panic(err)
	}
	server.AddHandlers(tls)

	tm := proto.NewTLSMatcher(handler)
	tm.NegotiatedProtocols = []string{"h2", "h2-14"}
	tm.Checks = proto.TLSCheckNegotiatedProtocol
	tm.Description = "HTTP2"

	server.AddHandlers(tm)
	l, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}

	server.Serve(l)
}
