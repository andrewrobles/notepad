package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

func archiveNotepad(title string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("get home dir: %w", err)
	}

	source := filepath.Join(home, "notepad", ".md")
	file, err := os.OpenFile(source, os.O_RDWR, 0)
	if err != nil {
		return fmt.Errorf("open notepad: %w", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("read notepad: %w", err)
	}

	if len(data) == 0 {
		fmt.Println("No notepad contents to archive.")
		return nil
	}

	date := time.Now().Format("2006-01-02")
	relPath := filepath.Join(date, title+".md")

	path, err := archiveMarkdown("notepad", relPath, string(data))
	if err != nil {
		return err
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("seek notepad before clearing: %w", err)
	}

	current, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("read notepad before clearing: %w", err)
	}

	if !bytes.Equal(current, data) {
		return fmt.Errorf(
			"notepad changed during archiving; archive saved to %s; source not cleared",
			path,
		)
	}

	if err := file.Truncate(0); err != nil {
		return fmt.Errorf("clear notepad after archiving to %s: %w", path, err)
	}

	if err := file.Sync(); err != nil {
		return fmt.Errorf("sync cleared notepad: %w", err)
	}

	fmt.Printf("Archived notepad to %s\n", path)
	return nil
}
