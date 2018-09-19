package revprox

import (
	"crypto/subtle"
	"crypto/tls"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
)

// Proxy is the revprox proxy
type Proxy struct {
	prox httputil.ReverseProxy
	once sync.Once

	Origin        *url.URL
	Host          string
	UserAgent     string
	Path          string
	ExtraRequest  string
	ExtraResponse string
	AuthRealm     string
	AuthUsername  string
	AuthPassword  string
	StripURI      bool
	HideServer    bool
	NoCache       bool
	AccessLog     bool
}

func (p *Proxy) init() {
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	p.prox = httputil.ReverseProxy{
		Transport: transport,
	}

	targetQuery := p.Origin.RawQuery

	p.prox.Director = func(req *http.Request) {
		req.URL.Scheme = p.Origin.Scheme
		req.URL.Host = p.Origin.Host
		if len(p.Host) > 0 {
			req.Host = p.Host
		}
		if targetQuery == "" || req.URL.RawQuery == "" {
			req.URL.RawQuery = targetQuery + req.URL.RawQuery
		} else {
			req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
		}
		if len(p.UserAgent) > 0 {
			req.Header.Set("User-Agent", p.UserAgent)
		} else if _, ok := req.Header["User-Agent"]; !ok {
			// explicitly disable User-Agent so it's not set to default value
			req.Header.Set("User-Agent", "")
		}
		if len(p.Path) > 0 {
			req.URL.Path = p.Path
			if p.StripURI {
				log.Printf("warning: strip-uri used with path override, path will always be /")
			}
		}
		if p.StripURI {
			req.URL.Path = "/"
		} else {
			req.URL.Path = singleJoiningSlash(p.Origin.Path, req.URL.Path)
		}
		if len(p.ExtraRequest) > 0 {
			for _, h := range strings.Split(p.ExtraRequest, ",") {
				s := strings.Split(h, ":")
				req.Header.Set(s[0], s[1])
			}
		}
		if p.AccessLog {
			requestDump, err := httputil.DumpRequest(req, true)
			if err != nil {
				log.Fatalf("%v", err)
			}
			log.Printf(string(requestDump))
		}
	}

	p.prox.ModifyResponse = func(resp *http.Response) error {
		if p.HideServer {
			resp.Header.Set("Server", serverString)
		} else {
			resp.Header.Add("Server", serverString)
		}
		if p.NoCache {
			resp.Header.Set("Cache-Control", "no-cache")
		}
		if len(p.ExtraResponse) > 0 {
			for _, h := range strings.Split(p.ExtraResponse, ",") {
				s := strings.Split(h, ":")
				resp.Header.Set(s[0], s[1])
			}
		}

		return nil
	}
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.once.Do(p.init)

	if p.AuthRealm != "" && p.AuthUsername != "" && p.AuthPassword != "" {
		username, password, _ := r.BasicAuth()

		if username != p.AuthUsername || subtle.ConstantTimeCompare([]byte(password), []byte(p.AuthPassword)) != 1 {
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
	}

	p.prox.ServeHTTP(w, r)
}

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
