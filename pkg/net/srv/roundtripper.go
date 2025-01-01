package srv

import "net/http"

// HTTPClient is an interface that models *http.Client.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// RoundTripFunc is a function that implements HTTPClient and
// http.RoundTripper.
type RoundTripFunc func(req *http.Request) (*http.Response, error)

func (H RoundTripFunc) Do(req *http.Request) (*http.Response, error)        { return H(req) }
func (H RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) { return H(req) }
