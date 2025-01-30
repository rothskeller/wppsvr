// Test server for the wppsvr CGI handler.
//
// This program listens on HTTP port 8200, and redirects all requests to the
// wppsvr CGI server (invoked as "./wppcgi").
package main

import (
	"net/http"
	"net/http/cgi"
)

func main() {
	handler := cgi.Handler{Path: "./wppcgi"}
	http.ListenAndServe("localhost:8200", &handler)
}
