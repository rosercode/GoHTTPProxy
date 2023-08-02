package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/go-ini/ini"
)

// Define reverse proxy mapping table
var proxyMap = make(map[string]*url.URL)

// Parse URL
func parseURL(rawurl string) (*url.URL, error) {
	parsedURL, err := url.Parse(rawurl)
	if err != nil {
		return nil, err
	}
	return parsedURL, nil
}

// Custom reverse proxy handler
type reverseProxyHandler struct{}

func (rph *reverseProxyHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Parse the requested domain
	hostParts := strings.Split(req.Host, ":")
	domain := hostParts[0]

	// Print request log
	log.Println("Received request for domain:", domain)
	log.Println("Request URL:", req.URL.String())
	log.Println("Request Method:", req.Method)

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
	// Load Configuration File
	cfg, err := ini.Load("config.ini")
	if err != nil {
		log.Fatalf("Failed to load configuration file: %v", err)
		return
	}

	// Parse Reverse Proxy Mapping Table from Configuration File
	section := cfg.Section("proxy")
	if section == nil {
		log.Fatal("Proxy section not found in configuration file")
		return
	}

	// Iterate through Configuration Entries and Parse URLs
	log.Println("Reverse proxy mapping table:")
	for _, key := range section.Keys() {
		targetURL, err := parseURL(key.String())
		// Print the reverse proxy mapping table
		log.Printf("%s => %s", key.Name(), targetURL.String())
		if err != nil {
			log.Fatalf("Failed to parse URL for key %s: %v", key, err)
		}
		proxyMap[key.Name()] = targetURL
	}

	// Read Server Configuration
	serverSection := cfg.Section("server")
	if serverSection == nil {
		log.Fatal("Server section not found in configuration file")
		return
	}

	address := serverSection.Key("address").String()
	port := serverSection.Key("port").String()

	// Create an instance of the custom reverse proxy handler
	handler := &reverseProxyHandler{}

	// Register the reverse proxy handler
	http.Handle("/", handler)

	// Start the server
	addr := address + ":" + port
	log.Printf("Reverse proxy server listening on %s...", addr)
	err = http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatal(err)
	}
}
