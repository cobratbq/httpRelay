package main

import (
	"flag"
	"log"
	"net/http"

	"golang.org/x/net/proxy"

	"dannyvanheumen.nl/pkg/httpRelay"
)

func main() {
	listenAddr := flag.String("listen", ":8080", "Listening address and port for HTTP relay proxy.")
	flag.Parse()
	// Start HTTP proxy server
	log.Println("HTTP proxy server started on", *listenAddr)
	log.Println(http.ListenAndServe(*listenAddr, &httpRelay.HTTPProxyHandler{Dialer: proxy.Direct}))
}
