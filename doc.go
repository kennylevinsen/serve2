/*
Package serve2 provides a mechanism to detect and handle multiple protocols on
a single net.Conn or net.Listener.

ProtocolHandlers are defined as handlers that, given the amount of header bytes
they require, can check if a header matches the protocol, as well as handle the
connection itself afterwards.

The read bytes from the header are provided to the handler by ProxyConn, that
simulates the first few reads until the header buffer is empty, at which point
it resumes normal operation.

Any protocol where the client writes data that is unique to the protocol to the
server immediately after opening the connection can be handled by serve2. Echo
and discard do not really fit here, as they would normally put no constaints on
the client - instead, serve2's test implementations require that you write ECHO
or DISCARD to the connection to trigger those protocol handlers.

A simple echo and discard server (Write ECHO or DISCARD to the socket to select
protocol):

	l, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}

	server := serve2.New()
	echo := serve2.NewEchoProtoHandler()
	discard := serve2.NewDiscardProtoHandler()
	server.AddHandlers(echo, discard, http)
	server.Serve(l)

Or with a HTTP server (Requires a http.Handler implementation as
"httpHandler"):

	type HTTPHandler struct{}

	func (h *HTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" || r.Method == "HEAD" {
			return
		}

		fmt.Fprintf(w, "<!DOCTYPE html><html><body>Welcome!</body></html>")

	}

	httpHandler := HTTPHandler{}

	// ...

	l, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}

	server := serve2.New()

	echo := serve2.NewEchoProtoHandler()
	discard := serve2.NewDiscardProtoHandler()
	http := serve2.NewHTTPProtoHandler(&httpHandler)

	server.AddHandlers(echo, discard, http)
	server.Serve(l)


*/
package serve2
