package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func openNotepad() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("get home dir: %w", err)
	}

	path := filepath.Join(home, "notepad", ".md")

	notepad := "notepad.exe"
	if root := os.Getenv("SystemRoot"); root != "" {
		notepad = filepath.Join(root, "System32", "notepad.exe")
	}

	cmd := exec.Command(notepad, path)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("open notepad: %w", err)
	}

	return nil
}
