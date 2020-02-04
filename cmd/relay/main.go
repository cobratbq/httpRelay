package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/cobratbq/httpRelay"
	"golang.org/x/net/proxy"
)

func main() {
	socksAddr := flag.String("socks", "localhost:8000", "Address and port of SOCKS5 proxy server.")
	listenAddr := flag.String("listen", ":8080", "Listening address and port for HTTP relay proxy.")
	blockedAddrs := flag.String("block", "", "Comma-separated list of blocked host names, zone names, ip addresses and CIDR addresses.")
	blocklist := flag.String("blocklist", "", "Filename referring to a hosts-formatted blocklist. (e.g. from energized.pro)")
	flag.Parse()
	// Prepare proxy relay with target SOCKS proxy
	dialer, err := proxy.SOCKS5("tcp", *socksAddr, nil, proxy.Direct)
	if err != nil {
		log.Println("Failed to create proxy definition:", err.Error())
		return
	}
	if *blocklist != "" {
		blocklistDialer := httpRelay.BlocklistDialer{
			List:   make(map[string]struct{}, 0),
			Dialer: dialer}
		if err := httpRelay.LoadHostsFile(&blocklistDialer, *blocklist); err != nil {
			os.Stderr.WriteString("Failed to load blocklist: " + err.Error() + "\n")
			os.Exit(1)
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
	log.Println("HTTP proxy relay server started on", *listenAddr, "relaying to SOCKS proxy", *socksAddr)
	log.Println(http.ListenAndServe(*listenAddr, &httpRelay.HTTPProxyHandler{Dialer: dialer}))
}
