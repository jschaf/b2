package firebase

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/jschaf/jsc/pkg/errs"
	"github.com/karrick/godirwalk"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

func GzipFile(path string, w io.Writer) (n int64, mErr error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, fmt.Errorf("gzip file: %w", err)
	}
	defer errs.Capture(&mErr, f.Close, "close file")
	var gzipW *gzip.Writer
	switch filepath.Ext(path) {
	case ".png", ".pdf":
		zw, _ := gzip.NewWriterLevel(w, gzip.NoCompression)
		gzipW = zw
	default:
		zw, _ := gzip.NewWriterLevel(w, gzip.DefaultCompression)
		gzipW = zw
	}

	defer errs.Capture(&mErr, gzipW.Close, "close gzip writer")
	gzipW.Name = path
	n, err = io.Copy(gzipW, f)
	return n, err
}

// SHA256Sum returns a string of the SHA256 hash of a byte slice using a
// hexadecimal encoding.
func SHA256Sum(b []byte) FileHash {
	sum := sha256.Sum256(b)
	return FileHash(fmt.Sprintf("%x", sum))
}

// FileHash is a hash of file content.
type FileHash string

type SiteFile struct {
	URL  string   // the URL that serves the file, e.g. /foo/index.html
	Path string   // the local, absolute file path.
	Hash FileHash // the SHA256 hash of the gzipped file contents
}

// SiteHashes is a bidirectional map of a site file to the SHA-256 hash of the
// gzipped contents of the file. This is a useful data structure for uploading
// to Firebase because deploying requires:
//
//  1. Uploading the URL path and the corresponding SHA256 hash of the gzipped
//     file contents via PopulateFiles.
//  2. Firebase responds with a list of SHA256 hashes should be uploaded and a
//     URL.
//  3. We upload each gzipped file using the provided URL.
//
// SiteHashes allows finding the file by it's hash and retains the gzipped
// content so we don't need to recompute it.
//
// One downside is that we store the entire contents of the generated site
// in memory. If this is problem moving forward, we can skip the gzipContents
// map and rely on the OS page cache to cache files for us.
type SiteHashes struct {
	hashes       map[SiteFile]FileHash
	files        map[FileHash]SiteFile
	gzipContents map[SiteFile][]byte
	mu           sync.Mutex
}

func NewSiteHashes() *SiteHashes {
	return &SiteHashes{
		hashes:       make(map[SiteFile]FileHash),
		files:        make(map[FileHash]SiteFile),
		gzipContents: make(map[SiteFile][]byte),
		mu:           sync.Mutex{},
	}
}

func (sh *SiteHashes) PopulateFromDir(dir string) error {
	start := time.Now()
	ctx := context.Background()

	sem := semaphore.NewWeighted(int64(runtime.NumCPU()))
	g, ctx := errgroup.WithContext(ctx)

	walkFunc := func(path string, dirent *godirwalk.Dirent) error {
		if isDir, err := dirent.IsDirOrSymlinkToDir(); err != nil {
			return err
		} else if isDir {
			return nil
		}

		g.Go(func() error {
			if err := sem.Acquire(ctx, 1); err != nil {
				return fmt.Errorf("acquire semaphore: %w", err)
			}
			defer sem.Release(1)

			f := SiteFile{
				URL:  strings.TrimPrefix(path, dir),
				Path: path,
			}

			buf := bytes.Buffer{}
			stat, err := os.Stat(path)
			if err != nil {
				return fmt.Errorf("stat file size: %w", err)
			}

			size := stat.Size()
			sizeEst := estimateCompressedSize(path, size)
			buf.Grow(sizeEst)
			_, err = GzipFile(f.Path, &buf)
			if err != nil {
				return fmt.Errorf("populate files gzip: %w", err)
			}
			sum := SHA256Sum(buf.Bytes())
			f.Hash = sum

			gzSize := len(buf.Bytes())
			diff := gzSize - sizeEst
			ratio := float64(gzSize) / float64(size)
			if 0 < diff {
				slog.Debug("file was larger than buffer", "path", path, "size", size, "gzip_size", gzSize, "diff", diff, "ratio", ratio)
			} else if diff < -4096 {
				slog.Debug("file was smaller than buffer", "path", path, "size", size, "gzip_size", gzSize, "diff", diff, "ratio", 1/ratio)
			}

			sh.mu.Lock()
			sh.hashes[f] = f.Hash
			sh.files[f.Hash] = f
			sh.gzipContents[f] = buf.Bytes()
			sh.mu.Unlock()

			return nil
		})
		return nil
	}

	err := godirwalk.Walk(dir, &godirwalk.Options{
		FollowSymbolicLinks: true,
		Unsorted:            true,
		Callback:            walkFunc,
	})
	if err != nil {
		return fmt.Errorf("populate from dir - walk: %w", err)
	}

	if err := g.Wait(); err != nil {
		return fmt.Errorf("populate from dir - wait err group: %w", err)
	}

	total := 0
	for _, b := range sh.gzipContents {
		total += len(b)
	}

	slog.Info("populated site files", "count", len(sh.files), "size", total, "duration", time.Since(start))
	return nil
}

// estimateCompressedSize estimates the compressed size of the file at path
// based on its initial size and file extension.
func estimateCompressedSize(path string, size int64) int {
	sizeEst := float64(size)
	switch filepath.Ext(path) {
	case ".png", ".pdf":
		sizeEst += 256 // we don't compress png or PDF
	case ".ico":
		sizeEst /= 5
	case ".html", ".css":
		switch {
		case size < 2048:
			sizeEst /= 1.8
		case size > 8196:
			sizeEst /= 2.65
		default:
			sizeEst /= 2.5
		}
	}
	return int(sizeEst)
}

// HashesByURL returns a map from the URL for a file to the SHA256 hash of the
// gzipped contents of the file.
func (sh *SiteHashes) HashesByURL() map[string]string {
	m := make(map[string]string)
	for file, hash := range sh.hashes {
		m[file.URL] = string(hash)
	}
	return m
}

// FindFilesForHashes returns a slice of site files. Each site file corresponds
// to one of the requested hashes. Errors if a hash has no corresponding site
// file.
func (sh *SiteHashes) FindFilesForHashes(hashes []string) ([]SiteFile, error) {
	filePaths := make([]SiteFile, 0, len(hashes))
	for _, hash := range hashes {
		file, ok := sh.files[FileHash(hash)]
		if !ok {
			return nil, fmt.Errorf("missing file for hash %q", hash)
		}
		filePaths = append(filePaths, file)
	}
	return filePaths, nil
}

// GzipContent return the Gzipped contents of the site file or a nil byte
// slice if no contents are found for the site file.
func (sh *SiteHashes) GzipContent(f SiteFile) []byte {
	return sh.gzipContents[f]
}
