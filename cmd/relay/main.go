package main

import (
	"flag"
	"log"
	"net/http"

	"dannyvanheumen.nl/pkg/httpRelay"

	"golang.org/x/net/proxy"
)

func main() {
	socksAddr := flag.String("socks", "localhost:8000", "Address and port of SOCKS5 proxy server.")
	listenAddr := flag.String("listen", ":8080", "Listening address and port for HTTP relay proxy.")
	flag.Parse()
	// Prepare proxy relay with target SOCKS proxy
	dialer, err := proxy.SOCKS5("tcp", *socksAddr, nil, proxy.Direct)
	if err != nil {
		log.Println("Failed to create proxy definition:", err.Error())
		return
	}
	// Start HTTP proxy server
	log.Println("HTTP proxy relay server started on", *listenAddr, "relaying to SOCKS proxy", *socksAddr)
	log.Println(http.ListenAndServe(*listenAddr, &httpRelay.HTTPProxyHandler{Dialer: dialer}))
}
