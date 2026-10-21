package main

import (
	"errors"
	"fmt"
	"os"
)

func HandleInit() {
	folderName := os.Getenv("FOLDERNAME")

	info, err := os.Stat(folderName)

	if err == nil {
		if info.IsDir() {
			fmt.Print("already a sync directory! Perhaps you meant the sync command?")
		} else {
			fmt.Print("There is a file with the name .goSync in this directory! Please rename this before you can use sync!")
		}
	}

	if errors.Is(err, os.ErrNotExist) {
		if createSubDir(folderName) != nil {
			return
		}
		detectFiles()
	}
}

func createSubDir(folderName string) error {
	err := os.MkdirAll(folderName, 0755)

	if err != nil {
		fmt.Printf("Failed to create directory: %v\n", err)
		return err
	}
	return nil

}

func detectFiles() error {
	files, err := os.ReadDir(".")

	trackedFiles := files

	if err != nil {
		fmt.Printf("Failed to read directory: %v\n", err)
		return err
	}

	for _, file := range trackedFiles {
		if !file.IsDir() {
			//trackedFiles = append(trackedFiles, os.ReadDir("./")) This is a graph problem do it later
			f, _ := file.Info()
			createBlob(f)
		}
	}

	return nil
}

func createBlob(file os.FileInfo) error {
	if file.IsDir() {
		fmt.Printf("Cannot parse directories. Tried parsing: %s", file.Name())
	}
}
