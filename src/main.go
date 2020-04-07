package main

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

// Get env var or default
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// Serve a reverse proxy for a given url
func serveReverseProxy(target string, res http.ResponseWriter, req *http.Request) {
	// parse the url
	//target, _ := url.Parse(u)

	// create the reverse proxy
	//proxy := httputil.NewSingleHostReverseProxy(url)

	//targetQuery := target.RawQuery
	director := func(req *http.Request) {
		req.URL.Scheme = "http"    //target.Scheme
		req.URL.Host = "localhost" //target.Host
		//req.URL.Path = singleJoiningSlash(target.Path, req.URL.Path)
		// if targetQuery == "" || req.URL.RawQuery == "" {
		// 	req.URL.RawQuery = targetQuery + req.URL.RawQuery
		// } else {
		// 	req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
		// }
		// if _, ok := req.Header["User-Agent"]; !ok {
		// 	// explicitly disable User-Agent so it's not set to default value
		// 	req.Header.Set("User-Agent", "")
		// }
	}

	var dialer func() (net.Conn, error)

	if strings.HasPrefix(target, "unix:/") {
		addr := strings.TrimPrefix(target, "unix:/")
		dialer = func() (net.Conn, error) {
			//fmt.Printf("dial [unix] to %s\n", addr)
			return net.Dial("unix", addr)
			//dialer := net.Dialer{} // don't know why we need a struct to use DialContext()
			//return dialer.DialContext(ctx, "unix", addr)
		}
	} else {
		url, _ := url.Parse(target)
		addr := url.Host
		dialer = func() (net.Conn, error) {
			//fmt.Printf("dial [tcp] to %s\n", addr)
			return net.Dial("tcp", addr)
			//dialer := net.Dialer{} // don't know why we need a struct to use DialContext()
			//return dialer.DialContext(ctx, "tcp", addr)
		}
	}

	transport := &http2.Transport{
		// So http2.Transport doesn't complain the URL scheme isn't 'https'
		AllowHTTP: true,
		// Pretend we are dialing a TLS endpoint. (Note, we ignore the passed tls.Config)
		DialTLS: func(network, addr string, _ *tls.Config) (net.Conn, error) {
			return dialer()
		},
	}

	proxy := &httputil.ReverseProxy{
		Director:  director,
		Transport: transport,
	}

	// Update the headers to allow for SSL redirection
	//req.URL.Host = url.Host
	//req.URL.Scheme = url.Scheme
	//req.Header.Set("X-Forwarded-Host", req.Header.Get("Host"))
	//req.Host = url.Host

	// Note that ServeHttp is non blocking and uses a go routine under the hood
	proxy.ServeHTTP(res, req)
}

func main() {
	bindTo := getEnv("BIND_TO", "0.0.0.0:1010")
	proxyTo := getEnv("PROXY_TO", "http://localhost:5216")

	h2s := &http2.Server{}

	handler := http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		serveReverseProxy(proxyTo, res, req)
	})

	server := &http.Server{
		Addr:    bindTo,
		Handler: h2c.NewHandler(handler, h2s),
	}

	fmt.Printf("Listening on [%s] proxy to %s\n", bindTo, proxyTo)

	if strings.HasPrefix(bindTo, "unix:/") {
		unixListener, err := net.Listen("unix", strings.TrimPrefix(bindTo, "unix:/"))
		if err != nil {
			panic(err)
		}
		server.Serve(unixListener)
	} else {
		if err := server.ListenAndServe(); err != nil {
			panic(err)
		}
	}
}
