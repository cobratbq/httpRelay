# HTTP proxy to SOCKS 5 proxy relay

Single-purpose HTTP proxy relaying requests to an existing (external) SOCKS 5 proxy server. This tool is useful in the case where the SOCKS proxy is the only proxy available, but the application that you wish to use does not support SOCKS. This little program is a welcome addition to SSH's capability of setting up a SOCKS proxy server.

*Please note*: This implementation is specifically limited to this use case only. Other use cases may be trivial to create, however they are not part of this implementation.

## Usage:

`./relay -listen :8080 -socks localhost:8000`

Start a HTTP relay proxy that listens on port 8080 of every interface and connects to a SOCKS proxy server on localhost port 8000 for relaying your requests.

# Note

Please make sure to use at least [Go 1.4.3](https://github.com/golang/go/issues/12741) or Go 1.5 when compiling this application. These versions of Go have some security-related fixes for the `net/http` package.

# References

* [Go extension library for proxy connections](http://golang.org/x/net/proxy)
* [What proxies must do](https://www.mnot.net/blog/2011/07/11/what_proxies_must_do)
