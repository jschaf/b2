// deploy deploys the contents of the public directory to firebase.
package main

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/jschaf/b2/pkg/errs"
	"github.com/jschaf/b2/pkg/git"
	"github.com/jschaf/b2/pkg/logs"
	"github.com/jschaf/b2/pkg/sites"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"golang.org/x/net/context"

	hosting "google.golang.org/api/firebasehosting/v1beta1"
	"google.golang.org/api/option"
)

const (
	authFile   = "/home/joe/.config/firebase/b2-admin-sdk.json"
	siteName   = "joe-blog-314159"
	siteParent = "sites/" + siteName
)

type bearerToken = string

type ServiceAccountCreds struct {
	Type         string `json:"type"`
	ProjectID    string `json:"project_id"`
	PrivateKeyID string `json:"private_key_id"`
	PrivateKey   string `json:"private_key"`
	ClientEmail  string `json:"client_email"`
	ClientID     string `json:"client_id"`
	AuthURI      string `json:"auth_uri"`
	TokenURI     string `json:"token_uri"`
}

func readServiceAccountCreds() (s ServiceAccountCreds, mErr error) {
	b, err := ioutil.ReadFile(authFile)
	if err != nil {
		return s, fmt.Errorf("read service account creds: %w", err)
	}
	if err := json.Unmarshal(b, &s); err != nil {
		return s, fmt.Errorf("unmarshal service account creds: %w", err)
	}
	return s, nil
}

func newTokenSource(accountCreds ServiceAccountCreds, ctx context.Context) oauth2.TokenSource {
	cfg := &jwt.Config{
		Email:      accountCreds.ClientEmail,
		PrivateKey: []byte(accountCreds.PrivateKey),
		Scopes:     []string{hosting.FirebaseScope},
		TokenURL:   google.JWTTokenURL,
	}
	tokSource := cfg.TokenSource(ctx)
	return tokSource
}

func gzipFile(path string, w io.Writer) (n int64, mErr error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, fmt.Errorf("gzip file: %w", err)
	}
	defer errs.CloseWithErrCapture(&mErr, f, "close file")
	zw := gzip.NewWriter(w)
	defer errs.CloseWithErrCapture(&mErr, zw, "close gzip writer")
	zw.Name = path
	n, err = io.Copy(zw, f)
	return n, err
}

func shaSum256String(b []byte) string {
	sum := sha256.Sum256(b)
	return fmt.Sprintf("%x", sum)
}

// buildPopulateFilesMap creates a map of a file's URL path to the SHA-256
// hash of the gzipped contents of the file. Adds a map entry for every file in
// dir recursively.
func buildPopulateFilesMap(dir string, l *zap.SugaredLogger) (map[string]string, error) {
	files := make(map[string]string)
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if isDir, err := isSymlinkDir(path, info); err != nil {
			return err
		} else if isDir {
			return nil
		}

		var buf = bytes.Buffer{}
		size := int(info.Size() / 4)
		buf.Grow(size)
		_, err = gzipFile(path, &buf)
		if err != nil {
			return err
		}
		// The key is the desired path on the website for the file, e.g. /foo
		// or /bar/baz.
		urlPath := strings.TrimPrefix(path, dir)
		sum := shaSum256String(buf.Bytes())
		l.Debugf("found url=%s sha=%s", urlPath, sum)
		files[urlPath] = sum
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("populate files walk: %w", err)
	}
	return files, nil
}

func isSymlinkDir(path string, info os.FileInfo) (bool, error) {
	if info.Mode()&os.ModeSymlink == 0 {
		return false, nil
	}

	resolved, err := os.Readlink(path)
	if err != nil {
		return false, fmt.Errorf("populate files resolve symlink: %w", err)
	}
	resolvedInfo, err := os.Lstat(resolved)
	if err != nil {
		return false, fmt.Errorf("populate files lstat symlink: %w", err)
	}
	return resolvedInfo.IsDir(), nil
}

// findFilesToUpload returns a slice of file paths that should be uploaded.
// baseDir contains the baseDir to turn a URL path into a file path.
// files contains a map of a URL path to the gzipped SHA-256 hash of the file.
// hashes contains the file hashes that should be uploaded.
func findFilesToUpload(baseDir string, files map[string]string, hashes []string) []string {
	sort.Strings(hashes)
	filePaths := make([]string, 0, len(hashes))
	for urlPath, shaSum := range files {
		n := sort.SearchStrings(hashes, shaSum)
		if n < len(hashes) && hashes[n] == shaSum {
			filePaths = append(filePaths, filepath.Join(baseDir, urlPath))
		}
	}
	return filePaths
}

