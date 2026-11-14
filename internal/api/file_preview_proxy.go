package api

import (
	"net/http"
	"strings"
	"time"
)

// Bound connection/header setup, not the duration of a successful media body.
// The incoming request context cancels the stream when the viewer disconnects.
var fileMediaProxyTransport = func() *http.Transport {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.ResponseHeaderTimeout = 20 * time.Second
	return transport
}()

func fileMediaProxyClient(base *http.Client, method, path string) *http.Client {
	parts := strings.Split(path, "/")
	if (method != http.MethodGet && method != http.MethodHead) || len(parts) != 5 || parts[0] != "api" || parts[1] != "servers" || parts[3] != "files" || parts[4] != "media" {
		return base
	}
	client := *base
	client.Timeout = 0
	if client.Transport == nil {
		client.Transport = fileMediaProxyTransport
	}
	return &client
}
