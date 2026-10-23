package objstore

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewCreatesDirs(t *testing.T) {
	root := t.TempDir()

	s, err := New(root)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	if _, err := os.Stat(s.objDir); err != nil {
		t.Errorf("objDir not created: %v", err)
	}
	if _, err := os.Stat(s.tmpDir); err != nil {
		t.Errorf("tmpDir not created: %v", err)
	}
}

func TestPutStoresRetrievableBlob(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)

	s, err := New(".")
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	content := []byte("hello fileshare")
	addr, err := s.Put(bytes.NewReader(content), int64(len(content)))
	if err != nil {
		t.Fatalf("Put returned error: %v", err)
	}
	if len(addr) != 64 {
		t.Errorf("expected 64 char hex digest, got %d chars: %q", len(addr), addr)
	}

	dst := filepath.Join(s.objDir, addr[:2], addr[2:])
	info, err := os.Stat(dst)
	if err != nil {
		t.Fatalf("expected blob file at %s: %v", dst, err)
	}
	if info.Mode().Perm() != 0o444 {
		t.Errorf("expected blob file mode 0444, got %v", info.Mode().Perm())
	}
}

func TestPutIsContentAddressedAndIdempotent(t *testing.T) {
	root := t.TempDir()
	s, err := New(root)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	content := []byte("same content twice")
	addr1, err := s.Put(bytes.NewReader(content), int64(len(content)))
	if err != nil {
		t.Fatalf("first Put returned error: %v", err)
	}
	addr2, err := s.Put(bytes.NewReader(content), int64(len(content)))
	if err != nil {
		t.Fatalf("second Put returned error: %v", err)
	}
	if addr1 != addr2 {
		t.Errorf("expected identical content to hash the same: %q != %q", addr1, addr2)
	}

	entries, err := os.ReadDir(s.objDir)
	if err != nil {
		t.Fatalf("ReadDir on objDir failed: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("expected exactly one shard dir after storing identical content twice, got %d", len(entries))
	}
}

func TestPutDifferentContentDifferentAddr(t *testing.T) {
	root := t.TempDir()
	s, err := New(root)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	addr1, err := s.Put(strings.NewReader("content a"), int64(len("content a")))
	if err != nil {
		t.Fatalf("Put a returned error: %v", err)
	}
	addr2, err := s.Put(strings.NewReader("content b"), int64(len("content b")))
	if err != nil {
		t.Fatalf("Put b returned error: %v", err)
	}
	if addr1 == addr2 {
		t.Errorf("expected different content to hash differently, both got %q", addr1)
	}
}

func TestPutSizeMismatchErrors(t *testing.T) {
	root := t.TempDir()
	s, err := New(root)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	content := []byte("short")
	_, err = s.Put(bytes.NewReader(content), int64(len(content))+10)
	if err == nil {
		t.Fatal("expected error when declared size does not match actual content length")
	}
}

func TestHasReturnsFalseForMissingHash(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)

	if _, err := New("."); err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	missing := strings.Repeat("0", 64)
	if Has(missing) {
		t.Error("expected Has to return false for a hash that was never stored")
	}
}

func TestHasReturnsTrueAfterPut(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)

	s, err := New(".")
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	content := []byte("tracked by Has")
	addr, err := s.Put(bytes.NewReader(content), int64(len(content)))
	if err != nil {
		t.Fatalf("Put returned error: %v", err)
	}

	if !Has(addr) {
		t.Error("expected Has to return true for a hash that was stored")
	}
}

func TestGetReturnsStoredContent(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)

	s, err := New(".")
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	content := []byte("round trip through Get")
	addr, err := s.Put(bytes.NewReader(content), int64(len(content)))
	if err != nil {
		t.Fatalf("Put returned error: %v", err)
	}

	rc, err := Get(addr)
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	defer rc.Close()

	got, err := io.ReadAll(rc)
	if err != nil {
		t.Fatalf("reading from Get result failed: %v", err)
	}
	if !bytes.Equal(got, content) {
		t.Errorf("Get content mismatch: got %q, want %q", got, content)
	}
}

func TestGetReturnsErrorForMissingHash(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)

	if _, err := New("."); err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	missing := strings.Repeat("0", 64)
	if _, err := Get(missing); err == nil {
		t.Error("expected Get to return an error for a hash that was never stored")
	}
}
