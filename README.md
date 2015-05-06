# serve2
A protocol detecting server

serve2 allows you to serve multiple protocols on a single socket. Example handlers include HTTP, TLS (through which HTTPS is handled), SSH, ECHO and DISCARD. More can easily be added, as long as the protocol sends some data that can be recognized. The proxy handler also allows you to redirect the connection to external services, such as OpenSSH or Nginx, in case you don't want or can't use a Go implementation.

To get:

    go get github.com/joushou/serve2

To test:

    go run examples/simple.go

More info at:

   https://godoc.org/github.com/Joushou/serve2
