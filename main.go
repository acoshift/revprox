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

const (
	version string = "1.1.0"
)

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
	serverString  = "revprox/" + version

	origin *url.URL
	err    error

	transport = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	proxy = &httputil.ReverseProxy{
		Transport:      transport,
		ModifyResponse: modifyResponse,
	}
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

func modifyResponse(resp *http.Response) error {
	if *hideServer {
		resp.Header.Set("Server", serverString)
	} else {
		resp.Header.Add("Server", serverString)
	}
	if *noCache {
		resp.Header.Set("Cache-Control", "no-cache")
	}
	if len(*extraResponse) > 0 {
		for _, h := range strings.Split(*extraResponse, ",") {
			s := strings.Split(h, ":")
			resp.Header.Set(s[0], s[1])
		}
	}

	return nil
}

func proxyHandler(w http.ResponseWriter, r *http.Request) {
	if len(*authRealm) > 1 && len(*authUsername) > 1 && len(*authPassword) > 1 {
		userName, password, _ := r.BasicAuth()

		if userName != *authUsername || password != *authPassword {
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
	}

	targetQuery := origin.RawQuery
	director := func(req *http.Request) {
		req.URL.Scheme = origin.Scheme
		req.URL.Host = origin.Host
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
			if *stripURI {
				log.Printf("warning: strip-uri used with path override, path will always be /")
			}
		}
		if *stripURI {
			req.URL.Path = "/"
		} else {
			req.URL.Path = singleJoiningSlash(origin.Path, req.URL.Path)
		}
		if len(*extraRequest) > 0 {
			for _, h := range strings.Split(*extraRequest, ",") {
				s := strings.Split(h, ":")
				req.Header.Set(s[0], s[1])
			}
		}
		if *accessLog {
			requestDump, err := httputil.DumpRequest(req, true)
			if err != nil {
				log.Fatalf("%v", err)
			}
			log.Printf(string(requestDump))
		}
	}
	proxy.Director = director
	proxy.ServeHTTP(w, r)
}

func main() {
	flag.Parse()

	// target url validation
	origin, err = url.Parse(*target)
	if err != nil {
		log.Fatalf("parse url; %v", err)
	}

	http.HandleFunc("/", proxyHandler)
	log.Printf("start revprox on %s\n", *addr)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatalf("http; %v", err)
	}
}
