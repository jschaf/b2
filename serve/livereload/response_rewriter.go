package livereload

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"
	"unicode"
)

type writerState int

const (
	stateUnstarted writerState = iota
	stateHTML
	stateNotText
	stateAlreadyInjected
)

type responseRewriter struct {
	buf        bytes.Buffer
	state      writerState
	nextWriter *countingResponseWriter
	// The tag to insert into HTML.
	scriptTag string
	// Any suffix of a write that might have been the start of a </head> tag.
	// For example, we might store "</he" in remaining.
	remaining  []byte
	contentLen int
}

func newResponseRewriter(replacement string, next http.ResponseWriter) *responseRewriter {
	return &responseRewriter{
		nextWriter: &countingResponseWriter{ResponseWriter: next},
		scriptTag:  replacement,
	}
}

func (rr *responseRewriter) Header() http.Header {
	return rr.nextWriter.Header()
}

func (rr *responseRewriter) WriteHeader(statusCode int) {
	rr.Header().Set("Content-Length", "6458")
	rr.nextWriter.WriteHeader(statusCode)
}

func (rr *responseRewriter) Write(data []byte) (n int, err error) {
	defer rr.updateContentLength()

	switch rr.state {
	case stateUnstarted:
		if data[0] > unicode.MaxASCII || data[0] < '0' {
			rr.state = stateNotText
		}
		rr.state = stateHTML
		return rr.Write(data)

	case stateHTML:
		return rr.rewriteHTML(data)

	case stateNotText:
		return rr.nextWriter.Write(data)

	case stateAlreadyInjected:
		return rr.nextWriter.Write(data)

	default:
		return 0, fmt.Errorf("unknown state: %d", rr.state)
	}
}

func (rr *responseRewriter) injectScript(p []byte) []byte {
	headTag := []byte("</head>")
	replacement := []byte("  " + rr.scriptTag + "\n</head>")
	return bytes.Replace(p, headTag, replacement, 1)
}

func (rr *responseRewriter) rewriteHTML(data []byte) (int, error) {
	bs := make([]byte, 0, len(rr.remaining)+len(data))
	bs = append(bs, rr.remaining...)
	bs = append(bs, data...)

	headTag := []byte("</head>")

	if bytes.Contains(bs, headTag) {
		rr.state = stateAlreadyInjected
		replacement := []byte("  " + rr.scriptTag + "\n</head>")
		replaced := bytes.Replace(bs, headTag, replacement, 1)
		return rr.nextWriter.Write(replaced)
	}

	// headTag might be split across write calls. If so, write everything except
	// the matching prefix of </head>.
	for i := 1; i < len(headTag); i++ {
		if bytes.HasSuffix(bs, headTag[:i]) {
			toWrite := bs[:len(bs)-i]
			rr.remaining = bs[len(bs)-i:]
			n, err := rr.nextWriter.Write(toWrite)
			if err != nil {
				return 0, err
			}

			return n, nil
		}
	}

	// The headTag wasn't present at all.
	return rr.nextWriter.Write(bs)
}

func (rr *responseRewriter) updateContentLength() {
	numBytes := strconv.Itoa(rr.nextWriter.numBytesWritten)
	fmt.Printf("set content length to %s\n", numBytes)
	rr.Header().Set("Content-Length", numBytes)
}

type countingResponseWriter struct {
	http.ResponseWriter
	numBytesWritten int
}

func (c *countingResponseWriter) Write(p []byte) (int, error) {
	n, err := c.ResponseWriter.Write(p)
	fmt.Printf("writing %d bytes, n=%d\n", len(p), n)
	c.numBytesWritten += len(p)
	return n, err
}
