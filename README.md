# serve2
A protocol detecting server

You don't like having to have to decide what port to use for a service? Maybe you're annoyed by a firewall that only allows traffic to port 80? Or even a packet inspecting one that only allows real TLS traffic on port 443, but you want to SSH through none the less?

Welcome to serve2, a protocol recognizing and stacking server/dispatcher.

serve2 allows you to serve multiple protocols on a single socket. Example handlers include proxy, HTTP, TLS (through which HTTPS is handled), ECHO and DISCARD. More can easily be added, as long as the protocol sends some data that can be recognized. The proxy handler allows you to redirect the connection to external services, such as OpenSSH or Nginx, in case you don't want or can't use a Go implementation. In most cases, proxy will be sufficient.

So, what does this mean? Well, it means that if you run serve2 for port 22, 80 and 443 (or all the ports, although I would suggest just having your firewall redirect things in that case, rather than having 65535 listening sockets), you could ask for HTTP(S) on port 22, SSH on port 80, and SSH over TLS (Meaning undetectable without a MITM attack) on port 443! You have to admit that it's kind of neat.

An easy to configure frontend will pop up in the near future.

Example:

   server := serve2.New()

   ssh := proto.NewProxy("SSH", "tcp", "localhost:8080")
   tls, err := proto.NewTLS([]string{"ssh"}, "cert.pem", "key.pem")
   if err != nil {
      panic(err)
   }

   server.AddHandlers(ssh, tls)
   l, err := net.Listen("tcp", ":8080")
   if err != nil {
      panic(err)
   }

   server.Serve(l)

To get:

    go get github.com/joushou/serve2

More info and examples at:

   https://godoc.org/github.com/Joushou/serve2
