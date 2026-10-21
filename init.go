package main

import (
	"fileshare/internal/objstore"
	"fileshare/internal/scan"
	"fmt"
	"os"
)

func HandleInit() error {
	store, err := objstore.New(".")
	if err != nil {
		return fmt.Errorf("Error creating directory %v\n", err)
	}
	fmt.Printf("store: %v\n", store)

	scanned, err := scan.Scan(".")
	if err != nil {
		return fmt.Errorf("Error scanning directory: %v\n", err)
	}

	for _, file := range scanned {
		f, err := os.Open(file)
		if err != nil {
			return fmt.Errorf("Error opening file: %s\n", file)
		}
		defer f.Close()

		info, _ := f.Stat()
		_, err = store.Put(f, info.Size())
		if err != nil {
			return fmt.Errorf("Error creating blob file: %s\n", file)
		}
	}

	return nil
}
