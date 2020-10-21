package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

// HeapForwarder forwards all requests to Heap.
type HeapForwarder struct {
	client *http.Client
}

func NewHeapForwarder() *HeapForwarder {
	return &HeapForwarder{
		client: &http.Client{},
	}
}

func (hf *HeapForwarder) ForwardHandler(w http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Printf("error: read body: %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	req.Body = ioutil.NopCloser(bytes.NewReader(body))
	url := "https://heapanalytics.com" + req.RequestURI
	proxyReq, err := http.NewRequest(req.Method, url, bytes.NewReader(body))
	if err != nil {
		log.Printf("error: new proxy request: %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	proxyReq.Header = req.Header

	resp, err := hf.client.Do(proxyReq)
	if err != nil {
		log.Printf("error: do proxy request: %s", err.Error())
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	if err := resp.Body.Close(); err != nil {
		log.Printf("error: close response body: %s", err.Error())
	}
}

func main() {
	log.Print("starting track server")
	hf := NewHeapForwarder()

	http.HandleFunc("/", hf.ForwardHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("listening on port %s", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}
