package main

import (
	"github.com/joushou/serve2"
	"net"
)

//
// To test:
//   nc localhost 8080
//
// ... And then write "ECHO" or "DISCARD", followed by return, and what you
// want echooed or discarded

func main() {

	l, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}

	server := serve2.New()

	// These two are silly, and requires that you write "ECHO" or "DISCARD" when
	// the connection is opened to recognize the protocol, as neither of these
	// actually have any initial request or handshake.
	echo := serve2.NewEchoProtoHandler()
	discard := serve2.NewDiscardProtoHandler()

	server.AddHandlers(echo, discard)
	server.Serve(l)
}
