package main

import (
	"flag"
	"log"
	"net/http"

	"golang.org/x/net/proxy"
)

func main() {
	socksAddr := flag.String("socks", "localhost:8000", "Address and port of SOCKS5 proxy server.")
	listenAddr := flag.String("listen", ":8080", "Listening address and port for HTTP relay proxy.")
	flag.Parse()
	// Prepare proxy relay with target SOCKS proxy
	dialer, err := proxy.SOCKS5("tcp", *socksAddr, nil, proxy.Direct)
	if err != nil {
		log.Println("error creating proxy definition:", err.Error())
		return
	}
	// Start HTTP proxy server
	log.Println("HTTP proxy relay server started on", *listenAddr, "relaying to", *socksAddr)
	log.Println(http.ListenAndServe(*listenAddr, &HTTPProxyHandler{Dialer: dialer}))
}
