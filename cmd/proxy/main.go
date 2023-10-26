package main

import (
	"flag"
	"log"
	"net/http"

	os_ "github.com/cobratbq/goutils/std/os"
	"github.com/cobratbq/goutils/std/strings"
	"github.com/cobratbq/httprelay"
	"golang.org/x/net/proxy"
)

func main() {
	listenAddr := flag.String("listen", ":8080", "Listening address and port for HTTP relay proxy.")
	blockAddrs := flag.String("block", "", "Comma-separated list of blocked host names, zone names, ip addresses and CIDR addresses.")
	blockLocal := flag.Bool("block-local", true, "Block known local addresses.")
	blocklist := flag.String("blocklist", "", "Filename referring to a hosts-formatted blocklist.")
	flag.Parse()
	// Prepare proxy dialer
	var dialer proxy.Dialer = proxy.Direct
	if *blocklist != "" {
		log.Println("Loading blocklist from file:", *blocklist)
		var wrapErr error
		if dialer, wrapErr = httprelay.WrapBlocklistBlocking(dialer, *blocklist); wrapErr != nil {
			os_.ExitWithError(1, "Failed to load blocklist: "+wrapErr.Error())
		}
	}
	if *blockLocal || *blockAddrs != "" {
		log.Println("Blocking local addresses:", *blockLocal, ", custom addresses:",
			strings.OrDefault(*blockAddrs, "<none>"))
		dialer = httprelay.WrapPerHostBlocking(dialer, *blockLocal, *blockAddrs)
	}
	// Start HTTP proxy server
	log.Println("HTTP proxy server started on", *listenAddr)
	log.Println(http.ListenAndServe(*listenAddr, &httprelay.HTTPProxyHandler{Dialer: dialer}))
}
