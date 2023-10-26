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
	socksAddr := flag.String("socks", "localhost:8000", "Address and port of SOCKS5 proxy server.")
	socksUsername := flag.String("socks-user", "", "Username for accessing the SOCKS5 proxy server.")
	socksPassword := flag.String("socks-pass", "", "Password for accessing the SOCKS5 proxy server.")
	listenAddr := flag.String("listen", ":8080", "Listening address and port for HTTP relay proxy.")
	blockAddrs := flag.String("block", "", "Comma-separated list of blocked host names, zone names, ip addresses and CIDR addresses.")
	blockLocal := flag.Bool("block-local", true, "Block known local addresses.")
	blocklist := flag.String("blocklist", "", "Filename referring to a hosts-formatted blocklist.")
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
	log.Println("HTTP proxy relay server started on", *listenAddr, "relaying to SOCKS proxy", *socksAddr)
	log.Println(http.ListenAndServe(*listenAddr, &httprelay.HTTPProxyHandler{Dialer: dialer}))
}
