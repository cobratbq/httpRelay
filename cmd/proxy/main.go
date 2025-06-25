package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"syscall"

	"github.com/cobratbq/goutils/std/log"
	net_ "github.com/cobratbq/goutils/std/net"
	"github.com/cobratbq/goutils/std/strings"
	"github.com/cobratbq/httprelay"
	"golang.org/x/net/proxy"
)

func main() {
	listenAddr := flag.String("listen", ":8080", "Listening address and port for HTTP relay proxy.")
	blockAddrs := flag.String("block", "", "Comma-separated list of blocked host names, zone names, ip addresses and CIDR addresses.")
	blockLocal := flag.Bool("block-local", true, "Block known local addresses.")
	blocklist := flag.String("blocklist", "", "Filename referring to a hosts-formatted blocklist.")
	tunnel := flag.Bool("tunnel", false, "Tunnel-mode: only allow CONNECT-method to establish raw tunneled connections.")
	flag.Parse()
	// Prepare proxy dialer
	baseDialer := httprelay.DirectDialer()
	var dialer proxy.Dialer = &baseDialer
	if *blocklist != "" {
		log.Infoln("Loading blocklist from file:", *blocklist)
		var wrapErr error
		if dialer, wrapErr = httprelay.WrapBlocklistBlocking(dialer, *blocklist); wrapErr != nil {
			log.Errorln("Failed to load blocklist:", wrapErr.Error())
			os.Exit(1)
		}
	}
	if *blockLocal || *blockAddrs != "" {
		log.Infoln("Blocking local addresses:", *blockLocal, ", custom addresses:",
			strings.OrDefault(*blockAddrs, "<none>"))
		dialer = httprelay.WrapPerHostBlocking(dialer, *blockLocal, *blockAddrs)
	}

	// Start HTTP proxy server
	listener, listenErr := net_.ListenWithOptions(context.Background(), "tcp", *listenAddr,
		map[net_.Option]int{{Level: syscall.SOL_IP, Option: syscall.IP_FREEBIND}: 1})
	if listenErr != nil {
		log.Errorln("Failed to open local address for proxy:", listenErr.Error())
		os.Exit(1)
	}
	var handler http.Handler
	if *tunnel {
		log.Infoln("Tunnel-mode: only CONNECT is allowed.")
		handler = &httprelay.HTTPConnectHandler{Dialer: dialer, UserAgent: ""}
	} else {
		handler = &httprelay.HTTPProxyHandler{Dialer: dialer, UserAgent: ""}
	}
	server := http.Server{Handler: handler}
	log.Infoln("HTTP proxy server started on", *listenAddr)
	log.Infoln(server.Serve(listener))
}
