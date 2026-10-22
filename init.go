package main

import (
	"errors"
	"fileshare/internal/objstore"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

func Init(root string, out io.Writer) error {
	final := filepath.Join(root, ".gosync")
	build := filepath.Join(root, ".gosync.tmp")

	// 1. already initialized?
	if _, err := os.Stat(final); err == nil {
		return ErrAlreadyInit
	} else if !errors.Is(err, fs.ErrNotExist) {
		return err
	}

	// 2. clear leftovers from a crashed init
	if err := os.RemoveAll(build); err != nil {
		return err
	}

	// 3. build everything inside build/
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

	rootHash, err := buildSnapshot(root, blobs, trees) // walk + Put + tree build
	if err != nil {
		return err
	}

	// HEAD, config, state: plain writes into build/ are fine
	if err := os.WriteFile(filepath.Join(build, "HEAD"), []byte(rootHash+"\n"), 0o644); err != nil {
		return err
	}

	// 4. commit point
	if err := os.Rename(build, final); err != nil {
		return err
	}
	committed = true
	return nil
}
