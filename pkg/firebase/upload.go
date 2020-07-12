package firebase

import (
	"bytes"
	"fmt"
	"github.com/jschaf/b2/pkg/errs"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
	"io/ioutil"
	"net/http"
	"time"
)

// Uploader uploads SiteFiles for a Firebase site version.
type Uploader struct {
	siteHashes *SiteHashes
	baseURL    string
	tokSrc     oauth2.TokenSource
	l          *zap.SugaredLogger
}

func NewUploader(siteHashes *SiteHashes, baseUploadURL string, tokSrc oauth2.TokenSource, l *zap.SugaredLogger) *Uploader {
	return &Uploader{
		siteHashes: siteHashes,
		baseURL:    baseUploadURL,
		tokSrc:     tokSrc,
		l:          l,
	}
}

func (u *Uploader) Upload(ctx context.Context, f SiteFile) (mErr error) {
	gzBytes := u.siteHashes.GzipContent(f)
	if gzBytes == nil {
		gzBuf := bytes.Buffer{}
		_, err := GzipFile(f.Path, &gzBuf)
		if err != nil {
			return fmt.Errorf("upload - gzip file: %w", err)
		}
		gzBytes := gzBuf.Bytes()
		sum := SHA256Sum(gzBytes)
		if sum != f.Hash {
			return fmt.Errorf("hash mismatch after recalculating %s, orig=%s, got=%s", f.Path, f.Hash, sum)
		}
	}

	shaUrl := u.baseURL + "/" + string(f.Hash)
	u.l.Debugf("uploading %s url=%s", f.URL, shaUrl)
	req, err := http.NewRequestWithContext(ctx, "POST", shaUrl, bytes.NewReader(gzBytes))
	if err != nil {
		return fmt.Errorf("upload - new request: %w", err)
	}
	token, err := u.tokSrc.Token()
	if err != nil {
		return fmt.Errorf("get bearer token: %w", err)
	}
	req.Header.Add("Content-Type", "application/octet-stream")
	req.Header.Add("Authorization", "Bearer "+token.AccessToken)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("upload - response: %w", err)
	}
	defer errs.CloseWithErrCapture(&mErr, resp.Body, "upload - close response body")

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("upload - read response body: %w", err)
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("upload - non-200 response: %d\n%s", resp.StatusCode, string(content))
	}
	return nil
}

func (u *Uploader) UploadAll(ctx context.Context, fs []SiteFile) error {
	start := time.Now()
	const maxUploads = 16
	sem := semaphore.NewWeighted(maxUploads)
	g, ctx := errgroup.WithContext(ctx)
	for _, f := range fs {
		f := f
		if err := sem.Acquire(ctx, 1); err != nil {
			return fmt.Errorf("acquire semaphore: %w", err)
		}
		g.Go(func() error {
			defer sem.Release(1)
			if err := u.Upload(ctx, f); err != nil {
				return fmt.Errorf("upload %s: %w", f.URL, err)
			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return fmt.Errorf("upload all wait err group: %w", err)
	}
	u.l.Infof("uploaded %d site files in %.3f seconds",
		len(fs), time.Since(start).Seconds())
	return nil
}
