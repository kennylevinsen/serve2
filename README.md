# serve2 [![Join the chat at https://gitter.im/joushou/serve2](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/joushou/serve2?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge) [![GoDoc](https://godoc.org/github.com/joushou/serve2?status.svg)](http://godoc.org/github.com/joushou/serve2) [![Build Status](https://travis-ci.org/joushou/serve2.svg?branch=master)](https://travis-ci.org/joushou/serve2)

A protocol detecting server library

serve2 accepts a connection, and runs it through the active ProtocolHandlers, reading the smallest amount of data necessary to identify the protocol. ProtocolHandlers do not need to be certain about how much data they need up front - they can ask for more as needed. A default ProtocolHandler can be used when the server fails to identify the protocol, or one can rely on the default behaviour that closes the socket.

ProtocolHandlers can implement transports themselves, by returning a new net.Conn that will be fed through the protocol detection mechanism again, allowing for transparent support of transports. HTTPS (for HTTP/1.1) is implemented in this manner simply by adding the TLS transport independent of HTTP, leaving HTTP completely unaware of TLS. This also allows for things like SSH over TLS, or any other protocol for that matter.

serve2 comes with a set of ProtocolHandlers ready for consumption, in the form of TLS, HTTP (consuming a http.Handler), ECHO and DISCARD. There is also a very convenient Proxy ProtocolHandler, which instead of managing the protocol itself simply checks for a predefined set of bytes, and when matched, dials a configured service, leaving the actual protocol up to someone else. This is useful for things like SSH.

Ensuring that the read bytes are fed back in is done by ProxyConn, a net.Conn-implementing type with a buffered Read.

# Why?
Well, I always kind of wanted to make something that could understand *everything*. I get those kinds of ideas occasionally. At one point I remembered that idea, and hving gotten caught by the Go fever, I thought I'd try it out in Go, which proved to be very suitable for the idea.

# Uses
Apart from how cool it is to be able to serve everything on any port, it also allows flexibility when firewall rules are present. Ever had to run SSH on port 80 to get through a firewall? You know, annoying corporate networks, or maybe difficulties with annoying ISP's and your home server. Well, now you can still have a nice web server on port 80 and still serve SSH. Or maybe you had a packet inspecting firewall that didn't think that was a good idea? Use openssl's s_client to open a TLS transport to port 443, and SSH to that then!

# Limitations
serve2 cannot detect /all/ protocols. It's not really possible to detect DISCARD or ECHO, for example, as the client does not send any recognizable array of bytes before expecting the server to reply. Instead, they require that you send "ECHO" or "DISCARD" as the first message you want echoed or discarded.

In order to be able to detect a protocol, the client will have to send something either immediately on connect, or at the latest before expecting the server to reply/do anything. What it sends must furthermore either be a static magic, or a dynamic message within such boundaries that a pattern can be validated programmatically. Static "magics" can be seen in the form of SSH that starts out by sending "SSH..." (... being a longer version string), and more dynamic ones involve TLS that does not send a magic, but always starts by sending a ClientHello, of which the first byte is 0x16 (to inform that this is a handshake), followed by major/minor version numbers and the handshake type (which for the first message is always ClientHello, 0x01). With this information, one can verify the message/handshake type and major/minor version number ranges, and establish with a decent probability that this is indeed a TLS ClientHello handshake.

While ProtocolHandlers can ask for as much data as they can dream of, and can incrementally increase how much data they need (in case dynamic patterns also have dynamic lengths, for example), but it is suggested that the detection amount is kept as small as possible while also maintaining good probability of the protocol. SSH can be detected with extremely high probability by reading 3 bytes, which is a nice and small amount, and HTTP can be detected by looking at the HTTP method (and increasing the read amount if longer method names must be tested).

# Performance
This depends heavily on both the quantity of registered handlers, and the individual handlers themselves. Every time more data is has to be read, serve2 will read the smallest amount that was requested by the remaining handlers (handlers that need len(requested) >= len(read)), resulting in a few unnecessary calls to handlers. As long as the handlers are simple checks, and it doesn't end up running through this iteration several hundred times, then the overhead should not be too bad. A future optimization to ensure that handlers are only called when len(read) >= len(requested), in order to lower the overhead.

# What does the name mean?
Nothing. I called the toy project "serve", and when making a new version I had to use a new folder name, and so "serve2" was born.

# Installation and documentation
To get:

      go get github.com/joushou/serve2

More info and examples at:
* https://godoc.org/github.com/joushou/serve2 (The core itself)
* https://godoc.org/github.com/joushou/serve2/proto (The bundled ProtocolHandlers)
* https://godoc.org/github.com/joushou/serve2/utils (Utilities like ProxyConn and ChannelListener)

