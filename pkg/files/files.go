package files

import (
	"fmt"
	"hash/fnv"
	"io"
	"os"
)

// IsSameBytes returns true if path1 and path2 have the same bytes, tested with a
// hash.
func IsSameBytes(path1, path2 string) (bool, error) {
	// Check if the same exact file.
	info1, err := os.Stat(path1)
	if err != nil {
		return false, err
	}
	info2, err := os.Stat(path2)
	if err != nil {
		return false, err
	}
	if os.SameFile(info1, info2) {
		return true, nil
	}

	if info1.Mode().IsRegular() && info2.Mode().IsRegular() {
		if info1.Size() != info2.Size() {
			return false, nil
		}
	}

	hash1, err := hashFileBytes(path1)
	if err != nil {
		return false, err
	}
	hash2, err := hashFileBytes(path2)
	if err != nil {
		return false, err
	}
	return hash1 == hash2, nil
}

func hashFileBytes(path string) (uint64, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, fmt.Errorf("failed to open file for hashing: %w", err)
	}
	h := fnv.New64()
	if _, err := io.Copy(h, f); err != nil {
		return 0, fmt.Errorf("failed to copy file contents to hash")
	}
	return h.Sum64(), nil
}

// HashContentsFnv64 hashes the contents of the file at path using a 64-bit FNV-1a hash.
func HashContentsFnv64(path string) (uint64, error) {
	bs, err := os.ReadFile(path)
	if err != nil {
		return 0, fmt.Errorf("read path %s to hash with FNV64: %w", path, err)
	}
	hasher := fnv.New64a()
	_, _ = hasher.Write(bs)
	return hasher.Sum64(), nil
}
