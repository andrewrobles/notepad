package main

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

//go:embed usage.txt
var usage string

var version = "0.0.0"

func main() {
	if len(os.Args) == 2 && (os.Args[1] == "-v" || os.Args[1] == "--version") {
		fmt.Println(version)
		return
	}

	if len(os.Args) >= 2 && os.Args[1] == "archive" {
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
		return
	}

	// Any other invocation is forwarded to the real Windows Notepad so this
	// tool can live alongside the built-in editor under the same name.
	if err := runRealNotepad(os.Args[1:]); err != nil {
		fmt.Println("notepad:", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Print(usage)
}

// runRealNotepad launches the built-in Windows Notepad with the given
// arguments. The executable is resolved via %SystemRoot% rather than PATH to
// avoid re-invoking this shadowing binary. Set NOTEPAD_EXE to override the
// target executable.
func runRealNotepad(args []string) error {
	exe := os.Getenv("NOTEPAD_EXE")
	if exe == "" {
		root := os.Getenv("SystemRoot")
		if root == "" {
			root = `C:\Windows`
		}

		exe = filepath.Join(root, "System32", "notepad.exe")
		if _, err := os.Stat(exe); err != nil {
			exe = filepath.Join(root, "notepad.exe")
		}
	}

	cmd := exec.Command(exe, args...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr

	// Notepad is a GUI app; Start and return without blocking, matching the
	// behavior of launching it directly from a shell.
	return cmd.Start()
}
