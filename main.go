package main

import (
	"crypto/tls"
	"flag"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

var (
	addr      = flag.String("addr", ":8080", "listen address and port")
	target    = flag.String("target", "http://localhost", "target upstream url")
	host      = flag.String("host", "", "override the host header sent to the upstream")
	userAgent = flag.String("user-agent", "", "override the user-agent header sent to the upstream")
	path      = flag.String("path", "", "override the request path")
	stripUri  = flag.Bool("strip-uri", false, "strip the href path")
	accessLog = flag.Bool("access-log", false, "enable access logging")
)

func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}

func main() {
	flag.Parse()

	u, err := url.Parse(*target)
	if err != nil {
		log.Fatalf("parse url; %v", err)
	}

	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	targetQuery := u.RawQuery
	director := func(req *http.Request) {
		req.URL.Scheme = u.Scheme
		req.URL.Host = u.Host
		if len(*host) > 0 {
			req.Host = *host
		}
		if targetQuery == "" || req.URL.RawQuery == "" {
			req.URL.RawQuery = targetQuery + req.URL.RawQuery
		} else {
			req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
		}
		if len(*userAgent) > 0 {
			req.Header.Set("User-Agent", *userAgent)
		} else if _, ok := req.Header["User-Agent"]; !ok {
			// explicitly disable User-Agent so it's not set to default value
			req.Header.Set("User-Agent", "")
		}
		if len(*path) > 0 {
			req.URL.Path = *path
			if *stripUri {
				log.Printf("warning: strip-uri used with path override, path will always be /")
			}
		}
		if *stripUri {
			req.URL.Path = "/"
		} else {
			req.URL.Path = singleJoiningSlash(u.Path, req.URL.Path)
		}
		if *accessLog {
			requestDump, err := httputil.DumpRequest(req, true)
			if err != nil {
				log.Fatalf("%v", err)
			}
			log.Printf(string(requestDump))
		}
	}

	h := &httputil.ReverseProxy{
		Director:  director,
		Transport: transport,
	}

	log.Printf("start revprox on %s\n", *addr)
	if err := http.ListenAndServe(*addr, h); err != nil {
		log.Fatalf("http; %v", err)
	}
}
