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
	socksAddr := flag.String("socks", "localhost:8000", "Address and port of SOCKS5 proxy server.")
	socksUsername := flag.String("username", "", "Username of SOCKS5 proxy server.")
	socksPassword := flag.String("password", "", "Password of SOCKS5 proxy server.")
	listenAddr := flag.String("listen", ":8080", "Listening address and port for HTTP relay proxy.")
	blockedAddrs := flag.String("block", "", "Comma-separated list of blocked host names, zone names, ip addresses and CIDR addresses.")
	blocklist := flag.String("blocklist", "", "Filename referring to a hosts-formatted blocklist. (e.g. from energized.pro)")
	flag.Parse()
	// Compose SOCKS auth
	var auth *proxy.Auth
	if *socksUsername != "" && *socksPassword != "" {
		auth = new(proxy.Auth)
		auth.User = *socksUsername
		auth.Password = *socksPassword
	}
	// Prepare proxy relay with target SOCKS proxy
	dialer, err := proxy.SOCKS5("tcp", *socksAddr, auth, proxy.Direct)
	if err != nil {
		log.Println("Failed to create proxy definition:", err.Error())
		return
	}
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
	log.Println("HTTP proxy relay server started on", *listenAddr, "relaying to SOCKS proxy", *socksAddr)
	log.Println(http.ListenAndServe(*listenAddr, &httprelay.HTTPProxyHandler{Dialer: dialer}))
}
