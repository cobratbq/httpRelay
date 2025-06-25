# HTTP proxy to SOCKS 5 proxy relay

Single-purpose HTTP proxy relaying requests to an existing (external) SOCKS 5 proxy server. This tool is useful in the case where the SOCKS proxy is the only proxy available, but the application that you wish to use does not support SOCKS. This little program is a welcome addition to SSH's capability of setting up a SOCKS proxy server.

*Please note*: This implementation is specifically limited to this use case only. Other use cases may be trivial to create, however they are not part of this implementation.

## Usage

`./relay -listen :8080 -socks localhost:8000 -socks-user socksUsername -socks-pass socksPassword -block "127.0.0.1,localhost,192.168/16"`

Start a HTTP relay proxy that listens on port 8080 of every interface and connects to a SOCKS proxy server on localhost port 8000 for relaying your requests. Block requests that attempt to access 127.0.0.1, 'localhost' or any address in the ip range 192.168.0.0-192.168.255.255.

`./proxy -listen localhost:8080 -block "127.0.0.1,localhost,192.168.0.1/16"`

Start a (tiny) generic HTTP proxy server that listens on port 8080 of 'localhost' and proxies requests directly to the internet. Block any requests to 127.0.0.1, 'localhost' or any address in the ip range 192.168.0.0-192.168.255.255.

## Program arguments

The program arguments that are available to both programs.

- `-block` provide any number of network addresses/ranges to protect from access through the proxy/relay.
- `-block-local` block private network IP-ranges. (Enabled by default.)
- `-blocklist` specify a `hosts`-formatted blocklist to be loaded and used.
- `-listen` specify the address and port on which to listen for incoming proxy connections.

The following program arguments are applicable to `relay` only.

- `-socks` the SOCKS proxy to which to forward http proxy requests.
- `-socks-user` the username of SOCKS5 proxy server.
- `-socks-pass` the password of SOCKS5 proxy server.

## Building

The simplest way to build is: `make`.

This build will use the build flag `-tags netgo` to make the result independent of `gcc`. Refer to `Makefile` for details.

## Changelog

- _2023-08-15_ Command-line flags to provide username/password authentication for SOCKS5 proxy (relay) by [developbranch-cn](<https://github.com/developbranch-cn>).
- _2020-02-04_ Added support for loading in blocklists that are checked as part of the proxying process.  
  The program argument `-blocklist <filename>` allows loading hostname blocklists formatted as the OS `hosts` file. Blocklists in these formats can be downloaded from various places, such as [NoTracking][github-notracking] and [EnergizedPro][github-energizedpro].
- _2019-12-22_ Added support for Go modules.
- _way back_ Support for http proxy/relay, with `-block` parameter to protect local network and/or specific addresses/networks from being accessed.

## References

- [Go extension library for proxy connections](http://golang.org/x/net/proxy)
- [What proxies must do](https://www.mnot.net/blog/2011/07/11/what_proxies_must_do)
- [NoTracking blocklist][github-notracking]
- [EnergizedPro][github-energizedpro]

[github-notracking]: https://github.com/EnergizedProtection/block "NoTracking blocklist"
[github-energizedpro]: https://github.com/EnergizedProtection/block "Energized Protection"
