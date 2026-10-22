package scan

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
)

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("failed to create dir for %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write %s: %v", path, err)
	}
}

func relResults(t *testing.T, root string, got []string) []string {
	t.Helper()
	rel := make([]string, len(got))
	for i, p := range got {
		r, err := filepath.Rel(root, p)
		if err != nil {
			t.Fatalf("filepath.Rel failed: %v", err)
		}
		rel[i] = r
	}
	sort.Strings(rel)
	return rel
}

func TestScanIncludesRegularFiles(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "a.txt"), "a")
	writeFile(t, filepath.Join(root, "sub", "b.txt"), "b")

	got, err := Scan(root)
	if err != nil {
		t.Fatalf("Scan returned error: %v", err)
	}

	rel := relResults(t, root, got)
	want := []string{"a.txt", filepath.Join("sub", "b.txt")}
	sort.Strings(want)

	if len(rel) != len(want) {
		t.Fatalf("got %v, want %v", rel, want)
	}
	for i := range want {
		if rel[i] != want[i] {
			t.Errorf("got %v, want %v", rel, want)
			break
		}
	}
}

func TestScanSkipsHiddenDirsEntirely(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "visible.txt"), "v")
	writeFile(t, filepath.Join(root, ".hidden-dir", "inside.txt"), "hidden")

	got, err := Scan(root)
	if err != nil {
		t.Fatalf("Scan returned error: %v", err)
	}

	rel := relResults(t, root, got)
	if len(rel) != 1 || rel[0] != "visible.txt" {
		t.Errorf("expected only visible.txt, got %v", rel)
	}
}

func TestScanSkipsHiddenAndCommonlyIgnoredFiles(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "keep.txt"), "k")
	writeFile(t, filepath.Join(root, ".hidden"), "h")
	writeFile(t, filepath.Join(root, "Thumbs.db"), "t")
	writeFile(t, filepath.Join(root, "desktop.ini"), "d")
	writeFile(t, filepath.Join(root, "~$lock.docx"), "l")
	writeFile(t, filepath.Join(root, "._resource"), "r")

	got, err := Scan(root)
	if err != nil {
		t.Fatalf("Scan returned error: %v", err)
	}

	rel := relResults(t, root, got)
	if len(rel) != 1 || rel[0] != "keep.txt" {
		t.Errorf("expected only keep.txt, got %v", rel)
	}
}

func TestScanExcludesSyncIgnoreFileItself(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "keep.txt"), "k")
	writeFile(t, filepath.Join(root, ".syncignore"), "ignored.txt\n")
	writeFile(t, filepath.Join(root, "ignored.txt"), "i")

	got, err := Scan(root)
	if err != nil {
		t.Fatalf("Scan returned error: %v", err)
	}

	rel := relResults(t, root, got)
	for _, p := range rel {
		if p == ".syncignore" {
			t.Errorf(".syncignore should never be included in scan results, got %v", rel)
		}
	}

	// scanSyncIgnore only reads and prints the file's lines today; it does not
	// yet filter matching paths out of the scan, so ignored.txt is still present.
	found := false
	for _, p := range rel {
		if p == "ignored.txt" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected ignored.txt to still be present given current scanSyncIgnore behavior, got %v", rel)
	}
}

func TestScanEmptyDir(t *testing.T) {
	root := t.TempDir()

	got, err := Scan(root)
	if err != nil {
		t.Fatalf("Scan returned error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected no files in an empty directory, got %v", got)
	}
}
