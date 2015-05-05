# serve2
A protocol detecting server

serve2 allows you to serve multiple protocols on a single socket. Example handlers include HTTP, TLS (through which HTTPS is handled), SSH, ECHO and DISCARD. More can easily be added, as long as the protocol sends some data that can be recognized.

An example can be found under examples/full.go
