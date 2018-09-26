package main

import (
	"flag"
	"net/http"
	"net/url"

	log "github.com/sirupsen/logrus"
	"github.com/acoshift/revprox"
)

func main() {
	var (
		addr          = flag.String("addr", ":8080", "listen address and port")
		target        = flag.String("target", "http://localhost", "target upstream url")
		host          = flag.String("host", "", "override the host header sent to the upstream")
		userAgent     = flag.String("user-agent", "", "override the user-agent header sent to the upstream")
		path          = flag.String("path", "", "override the request path")
		extraRequest  = flag.String("extra-request", "", "extra comma-separated request headers to send to the upstream")
		extraResponse = flag.String("extra-response", "", "extra comma-separated response headers to send back to the client")
		authRealm     = flag.String("auth-realm", "Restricted", "http basic auth realm (frontend)")
		authUsername  = flag.String("auth-user", "admin", "http basic auth username (frontend)")
		authPassword  = flag.String("auth-pass", "", "http basic auth password (frontend)")
		stripURI      = flag.Bool("strip-uri", false, "strip the href path")
		hideServer    = flag.Bool("hide-server", false, "hide the upstream server in responses")
		noCache       = flag.Bool("no-cache", false, "send no-cache header in responses")
		accessLog     = flag.Bool("access-log", false, "enable access logging")
	)

	flag.Parse()

	log.Printf("revprox %s", revprox.Version)

	// target url validation
	origin, err := url.Parse(*target)
	if err != nil {
		log.Fatalf("parse url; %v", err)
	}

	proxy := &revprox.Proxy{
		Origin:        origin,
		Host:          *host,
		UserAgent:     *userAgent,
		Path:          *path,
		ExtraRequest:  *extraRequest,
		ExtraResponse: *extraResponse,
		AuthRealm:     *authRealm,
		AuthUsername:  *authUsername,
		AuthPassword:  *authPassword,
		StripURI:      *stripURI,
		HideServer:    *hideServer,
		NoCache:       *noCache,
		AccessLog:     *accessLog,
	}

	log.Printf("start revprox on %s", *addr)
	if err := http.ListenAndServe(*addr, proxy); err != nil {
		log.Fatalf("revprox; %v", err)
	}
}
