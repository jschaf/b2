package paths

import (
	"context"
	"errors"
	"fmt"
	"github.com/jschaf/b2/pkg/errs"
	"github.com/jschaf/b2/pkg/files"
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

// WalkConcurrent walks dir, recursively calling walkFunc on each entry.
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

type WalkFuncCollectString func(osPathname string, directoryEntry *godirwalk.Dirent) ([]string, error)

// WalkCollectStrings walks dir, recursively calling walkFunc on each entry
// collecting a slice of strings from each walkFunc.
func WalkCollectStrings(dir string, maxParallel int, walkFunc WalkFuncCollectString) ([]string, error) {
	sem := semaphore.NewWeighted(int64(maxParallel))
	g, ctx := errgroup.WithContext(context.Background())

	strs := make([]string, 0, 4)
	strsC := make(chan string)
	doneC := make(chan struct{})
	go func() {
		for d := range strsC {
			strs = append(strs, d)
		}
		close(doneC)
	}()

	callback := func(path string, dirent *godirwalk.Dirent) error {
		if err := sem.Acquire(ctx, 1); err != nil {
			return fmt.Errorf("walk collect acquire semaphore: %w", err)
		}
		g.Go(func() error {
			defer sem.Release(1)
			ss, err := walkFunc(path, dirent)
			if err != nil {
				return err
			}
			for _, s := range ss {
				strsC <- s
			}
			return nil
		})
		return nil
	}
	err := godirwalk.Walk(dir, &godirwalk.Options{Unsorted: true, Callback: callback})
	if err != nil {
		return nil, fmt.Errorf("walk collect walk error: %w", err)
	}
	if err := g.Wait(); err != nil {
		return nil, fmt.Errorf("walk collect wait err group: %w", err)
	}
	close(strsC)
	<-doneC
	return strs, nil
}

// Copy the contents of the src file to dest. Any existing file will be
// overwritten and will not copy file attributes.
func Copy(dest, src string) (mErr error) {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer errs.Capturing(&mErr, in.Close, "")

	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return fmt.Errorf("mkdir to copy to dest %q: %w", dest, err)
	}
	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer errs.Capturing(&mErr, out.Close, "")

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return nil
}

// CopyLazy copies the contents of the src file to dest only if the contents
// are different. If the files are different, the existing file will be
// overwritten and will not copy file attributes. Returns true if the file was
// same, otherwise false.
func CopyLazy(dest, src string) (b bool, mErr error) {
	if isSame, err := files.IsSameBytes(src, dest); errors.Is(err, os.ErrNotExist) {
		// Ok for file not to exist.
	} else if err != nil {
		return false, fmt.Errorf("check if same file before copy: %w", err)
	} else if isSame {
		return true, nil
	}
	err := Copy(dest, src)
	return false, err
}
