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
	blockedAddrs := flag.String("block", "", "Comma-separated list of blocked host names, zone names, ip addresses and CIDR addresses.")
	flag.Parse()
	// Prepare proxy dialer
	var dialer proxy.Dialer = proxy.Direct
	if *blockedAddrs != "" {
		// Prepare dialer for blocked addresses
		perHostDialer := proxy.NewPerHost(dialer, &httpRelay.NopDialer{})
		perHostDialer.AddFromString(*blockedAddrs)
		dialer = perHostDialer
	}
	// Start HTTP proxy server
	log.Println("HTTP proxy server started on", *listenAddr)
	log.Println(http.ListenAndServe(*listenAddr, &httpRelay.HTTPProxyHandler{Dialer: dialer}))
}
