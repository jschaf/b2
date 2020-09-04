// deploy deploys the contents of the public directory to firebase.
package main

import (
	"fmt"
	"github.com/jschaf/b2/pkg/dirs"
	"github.com/jschaf/b2/pkg/firebase"
	"github.com/jschaf/b2/pkg/logs"
	"github.com/jschaf/b2/pkg/sites"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log"
	"time"

	"golang.org/x/net/context"

	hosting "google.golang.org/api/firebasehosting/v1beta1"
	"google.golang.org/api/option"
)

const (
	siteName     = "joe-blog-314159"
	siteParent   = "sites/" + siteName
	deployPubDir = dirs.PublicMemfs
)

func deploy(l *zap.SugaredLogger) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()

	l.Infof("start deployment")
	start := time.Now()

	accountCreds, err := firebase.ReadServiceAccountCreds()
	if err != nil {
		return err
	}
	tokSource := firebase.NewTokenSource(ctx, accountCreds)

	svc, err := hosting.NewService(ctx, option.WithTokenSource(tokSource))
	if err != nil {
		return fmt.Errorf("new hosting service: %w", err)
	}
	versionSvc := svc.Projects.Sites.Versions

	// Create version: we'll eventually release this version.
	createVersionStart := time.Now()
	createVersion := versionSvc.Create(siteParent, &hosting.Version{})
	createVersion.Context(ctx)
	version, err := createVersion.Do()
	if err != nil {
		return fmt.Errorf("create site version: %w", err)
	}
	l.Infof("created new version %q in %.3f seconds",
		version.Name, time.Since(createVersionStart).Seconds())

	// Populate files: get the SHA256 hash of all gzipped files in the public
	// directory, send them to Firebase with the URL that serves the file.
	// Firebase returns the SHA256 hashes of the files we need to upload to
	// firebase.
	siteHashes := firebase.NewSiteHashes(l)
	if err := siteHashes.PopulateFromDir(deployPubDir); err != nil {
		return fmt.Errorf("populate from dir: %w", err)
	}
	popFilesStart := time.Now()
	popFilesReq := hosting.PopulateVersionFilesRequest{Files: siteHashes.HashesByURL()}
	popFiles := versionSvc.PopulateFiles(version.Name, &popFilesReq)
	popFiles.Context(ctx)
	popFilesResp, err := popFiles.Do()
	if err != nil {
		return fmt.Errorf("populate files: %w", err)
	}
	l.Infof("populate files response requests %d files to upload in %.3f seconds",
		len(popFilesResp.UploadRequiredHashes), time.Since(popFilesStart).Seconds())

	// Upload files: only upload files that have a SHA256 hash in the populate
	// files response
	filesToUpload, err := siteHashes.FindFilesForHashes(popFilesResp.UploadRequiredHashes)
	if err != nil {
		return fmt.Errorf("find files for hashes: %w", err)
	}
	uploader := firebase.NewUploader(siteHashes, popFilesResp.UploadUrl, tokSource, l)
	if err := uploader.UploadAll(ctx, filesToUpload); err != nil {
		return fmt.Errorf("upload all: %w", err)
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
	l, err := logs.NewShortDevSugaredLogger(zapcore.InfoLevel)
	if err != nil {
		log.Fatal(err.Error())
	}

	if err := sites.Rebuild(deployPubDir, l.Desugar()); err != nil {
		l.Fatalf("rebuild site: %s", err)
	}

	if err := deploy(l); err != nil {
		l.Fatalf("failed deploy: %s", err)
	}
}
