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
	socksAddr := flag.String("socks", "localhost:8000", "Address and port of SOCKS5 proxy server.")
	socksUsername := flag.String("socks-user", "", "Username for accessing the SOCKS5 proxy server.")
	socksPassword := flag.String("socks-pass", "", "Password for accessing the SOCKS5 proxy server.")
	listenAddr := flag.String("listen", ":8080", "Listening address and port for HTTP relay proxy.")
	blockAddrs := flag.String("block", "", "Comma-separated list of blocked host names, zone names, ip addresses and CIDR addresses.")
	blockLocal := flag.Bool("block-local", true, "Block known local addresses.")
	blocklist := flag.String("blocklist", "", "Filename referring to a hosts-formatted blocklist.")
	tunnel := flag.Bool("tunnel", false, "Tunnel-mode: only allow CONNECT-method to establish raw tunneled connections.")
	flag.Parse()
	// Compose SOCKS auth
	var auth *proxy.Auth
	if *socksUsername != "" && *socksPassword != "" {
		auth = new(proxy.Auth)
		auth.User = *socksUsername
		auth.Password = *socksPassword
	}
	// Prepare proxy relay with target SOCKS proxy
	baseDialer := httprelay.DirectDialer()
	dialer, err := proxy.SOCKS5("tcp", *socksAddr, auth, &baseDialer)
	if err != nil {
		log.Errorln("Failed to create proxy definition:", err.Error())
		os.Exit(1)
	}
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
	log.Infoln("HTTP proxy relay server started on", *listenAddr, "relaying to SOCKS proxy", *socksAddr)
	log.Infoln(server.Serve(listener))
}
