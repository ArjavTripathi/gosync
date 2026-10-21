package main

import(
	"os"
	"errors"
	"fmt"
)

func HandleInit(){
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
		createSubDir(folderName)	
	}
}

func createSubDir(folderName string){
	err := os.MkdirAll(folderName, 0755)

	if err != nil {
		fmt.Printf("Failed to create directory: %v\n", err)
		return
	}


}
