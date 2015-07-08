package proto_test

import (
	"fmt"
	"net"
	"net/http"

	"github.com/joushou/serve2"
	"github.com/joushou/serve2/proto"
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
	proxy := proto.NewProxy("SSH", "tcp", "localhost:22")

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
