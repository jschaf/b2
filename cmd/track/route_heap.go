package main

import (
	"bytes"
	"fmt"
	"github.com/jschaf/jsc/pkg/errs"
	"github.com/jschaf/jsc/pkg/net/srv"
	"golang.org/x/sync/singleflight"
	"io"
	"net/http"
	"net/http/httputil"
	"strings"
	"sync"
	"time"
)

type HeapAPIProxy struct {
	rp *httputil.ReverseProxy
}

// NewHeapAPIProxy returns an HTTP reverse proxy that forwards requests
// to heapanalytics.com. Useful so that clients don't need to connect to another
// domain and to avoid being blocked by adblock.
func NewHeapAPIProxy() HeapAPIProxy {
	rp := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL.Scheme = "https"
			req.URL.Host = "heapanalytics.com"
			req.URL.Path = strings.TrimPrefix(req.URL.Path, "/_/heap")
			req.URL.RawPath = strings.TrimPrefix(req.URL.Path, "/_/heap")
		},
	}
	return HeapAPIProxy{rp: rp}
}

func (p HeapAPIProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.rp.ServeHTTP(w, r)
}

type HeapCDNProxy struct {
	cacheResp     http.Response
	rp            *httputil.ReverseProxy
	lastCacheTime time.Time
	cacheBody     []byte
	maxAge        time.Duration
	mu            sync.Mutex
}

// NewHeapCDNProxy returns an HTTP reverse proxy that forwards requests
// to cdn.heapanalytics.com. Useful so that clients don't need to connect to
// another domain, to avoid being blocked by adblock, and to cache the heap
// JS response for longer.
func NewHeapCDNProxy(transport http.RoundTripper) *HeapCDNProxy {
	p := &HeapCDNProxy{
		maxAge: 10 * time.Minute,
	}
	single := &singleflight.Group{}
	p.rp = &httputil.ReverseProxy{
		Director: func(r *http.Request) {
			r.URL.Scheme = "https"
			r.URL.Host = "cdn.heapanalytics.com"
			r.URL.Path = strings.TrimPrefix(r.URL.Path, "/_/heap")
			r.URL.RawPath = strings.TrimPrefix(r.URL.Path, "/_/heap")
		},
		Transport: srv.RoundTripFunc(func(r *http.Request) (*http.Response, error) {
			getHeap := func() (*http.Response, error) {
				anyResp, err, _ := single.Do("get_heap", func() (interface{}, error) {
					r.Host = "cdn.heapanalytics.com"
					resp, err := transport.RoundTrip(r)
					if err != nil {
						return resp, err
					}
					err = p.setResponse(resp)
					if err != nil {
						return resp, err
					}
					// Increase cache duration. By default, the Heap CDN caches JavaScript
					// for 10 minutes.
					resp.Header.Set("Cache-Control", fmt.Sprintf("public, max-age=%d", 6*time.Hour/time.Second))
					return resp, nil
				})
				return anyResp.(*http.Response), err
			}
			resp, ok := p.newCachedResponse()
			switch {
			case resp == nil:
				// No cached response.
				return getHeap()
			case ok:
				// Cached response still valid.
				return resp, nil
			default:
				// Return a stale response and fetch a new version
				//asynchronously.
				go func() { _, _ = getHeap() }()
				return resp, nil
			}
		}),
	}
	return p
}

func (p *HeapCDNProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.rp.ServeHTTP(w, r)
}

func (p *HeapCDNProxy) newCachedResponse() (*http.Response, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.cacheBody) == 0 {
		return nil, false
	}
	resp := p.cacheResp
	resp.Body = io.NopCloser(bytes.NewReader(p.cacheBody))
	isValid := time.Now().Before(p.lastCacheTime.Add(p.maxAge))
	return &resp, isValid
}

func (p *HeapCDNProxy) setResponse(resp *http.Response) (mErr error) {
	newResp := http.Response{
		Status:           resp.Status,
		StatusCode:       resp.StatusCode,
		Proto:            resp.Proto,
		ProtoMajor:       resp.ProtoMajor,
		ProtoMinor:       resp.ProtoMinor,
		Header:           resp.Header,
		Body:             nil,
		ContentLength:    resp.ContentLength,
		TransferEncoding: resp.TransferEncoding,
		Close:            resp.Close,
		Uncompressed:     resp.Uncompressed,
		Trailer:          resp.Trailer,
		Request:          nil,
		TLS:              nil,
	}
	origBody := resp.Body
	bodyBytes, err := io.ReadAll(origBody)
	if err != nil {
		return err
	}
	defer errs.Capture(&mErr, origBody.Close, "close body")
	resp.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	p.mu.Lock()
	p.lastCacheTime = time.Now()
	p.cacheResp = newResp
	p.cacheBody = bodyBytes
	p.mu.Unlock()
	return nil
}
