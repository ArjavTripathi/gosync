package scan

import (
	"io/fs"
	"path/filepath"
	"strings"
)

var ignoreExact = map[string]bool{
	".DS_Store":   true,
	"Thumbs.db":   true,
	"desktop.ini": true,
}

func Scan(root string) ([]string, error) {
	var filePaths []string

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			if d.Name() != "." && d.Name()[0] == '.' {
				return filepath.SkipDir
			}
			return nil
		}

		if !d.Type().IsRegular() {
			return nil
		}

		name := d.Name()

		if name[0] == '.' ||
			ignoreExact[name] ||
			strings.HasPrefix(name, "~$") ||
			strings.HasPrefix(name, "._") {
			return nil
		}

		filePaths = append(filePaths, path)

		return nil
	})

	return filePaths, err
}
