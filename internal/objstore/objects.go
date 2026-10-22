package objstore

import (
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

var foldername = ".gosync.tmp"

type Store struct {
	objDir string
	tmpDir string
}

func New(root string) (*Store, error) {
	s := &Store{
		objDir: filepath.Join(root, foldername, "objects"),
		tmpDir: filepath.Join(root, foldername, "tmp"),
	}
	for _, d := range []string{s.objDir, s.tmpDir} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return nil, err
		}
	}
	return s, nil
}

func (s *Store) Put(r io.Reader, size int64) (string, error) {
	tmp, err := os.CreateTemp(s.tmpDir, "blob-*")
	if err != nil {
		return "", fmt.Errorf("create temp: %w", err)
	}
	defer os.Remove(tmp.Name())

	h := sha256.New()
	fmt.Fprintf(h, "blob %d\x00", size)
	gz, _ := gzip.NewWriterLevel(tmp, gzip.BestSpeed)

	n, err := io.Copy(io.MultiWriter(h, gz), r)
	if err != nil {
		tmp.Close()
		return "", err
	}
	if n != size {
		tmp.Close()
		return "", errors.New("file changed during read")
	}
	if err := gz.Close(); err != nil {
		tmp.Close()
		return "", err
	}
	if err := tmp.Close(); err != nil {
		return "", err
	}

	addr := hex.EncodeToString(h.Sum(nil))
	dst := filepath.Join(s.objDir, addr[:2], addr[2:])
	if _, err := os.Stat(dst); err == nil {
		return addr, nil
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return "", err
	}
	if err := os.Rename(tmp.Name(), dst); err != nil {
		return "", err
	}
	return addr, os.Chmod(dst, 0o444)
}

func Get(hash string) (io.ReadCloser, error) {
	path := filepath.Join(".", foldername, "objects", hash[:2], hash[2:])

	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("Error opening file using Get: %v\n", err)
	}

	gz, err := gzip.NewReader(f)
	if err != nil {
		return nil, fmt.Errorf("Error creating gzip reader for file: %v\n", err)
	}
	defer gz.Close()

	return gz, nil
}

func Has(hash string) bool {
	path := filepath.Join(".", foldername, "objects", hash[:2], hash[2:])
	_, err := os.Stat(path)

	if err != nil {
		return false
	}

	return true
}
