package main

import (
	"net/http"
)

func buildRoutes() *http.ServeMux {
	heapAPIProxy := NewHeapAPIProxy()
	heapCDNProxy := NewHeapCDNProxy(http.DefaultTransport)

	mux := http.NewServeMux()
	mux.Handle("GET /_/heap/js/", heapCDNProxy)
	mux.Handle("GET /_/heap/", heapAPIProxy)
	return mux
}
