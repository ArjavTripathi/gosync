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

type Store struct {
	kind   string
	dir    string
	tmpDir string
}

func New(base, kind, subdir string) (*Store, error) {
	s := &Store{
		kind:   kind,
		dir:    filepath.Join(base, subdir),
		tmpDir: filepath.Join(base, "tmp"),
	}
	for _, d := range []string{s.dir, s.tmpDir} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return nil, err
		}
	}
	return s, nil
}

func (s *Store) objPath(hash string) (string, error) {
	if len(hash) < 3 {
		return "", fmt.Errorf("invalid hash %q", hash)
	}
	return filepath.Join(s.dir, hash[:2], hash[2:]), nil
}

type gzFile struct {
	*gzip.Reader
	f *os.File
}

func (g gzFile) Close() error {
	g.Reader.Close()
	return g.f.Close()
}

func (s *Store) Get(hash string) (io.ReadCloser, error) {
	p, err := s.objPath(hash)
	if err != nil {
		return nil, err
	}
	f, err := os.Open(p)
	if err != nil {
		return nil, fmt.Errorf("open object: %w", err)
	}
	gz, err := gzip.NewReader(f)
	if err != nil {
		f.Close()
		return nil, fmt.Errorf("gzip reader: %w", err)
	}
	return gzFile{gz, f}, nil
}

func (s *Store) Has(hash string) bool {
	p, err := s.objPath(hash)
	if err != nil {
		return false
	}
	_, err = os.Stat(p)
	return err == nil
}

func (s *Store) Put(r io.Reader, size int64) (string, error) {
	tmp, err := os.CreateTemp(s.tmpDir, s.kind+"-*")
	if err != nil {
		return "", fmt.Errorf("create temp: %w", err)
	}
	defer os.Remove(tmp.Name())

	h := sha256.New()
	fmt.Fprintf(h, "%s %d\x00", s.kind, size)

	gz, err := gzip.NewWriterLevel(tmp, gzip.BestSpeed)
	if err != nil {
		tmp.Close()
		return "", fmt.Errorf("create gzip writer: %w", err)
	}

	n, err := io.Copy(io.MultiWriter(h, gz), r)
	if err != nil {
		tmp.Close()
		return "", fmt.Errorf("write %s: %w", s.kind, err)
	}
	if n != size {
		tmp.Close()
		return "", errors.New("file changed during read")
	}
	if err := gz.Close(); err != nil {
		tmp.Close()
		return "", fmt.Errorf("finish gzip: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return "", fmt.Errorf("sync temp: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return "", fmt.Errorf("close temp: %w", err)
	}
	if err := os.Chmod(tmp.Name(), 0o444); err != nil {
		return "", fmt.Errorf("chmod temp: %w", err)
	}

	addr := hex.EncodeToString(h.Sum(nil))
	dst, err := s.objPath(addr)
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(dst); err == nil {
		return addr, nil
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return "", fmt.Errorf("create object dir: %w", err)
	}
	if err := os.Rename(tmp.Name(), dst); err != nil {
		if s.Has(addr) {
			return addr, nil
		}
		return "", fmt.Errorf("store object: %w", err)
	}
	return addr, nil
}
