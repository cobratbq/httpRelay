package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"golang.org/x/net/proxy"

	"github.com/cobratbq/httpRelay"
)

func main() {
	listenAddr := flag.String("listen", ":8080", "Listening address and port for HTTP relay proxy.")
	blockedAddrs := flag.String("block", "", "Comma-separated list of blocked host names, zone names, ip addresses and CIDR addresses.")
	blocklist := flag.String("blocklist", "", "Filename referring to a hosts-formatted blocklist. (e.g. from energized.pro)")
	flag.Parse()
	// Prepare proxy dialer
	var dialer proxy.Dialer = proxy.Direct
	if *blocklist != "" {
		listfile, err := os.Open(*blocklist)
		if err != nil {
			panic("Failed to access blocklist: " + *blocklist)
		}
		blocklistDialer := httpRelay.BlocklistDialer{List: make(map[string]struct{}, 0), Dialer: dialer}
		if err := blocklistDialer.Load(listfile); err != nil {
			panic("Failed to load content into blocklist dialer: " + err.Error())
		}
		dialer = &blocklistDialer
	}
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
