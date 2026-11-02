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

	if len(os.Args) != 2 || os.Args[1] != "archive" {
		printUsage()
		os.Exit(1)
	}

	// if err := archiveReminders(); err != nil {
	// 	fmt.Println("reminders:", err)
	// 	os.Exit(1)
	// }
}

func printUsage() {
	fmt.Print(usage)
}
