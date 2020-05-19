package linkio

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"time"
)

type FetchResult struct {
	// The link path that provided the result.
	Path string
	// The time the result was fetched.
	Time time.Time
	// The MIME type of the fetched doc.
	MimeType string
	// The raw document content for this fetch.
	Doc []byte
}

type Fetcher interface {
	Fetch(link string) (FetchResult, error)
}

// WikiSummaryFetcher fetches link summaries from Wikipedia.
type WikiSummaryFetcher struct {
}

func (w WikiSummaryFetcher) Fetch(link string) (FetchResult, error) {
	const url = "https://en.wikipedia.org/api/rest_v1/page/summary/"
	name := path.Base(link)
	if name == "" {
		return FetchResult{}, fmt.Errorf("no base name for link %s", link)
	}
	resp, err := http.Get(path.Join(url, name))
	if err != nil {
		return FetchResult{}, fmt.Errorf("wiki summary fetcher GET: %w", err)
	}
	defer resp.Body.Close()
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return FetchResult{}, fmt.Errorf("wiki summary fetcher body: %w", err)
	}
	resp.Header.Get()

	panic("implement me")
}
