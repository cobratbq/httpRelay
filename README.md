# HTTP proxy to SOCKS 5 proxy relay

Single-purpose HTTP proxy relaying requests to an existing (external) SOCKS 5 proxy server. This tool is useful in the case where the SOCKS proxy is the only proxy available, but the application that you wish to use does not support SOCKS. This little program is a welcome addition to SSH's capability of setting up a SOCKS proxy server.

*Please note*: This implementation is specifically limited to this use case only. Other use cases may be trivial to create, however they are not part of this implementation.

## Usage

`./relay -listen :8080 -socks localhost:8000 -block "127.0.0.1,localhost,192.168/16"`

Start a HTTP relay proxy that listens on port 8080 of every interface and connects to a SOCKS proxy server on localhost port 8000 for relaying your requests. Block requests that attempt to access 127.0.0.1, 'localhost' or any address in the ip range 192.168.0.0-192.168.255.255.

`./proxy -listen localhost:8080 -block "127.0.0.1,localhost,192.168/16"`

Start a (tiny) generic HTTP proxy server that listens on port 8080 of 'localhost' and proxies requests directly to the internet. Block any requests to 127.0.0.1, 'localhost' or any address in the ip range 192.168.0.0-192.168.255.255.

## Building

The simplest way to build is: `make`.

This build will use the build flag `-tags netgo` to make the result independent of `gcc`. Refer to `Makefile` for details.

## References

* [Go extension library for proxy connections](http://golang.org/x/net/proxy)
* [What proxies must do](https://www.mnot.net/blog/2011/07/11/what_proxies_must_do)
