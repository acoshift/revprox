package main

import (
	"flag"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

var (
	addr   = flag.String("addr", ":8080", "Listen address and port")
	target = flag.String("target", "http://localhost", "Target URL")
)

func main() {
	flag.Parse()

	u, err := url.Parse(*target)
	if err != nil {
		log.Fatalf("parse url; %v", err)
	}

	h := httputil.NewSingleHostReverseProxy(u)
	log.Printf("start revprox on %s\n", *addr)
	if err := http.ListenAndServe(*addr, h); err != nil {
		log.Fatalf("http; %v", err)
	}
}
