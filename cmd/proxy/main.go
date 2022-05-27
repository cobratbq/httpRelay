package main

import (
	"flag"
	"log"
	"net/http"

	osutils "github.com/cobratbq/goutils/std/os"
	"github.com/cobratbq/httprelay"
	"golang.org/x/net/proxy"
)

func main() {
	listenAddr := flag.String("listen", ":8080", "Listening address and port for HTTP relay proxy.")
	blockedAddrs := flag.String("block", "", "Comma-separated list of blocked host names, zone names, ip addresses and CIDR addresses.")
	blocklist := flag.String("blocklist", "", "Filename referring to a hosts-formatted blocklist. (e.g. from energized.pro)")
	flag.Parse()
	// Prepare proxy dialer
	var dialer proxy.Dialer = proxy.Direct
	if *blocklist != "" {
		blocklistDialer := httprelay.BlocklistDialer{
			List:   make(map[string]struct{}, 0),
			Dialer: dialer}
		if err := httprelay.LoadHostsFile(&blocklistDialer, *blocklist); err != nil {
			osutils.ExitWithError(1, "Failed to load blocklist: "+err.Error())
		}
		dialer = &blocklistDialer
	}
	if *blockedAddrs != "" {
		// Prepare dialer for blocked addresses
		perHostDialer := proxy.NewPerHost(dialer, &httprelay.NopDialer{})
		perHostDialer.AddFromString(*blockedAddrs)
		dialer = perHostDialer
	}
	// Start HTTP proxy server
	log.Println("HTTP proxy server started on", *listenAddr)
	log.Println(http.ListenAndServe(*listenAddr, &httprelay.HTTPProxyHandler{Dialer: dialer}))
}