func uploadFile(ctx context.Context, baseDir, path, uploadURL string, tok bearerToken, l *zap.SugaredLogger) (mErr error) {
	gzBuf := bytes.Buffer{}
	_, err := gzipFile(path, &gzBuf)
	if err != nil {
		return fmt.Errorf("upload - gzip file: %w", err)
	}
	shaSum := shaSum256String(gzBuf.Bytes())

	urlPath := strings.TrimPrefix(path, baseDir)
	shaUrl := uploadURL + "/" + shaSum
	l.Debugf("uploading %s, sha=%s, url=%s", urlPath, shaSum, shaUrl)
	req, err := http.NewRequestWithContext(ctx, "POST", shaUrl, &gzBuf)
	if err != nil {
		return fmt.Errorf("upload - new post request: %w", err)
	}
	req.Header.Add("Content-Type", "application/octet-stream")
	req.Header.Add("Authorization", "Bearer "+tok)
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

func run(l *zap.SugaredLogger) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()

	l.Infof("start deployment")
	start := time.Now()

	accountCreds, err := readServiceAccountCreds()
	if err != nil {
		return err
	}
	tokSource := newTokenSource(accountCreds, ctx)

	svc, err := hosting.NewService(ctx, option.WithTokenSource(tokSource))
	if err != nil {
		return fmt.Errorf("new hosting service: %w", err)
	}
	versionSvc := svc.Projects.Sites.Versions

	// Create version: we'll eventually release this version.
	createVersion := versionSvc.Create(siteParent, &hosting.Version{})
	createVersion.Context(ctx)
	version, err := createVersion.Do()
	if err != nil {
		return fmt.Errorf("create site version: %w", err)
	}
	l.Infof("created new version: %s", version.Name)

	// Populate files: get the SHA256 hash of all gzipped files in the public
	// directory, send them to Firebase with the URL that serves the file and
	// receive the SHA256 hashes of the files we need to upload to firebase.
	popFilesStart := time.Now()
	root, err := git.FindRootDir()
	if err != nil {
		return fmt.Errorf("find root dir: %w", err)
	}
	pubDir := filepath.Join(root, "public")
	fileSums, err := buildPopulateFilesMap(pubDir, l)
	if err != nil {
		return fmt.Errorf("build populate files: %w", err)
	}
	l.Debugf("generated populate files map in %.3f seconds", time.Since(popFilesStart).Seconds())
	l.Infof("found %d files to populate", len(fileSums))
	popFilesReq := hosting.PopulateVersionFilesRequest{Files: fileSums}
	popFiles := versionSvc.PopulateFiles(version.Name, &popFilesReq)
	popFiles.Context(ctx)
	popFilesResp, err := popFiles.Do()
	if err != nil {
		return fmt.Errorf("populate files: %w", err)
	}
	l.Infof("populate files response requests %d files to upload", len(popFilesResp.UploadRequiredHashes))

	// Upload files: only upload files that have a SHA256 hash in the populate
	// files response
	uploads := findFilesToUpload(pubDir, fileSums, popFilesResp.UploadRequiredHashes)
	token, err := tokSource.Token()
	if err != nil {
		return fmt.Errorf("get bearer token: %w", err)
	}
	for _, upload := range uploads {
		l.Infof("uploading %s", upload)
		err := uploadFile(ctx, pubDir, upload, popFilesResp.UploadUrl, token.AccessToken, l)
		if err != nil {
			return fmt.Errorf("upload %s: %w", upload, err)
		}
	}

	// Finalize version: prevent adding any new resources.
	versionFinal := hosting.Version{Status: "FINALIZED"}
	patchVersion := versionSvc.Patch(version.Name, &versionFinal)
	patchVersion.Context(ctx)
	patchVersionResp, err := patchVersion.Do()
	if err != nil {
		return fmt.Errorf("finalize version: %w", err)
	}
	if patchVersionResp.Status != "FINALIZED" {
		return fmt.Errorf("finalize version status not 'FINALIZED', got %q", patchVersionResp.Status)
	}

	// Release version: promote a version to release so it's shown on the website.
	release := hosting.Release{}
	createRelease := svc.Sites.Releases.Create(siteParent, &release)
	createRelease.Context(ctx)
	createRelease.VersionName(patchVersionResp.Name)
	createReleaseResp, err := createRelease.Do()
	if err != nil {
		return fmt.Errorf("create release: %w", err)
	}
	l.Infof("created release: %s", createReleaseResp.Name)

	l.Infof("completed deployment in %.3f seconds", time.Since(start).Seconds())
	return nil
}

func main() {
	l, err := logs.NewShortDevSugaredLogger(zapcore.DebugLevel)
	if err != nil {
		log.Fatal(err.Error())
	}

	if err := sites.Rebuild(l.Desugar()); err != nil {
		l.Fatalf("rebuild site: %s", err)
	}

	if err := run(l); err != nil {
		l.Fatalf("failed deploy: %s", err)
	}
}
