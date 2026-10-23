package main

import (
	"errors"
	"fileshare/src/internal/objstore"
	"fileshare/src/internal/scan"
	"io/fs"
	"os"
	"path/filepath"
)

func Init(root string) error {
	final := filepath.Join(root, ".gosync")
	build := filepath.Join(root, ".gosync.tmp")

	if _, err := os.Stat(final); err == nil {
		return ErrAlreadyInit
	} else if !errors.Is(err, fs.ErrNotExist) {
		return err
	}

	if err := os.RemoveAll(build); err != nil {
		return err
	}

	committed := false
	defer func() {
		if !committed {
			os.RemoveAll(build)
		}
	}()

	blobs, err := objstore.New(build, "blob", "objects")
	if err != nil {
		return err
	}
	trees, err := objstore.New(build, "tree", "trees")
	if err != nil {
		return err
	}

	rootHash, err := buildSnapshot(root, blobs, trees)
	if err != nil {
		return err
	}

	if err := os.WriteFile(filepath.Join(build, "HEAD"), []byte(rootHash+"\n"), 0o644); err != nil {
		return err
	}

	if err := os.Rename(build, final); err != nil {
		return err
	}
	committed = true
	return nil
}

func buildSnapshot(root string, blobs *objstore.Store, trees *objstore.Store) (string, error) {
	// 1 - Scan files and create blobs

	files, err := scan.Scan(root)
	if err != nil {
		return "", err
	}

	addrlist := make([]string, len(files))

	for _, file := range files {
		f, err := os.Open(file)
		if err != nil {
			return "", err
		}
		info, err := f.Stat()
		if err != nil {
			return "", err
		}
		addr, err := blobs.Put(f, info.Size())
		if err != nil {
			return "", err
		}
		addrlist = append(addrlist, addr)
	} // First one should be root

	// 2 - Create tree and store
	// 3 - Return root node hash
}
