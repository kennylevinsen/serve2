package main

import (
	"fmt"
	"github.com/joushou/serve2"
	"net"
	"net/http"
)

//
// Accepts HTTP, ECHO and DISCARD. HTTP can be tested with a browser or
// curl/wget.
//
// To test ECHO/DISCARD
//   nc localhost 8080
//
// ... And then write "ECHO" or "DISCARD", followed by return, and what you
// want echooed or discarded

type HTTPHandler struct{}

func (h *HTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" || r.Method == "HEAD" {
		return
	}

	fmt.Fprintf(w, "<!DOCTYPE html><html><head></head><body>Welcome to %s</body></html>", r.URL.Path)

}

func main() {

	l, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}

	server := serve2.New()

	// See the HTTPHandler above
	http := serve2.NewHTTPProtoHandler(&HTTPHandler{})
	echo := serve2.NewEchoProtoHandler()
	discard := serve2.NewDiscardProtoHandler()

	server.AddHandlers(echo, discard, http)
	server.Serve(l)
}
