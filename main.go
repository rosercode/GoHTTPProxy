package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

// Define reverse proxy mapping table
var proxyMap = map[string]*url.URL{
	"test1.example.com": parseURL("http://localhost:8000"), // Forward test1.example.com to local HTTP server (port 8000)
	"test2.example.com": parseURL("http://localhost:9000"), // Forward test2.example.com to local HTTP server (port 9000)
}

// Parse URL
func parseURL(rawurl string) *url.URL {
	parsedURL, err := url.Parse(rawurl)
	if err != nil {
		log.Fatalf("Error parsing URL: %s", err)
	}
	return parsedURL
}

// Custom reverse proxy handler
type reverseProxyHandler struct{}

func (rph *reverseProxyHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Parse the requested domain
	hostParts := strings.Split(req.Host, ":")
	domain := hostParts[0]

	// Look up the corresponding target URL in the proxy mapping table
	targetURL, ok := proxyMap[domain]
	if !ok {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	// Create a reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	// Modify the host information in the request header
	req.URL.Host = targetURL.Host
	req.URL.Scheme = targetURL.Scheme
	req.Host = targetURL.Host

	// Set custom reverse proxy handler
	proxy.Director = func(req *http.Request) {
		req.Header.Set("X-Forwarded-Host", req.Header.Get("Host"))
	}

	// Process the request with reverse proxy
	proxy.ServeHTTP(w, req)
}

func main() {
	// Print the address and port that the server is listening on
	addr := ":18080"
	log.Printf("Reverse proxy server listening on %s...", addr)

	// Print the reverse proxy mapping table
	log.Println("Reverse proxy mapping table:")
	for domain, targetURL := range proxyMap {
		log.Printf("%s => %s", domain, targetURL.String())
	}

	// Create an instance of the custom reverse proxy handler
	handler := &reverseProxyHandler{}

	// Register the reverse proxy handler
	http.Handle("/", handler)

	// Start the server
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatal(err)
	}
}
