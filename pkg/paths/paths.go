package paths

import (
	"context"
	"fmt"
	"github.com/jschaf/b2/pkg/errs"
	"github.com/karrick/godirwalk"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
	"io"
	"os"
	"path/filepath"
)

// WalkUp traverses up directory tree until it finds an ancestor directory that
// contains dirToFind. WalkUp checks the current directory and then
func WalkUp(dirToFind string) (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current working dir: %s", err)
	}

	for dir != string(os.PathSeparator) {
		p := filepath.Join(dir, dirToFind)

		if stat, err := os.Stat(p); err != nil {
			if !os.IsNotExist(err) {
				return "", fmt.Errorf("stat dirToFind %s: %w", p, err)
			}
		} else if stat.IsDir() {
			return dir, nil
		}

		dir = filepath.Dir(dir)
	}
	return "", fmt.Errorf("dir not found in directory tree starting from %s", dir)
}

// WalkConcurrent walks directory recursively calling walkFunc on each entry.
func WalkConcurrent(dir string, maxParallel int, walkFunc godirwalk.WalkFunc) error {
	sem := semaphore.NewWeighted(int64(maxParallel))
	g, ctx := errgroup.WithContext(context.Background())

	callback := func(path string, dirent *godirwalk.Dirent) error {
		if err := sem.Acquire(ctx, 1); err != nil {
			return fmt.Errorf("walk concurrent acquire semaphore: %w", err)
		}
		g.Go(func() error {
			defer sem.Release(1)
			return walkFunc(path, dirent)
		})
		return nil
	}
	err := godirwalk.Walk(dir, &godirwalk.Options{Unsorted: true, Callback: callback})
	if err != nil {
		return fmt.Errorf("walk concurrent walk error: %w", err)
	}
	if err := g.Wait(); err != nil {
		return fmt.Errorf("walk concurrent wait err group: %w", err)
	}
	return nil
}

// Copy the contents of the src file to dest. Any existing file will be
// overwritten and will not copy file attributes.
func Copy(dest, src string) (mErr error) {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer errs.CapturingClose(&mErr, in, "")

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer errs.CapturingClose(&mErr, out, "")

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return nil
}
