package main

import (
	"fmt"
	"os"			

)

func main(){
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "init":
		HandleInit()
	default:
		fmt.Printf("Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}

}




func printUsage() {
	print("init - Init")
}
