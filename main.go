package main

import (
	_ "embed"
	"fmt"
	"os"
)

//go:embed usage.txt
var usage string

var version = "0.0.0"

func main() {
	if len(os.Args) == 1 {
		printUsage()
		os.Exit(1)
	}

	if len(os.Args) == 2 && (os.Args[1] == "-v" || os.Args[1] == "--version") {
		fmt.Println(version)
		return
	}

	switch os.Args[1] {
	case "archive":
		if len(os.Args) > 3 {
			printUsage()
			os.Exit(1)
		}
		var title string
		if len(os.Args) == 3 {
			title = os.Args[2]
		}
		if err := archiveNotepad(title); err != nil {
			fmt.Println("notepad:", err)
			os.Exit(1)
		}
	case "open":
		if len(os.Args) > 2 {
			printUsage()
			os.Exit(1)
		}
		if err := openNotepad(); err != nil {
			fmt.Println("notepad:", err)
			os.Exit(1)
		}
	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Print(usage)
}
