package scan

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
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

		if name == ".syncignore" {
			err := scanSyncIgnore(path)
			if err != nil {
				return err
			}
			return nil
		}

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

func scanSyncIgnore(path string) error {
	b, err := os.Open(path)

	if err != nil {
		return fmt.Errorf("Cant read syncfile: %v\n", err)
	}

	scanner := bufio.NewScanner(b)

	for scanner.Scan() {
		line := scanner.Text()
		ignoreExact[line] = true
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("Error reading syncfile with scanner: %v\n", err)
	}
	return nil
}
